package sql_generator

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/pingcap/parser/compatibility_reporter/yacc_parser"
)

const maxLoopback = 2
const maxBuildTreeTime = 10 * 1000

type node interface {
	walk() bool
	materialize(writer io.StringWriter) error
	loopbackDetection(productionName string, sameParent uint) bool
}

type literalNode struct {
	value string
}

func (ln *literalNode) walk() bool {
	return true
}

func (ln *literalNode) materialize(writer io.StringWriter) error {
	if len(ln.value) != 0 {
		_, err := writer.WriteString(ln.value)
		if err != nil {
			return err
		}
		_, err = writer.WriteString(" ")
		if err != nil {
			return err
		}
	}
	return nil
}

func (ln *literalNode) loopbackDetection(productionName string, sameParent uint) (loop bool) {
	panic("unreachable")
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

func newProductionNode(production yacc_parser.Production, parent *expressionNode) *productionNode {
	return &productionNode{
		name:   production.Head,
		exprs:  make([]*expressionNode, len(production.Alter)),
		parent: parent,
	}
}

func newExpressionNode(seq yacc_parser.Seq, parent *productionNode) *expressionNode {
	return &expressionNode{
		items:  make([]node, len(seq.Items)),
		parent: parent,
	}
}

func literal(token string) (string, bool) {
	if strings.HasPrefix(token, "'") && strings.HasSuffix(token, "'") {
		return strings.Trim(token, "'"), true
	}
	return "", false
}

var startBuildTree int64

func buildTree(productionName string, parent *expressionNode) node {
	startTime := time.Now().UnixNano() / 1e6
	if startTime-startBuildTree > maxBuildTreeTime {
		println("build tree time over", maxBuildTreeTime, "ms")
		return &terminator{}
	}
	if parent != nil && parent.loopbackDetection(productionName, 0) {
		return &terminator{}
	}
	production, exist := productionMap[productionName]
	if !exist {
		panic(fmt.Sprintf("Production '%s' not found", productionName))
	}
	root := newProductionNode(production, parent)
	for i, seq := range production.Alter {
		root.exprs[i] = newExpressionNode(seq, root)
		for j, item := range seq.Items {
			if literalStr, isLiteral := literal(item); isLiteral {
				root.exprs[i].items[j] = &literalNode{value: literalStr}
			} else {
				root.exprs[i].items[j] = buildTree(item, root.exprs[i])
			}
		}
	}
	useTime := time.Now().UnixNano()/1e6 - startTime
	if useTime > 5 {
		println("build tree", productionName, "use", useTime, "ms")
	}
	return root
}

// SQLIterator is a iterator interface of sql generator
type SQLIterator interface {

	// HasNext returns whether the iterator exists next sql case
	HasNext() bool

	// Next returns next sql case in iterator
	// it will panic when the iterator doesn't exist next sql case
	Next() string
}

// SQLSequentialIterator is a iterator of sql generator
type SQLSequentialIterator struct {
	root             *productionNode
	alreadyPointNext bool
	noNext           bool
}

// HasNext returns whether the iterator exists next sql case
func (i *SQLSequentialIterator) HasNext() bool {
	if !i.alreadyPointNext {
		i.noNext = i.root.walk()
		i.alreadyPointNext = true
	}
	return !i.noNext
}

// Next returns next sql case in iterator
// it will panic when the iterator doesn't exist next sql case
func (i *SQLSequentialIterator) Next() string {
	if !i.HasNext() {
		panic("there isn't next item in this sql iterator")
	}
	i.alreadyPointNext = false
	stringBuffer := bytes.NewBuffer([]byte{})
	err := i.root.materialize(stringBuffer)
	if err != nil {
		panic("buffer write failure" + err.Error())
	}
	return stringBuffer.String()
}

var productionMap map[string]yacc_parser.Production

func checkProductionMap() {
	for _, production := range productionMap {
		for _, seqs := range production.Alter {
			for _, seq := range seqs.Items {
				if _, isLiteral := literal(seq); isLiteral {
					continue
				}
				if _, exist := productionMap[seq]; !exist {
					panic(fmt.Sprintf("Production '%s' not found", seq))
				}
			}
		}
	}
}

func initProductionMap(productions []yacc_parser.Production) {
	productionMap = make(map[string]yacc_parser.Production)
	for _, production := range productions {
		if pm, exist := productionMap[production.Head]; exist {
			pm.Alter = append(pm.Alter, production.Alter...)
			productionMap[production.Head] = pm
		} else {
			productionMap[production.Head] = production
		}
	}
	checkProductionMap()
}

// GenerateSQLSequentially returns a `SQLSequentialIterator` which can generate sql case by case sequential
// productions is a `Production` array created by `yacc_parser.Parse`
// productionName assigns a production name as the root node.
func GenerateSQLSequentially(productions []yacc_parser.Production, productionName string) SQLIterator {
	println("finish parse bnf file")
	initProductionMap(productions)
	println("finish create production map, map size:", len(productionMap))
	startBuildTree = time.Now().UnixNano() / 1e6
	pNode := buildTree(productionName, nil).(*productionNode)
	println("finish build tree")
	pNode.pruneTerminator()
	println("finish prune terminator branch")
	return &SQLSequentialIterator{
		root:             pNode,
		alreadyPointNext: true,
		noNext:           false,
	}
}

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

// GenerateSQLSequentially returns a `SQLSequentialIterator` which can generate sql case by case randomly
// productions is a `Production` array created by `yacc_parser.Parse`
// productionName assigns a production name as the root node.
func GenerateSQLRandomly(productions []yacc_parser.Production, productionName string) SQLIterator {
	initProductionMap(productions)
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
	if sameParentNum >= maxLoopback {
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
