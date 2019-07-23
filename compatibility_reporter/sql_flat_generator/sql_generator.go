package sql_flat_generator

import (
	"fmt"
	"github.com/pingcap/parser/compatibility_reporter/sql_generator"
	. "github.com/pingcap/parser/compatibility_reporter/yacc_parser"
	"io"
	"strings"
)

const maxDepthCount = 2

type ProdNode struct {
	prod string
	next *ProdNode
}

type ProdList struct {
	head *ProdNode
	size int
}

func newProdList(productions []Production) *ProdList {
	prodList := &ProdList{nil, 0 }
	for i := len(productions) - 1; i >= 0; i-- {
		prodList.prepend(productions[i].Head)
	}
	return prodList
}

func (pl *ProdList) prepend(production string) {
	oldHead := pl.head
	pl.head = &ProdNode{production, oldHead}
	pl.size++
}

func (pl *ProdList) popHead() *ProdNode {
	if pl.size == 0 {
		return nil
	}
	head := pl.head
	pl.head = head.next
	head.next = nil
	pl.size--
	return head
}

func (pl *ProdList) dropWhile(pred func(*ProdNode) bool) {
	for pl.head != nil && pred(pl.head) {
		pl.popHead()
	}
}

func concat(first, second *ProdList) *ProdList {
	if first == nil || first.size == 0 {
		return second
	}
	lastNode := first.head
	for lastNode.next != nil {
		lastNode = lastNode.next
	}
	lastNode.next = second.head
	return first
}

type WalkIndices struct {
	arr     []int
	current int
}
func newWalkIndices() WalkIndices {
	return WalkIndices{nil, -1}
}

func (wi *WalkIndices) tryCarry() (reachEnding bool){
	if wi.current <= 0 {
		return true
	}
	wi.arr[wi.current] = 0
	wi.arr[wi.current - 1]++
	return false
}
func (wi *WalkIndices) produceChoice() int {
	wi.current++
	if wi.current >= len(wi.arr) {
		wi.arr = append(wi.arr, 0)
	}
	return wi.arr[wi.current]
}
func (wi *WalkIndices) reset() {
	wi.current = -1
}
func (wi *WalkIndices) nextPermutation() {
	wi.arr[wi.current]++
}

// SQLSequentialIterator is a iterator of sql generator
type SQLEnumIterator struct {
	wi WalkIndices
	prodMap map[string]Production
	start string
	cache string
}

func NewSQLEnumIterator(prods []Production, prodName string) sql_generator.SQLIterator {
	prodMap := initProductionMap(prods)
	wi := newWalkIndices()
	cache, err := generateSQL(&wi, prodName, prodMap)
	if err != nil {
		panic(fmt.Sprintf("%v", err.Error()))
	}
	return &SQLEnumIterator{
		wi: wi,
		prodMap: prodMap,
		start: prodName,
		cache: cache,
	}
}

func (se *SQLEnumIterator) HasNext() bool {
	return se.cache != ""
}

func (se *SQLEnumIterator) Next() string {
	ret := se.cache
	next, err := generateSQL(&se.wi, se.start, se.prodMap)
	if err == io.EOF {
		se.cache = ""
	} else if err != nil {
		panic(fmt.Sprintf("%v", err.Error()))
	} else {
		se.cache = next
	}
	return ret
}

func initProductionMap(productions []Production) map[string]Production {
	productionMap := make(map[string]Production)
	for _, production := range productions {
		if pm, exist := productionMap[production.Head]; exist {
			pm.Alter = append(pm.Alter, production.Alter...)
			productionMap[production.Head] = pm
		}
		productionMap[production.Head] = production
	}
	checkProductionMap(productionMap)
	return productionMap
}

func checkProductionMap(productionMap map[string]Production) {
	for _, production := range productionMap {
		for _, seqs := range production.Alter {
			for _, seq := range seqs.Items {
				if isLiteral(seq) {
					continue
				}
				if _, exist := productionMap[seq]; !exist {
					panic(fmt.Sprintf("Production '%s' not found", seq))
				}
			}
		}
	}
}

func generateSQL(walkIndices *WalkIndices, start string, prodMap map[string]Production) (string, error) {
	depthCounts := newDepthCountMap(prodMap)
	var prodList = &ProdList{&ProdNode{start, nil},1}
	var sql []string
	walkIndices.reset()

	for prodList.size > 0 {
		head := prodList.popHead()
		depthCounts[head.prod]++

		seqs := filterMaxDepth(prodMap[head.prod].Alter, depthCounts, maxDepthCount)
		nextChoice := walkIndices.produceChoice()
		if nextChoice >= len(seqs) {
			if ending := walkIndices.tryCarry(); ending {
				return "", io.EOF
			}
			return generateSQL(walkIndices, start, prodMap)
		}

		subProds := seqs[nextChoice].Items
		for i := len(subProds)-1; i >= 0; i-- {
			prodList.prepend(subProds[i])
		}

		for prodList.size > 0 && isLiteral(prodList.head.prod) {
			s := prodList.popHead().prod
			if len(s) == 2 { // ignore emptiness
				continue
			}
			sql = append(sql, s[1:len(s)-1])
		}
	}

	walkIndices.nextPermutation()
	return strings.Join(sql, " "), nil
}

func newDepthCountMap(prodMap map[string]Production) map[string]int {
	depthCounts := make(map[string]int, len(prodMap))
	for k, _ := range prodMap {
		depthCounts[k] = 0
	}
	return depthCounts
}

func filterMaxDepth(seqs []Seq, depthCount map[string]int, depthCountLimit int) []Seq {
	var result []Seq
	for _, seq := range seqs {
		if maxDepth(seq, depthCount) < depthCountLimit {
			result = append(result, seq)
		}
	}
	return result
}

func maxDepth(seq Seq, depthCount map[string]int) int {
	max := 0
	for _, s := range seq.Items {
		d, ok := depthCount[s]
		if !ok {
			continue
		}
		if max < d {
			max = d
		}
	}
	return max
}

func isLiteral(s string) bool {
	return strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")
}