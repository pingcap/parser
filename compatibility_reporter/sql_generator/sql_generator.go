package sql_generator

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"strings"

	. "github.com/pingcap/parser/compatibility_reporter/yacc_parser"
)

const maxLoopback = 2

type node interface {
	walk() bool
	materialize(writer io.StringWriter) error
	loopbackDetection(productionName string, sameParent uint) bool
}

type terminator struct {
}

func (t *terminator) walk() bool {
	panic("unreachable, you maybe forget calling `pruneTerminator` before calling `walk`")
}

func (t *terminator) materialize(writer io.StringWriter) error {
	panic("unreachable, you maybe forget calling `pruneTerminator` before calling `walk`")
}

func (t *terminator) loopbackDetection(productionName string, sameParent uint) (loop bool) {
	panic("unreachable, you maybe forget calling `pruneTerminator` before calling `walk`")
}

type expressionNode struct {
	items  []node
	parent *productionNode
}

func (en *expressionNode) materialize(writer io.StringWriter) error {
	for _, item := range en.items {
		err := item.materialize(writer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (en *expressionNode) walk() (carry bool) {
	previousCarry := true
	for i := len(en.items) - 1; i >= 0 && previousCarry; i-- {
		previousCarry = en.items[i].walk()
	}
	return previousCarry
}

func (en *expressionNode) existTerminator() bool {
	for _, item := range en.items {
		if _, ok := item.(*terminator); ok {
			return true
		}
	}
	for _, item := range en.items {
		if pNode, ok := item.(*productionNode); ok {
			pNode.pruneTerminator()
		}
	}
	return false
}

func (en *expressionNode) loopbackDetection(productionName string, sameParent uint) (loop bool) {
	if sameParent >= maxLoopback {
		return true
	}
	return en.parent != nil && en.parent.loopbackDetection(productionName, sameParent)
}

type productionNode struct {
	name      string
	exprs     []*expressionNode
	parent    *expressionNode
	walkIndex int
	pruned    bool
}

func (pn *productionNode) walk() (carry bool) {
	if pn.exprs[pn.walkIndex].walk() {
		pn.walkIndex += 1
		if pn.walkIndex >= len(pn.exprs) {
			pn.walkIndex = 0
			carry = true
		}
	}
	return
}

func (pn *productionNode) materialize(writer io.StringWriter) error {
	return pn.exprs[pn.walkIndex].materialize(writer)
}

func (pn *productionNode) loopbackDetection(productionName string, sameParent uint) (loop bool) {
	if pn.name == productionName {
		sameParent++
	}
	if sameParent >= maxLoopback {
		return true
	}
	return pn.parent != nil && pn.parent.loopbackDetection(productionName, sameParent)
}

// pruneTerminator remove the branch whose include terminator node.
func (pn *productionNode) pruneTerminator() {
	if pn.pruned {
		return
	}
	var newExprs []*expressionNode
	for _, expr := range pn.exprs {
		if !expr.existTerminator() {
			newExprs = append(newExprs, expr)
		}
	}
	pn.exprs = newExprs
	pn.pruned = true
}

func literal(token string) (string, bool) {
	if strings.HasPrefix(token, "'") && strings.HasSuffix(token, "'") {
		return strings.Trim(token, "'"), true
	}
	return "", false
}

// SQLIterator is a iterator interface of sql generator
type SQLIterator interface {

	// HasNext returns whether the iterator exists next sql case
	HasNext() bool

	// Next returns next sql case in iterator
	// it will panic when the iterator doesn't exist next sql case
	Next() string
}

var productionMap map[string]Production

// SQLRandomlyIterator is a iterator of sql generator
type SQLRandomlyIterator struct {
	productionName string
}

// HasNext returns whether the iterator exists next sql case
func (i *SQLRandomlyIterator) HasNext() bool {
	return true
}

// Next returns next sql case in iterator
// it will panic when the iterator doesn't exist next sql case
func (i *SQLRandomlyIterator) Next() string {
	stringBuffer := bytes.NewBuffer([]byte{})
	generateSQLRandomly(i.productionName, nil, stringBuffer)
	output := stringBuffer.String()
	if strings.Contains(output, "####Terminator####") {
		return i.Next()
	}
	return output
}

// GenerateSQLRandomly returns a `SQLSequentialIterator` which can generate sql case by case randomly
// productions is a `Production` array created by `yacc_parser.Parse`
// productionName assigns a production name as the root node.
func GenerateSQLRandomly(productions []Production, productionName string) SQLIterator {
	productionMap = initProductionMap(productions)
	return &SQLRandomlyIterator{
		productionName: productionName,
	}
}

func generateSQLRandomly(productionName string, parents []string, writer io.StringWriter) {
	production, exist := productionMap[productionName]
	if !exist {
		panic(fmt.Sprintf("Production '%s' not found", productionName))
	}
	sameParentNum := 0
	for _, parent := range parents {
		if parent == productionName {
			sameParentNum++
		}
	}
	if sameParentNum >= maxDepthCount {
		_, err := writer.WriteString("####Terminator####")
		if err != nil {
			panic("fail to write `io.StringWriter`")
		}
		return
	}
	parents = append(parents, productionName)
	seqs := production.Alter[rand.Intn(len(production.Alter))]
	for _, seq := range seqs.Items {
		if literalStr, isLiteral := literal(seq); isLiteral {
			if literalStr != "" {
				_, err := writer.WriteString(literalStr)
				if err != nil {
					panic("fail to write `io.StringWriter`")
				}
				_, err = writer.WriteString(" ")
				if err != nil {
					panic("fail to write `io.StringWriter`")
				}
			}
		} else {
			generateSQLRandomly(seq, parents, writer)
		}
	}
}

const maxDepthCount = 2

var whiteList = map[string]*struct{}{
	"ident":     {},
	"%empty":    {},
	"IDENT":     {},
	"IDENT_sys": {},
}

type ProdNode struct {
	prod string
	next *ProdNode
}

type ProdList struct {
	head *ProdNode
	size int
}

func newProdList(productions []Production) *ProdList {
	prodList := &ProdList{nil, 0}
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
	max     []int
	arr     []int
	current int
}

func newWalkIndices() WalkIndices {
	return WalkIndices{nil, nil, -1}
}

func (wi *WalkIndices) tryCarry() (reachEnding bool) {
	var i int
	for i = wi.current; i > 0 && wi.arr[i] > wi.max[i]; i-- {
		wi.arr[i] = 0
		wi.arr[i-1]++
	}
	if i == 0 && wi.arr[0] > wi.max[0] {
		return true
	}
	wi.max = wi.max[:i+1]
	wi.arr = wi.arr[:i+1]
	return false
}
func (wi *WalkIndices) produceChoice(currentMaxChoice int) int {
	wi.current++
	if wi.current >= len(wi.arr) {
		wi.arr = append(wi.arr, 0)
		wi.max = append(wi.max, currentMaxChoice)
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
	wi      WalkIndices
	prodMap map[string]Production
	start   string
	cache   string
}

func NewSQLEnumIterator(prods []Production, prodName string) SQLIterator {
	prodMap := initProductionMap(prods)
	wi := newWalkIndices()
	cache, err := generateSQL(&wi, prodName, prodMap)
	if err != nil {
		panic(fmt.Sprintf("%v", err.Error()))
	}
	return &SQLEnumIterator{
		wi:      wi,
		prodMap: prodMap,
		start:   prodName,
		cache:   cache,
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
		} else {
			productionMap[production.Head] = production
		}
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
	var prodList = &ProdList{&ProdNode{start, nil}, 1}
	var sql []string
	walkIndices.reset()

	for prodList.size > 0 {
		head := prodList.popHead()
		depthCounts[head.prod]++

		seqs := filterMaxDepth(prodMap[head.prod].Alter, depthCounts, maxDepthCount)
		nextChoice := walkIndices.produceChoice(len(seqs) - 1)
		if nextChoice >= len(seqs) {
			if ending := walkIndices.tryCarry(); ending {
				return "", io.EOF
			}
			return generateSQL(walkIndices, start, prodMap)
		}

		subProds := seqs[nextChoice].Items
		for i := len(subProds) - 1; i >= 0; i-- {
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
	for k := range prodMap {
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
		if !ok || whiteList[s] != nil {
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
