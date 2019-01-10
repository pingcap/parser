package main

import (
	"fmt"
	"github.com/pingcap/parser"
	. "github.com/pingcap/parser/ast"
	. "github.com/pingcap/parser/format"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"strings"
)

type printVisitor struct {
}

func (printVisitor) Enter(n Node) (node Node, skipChildren bool) {
	fmt.Printf("%#v\n", n)
	return n, false
}

func (printVisitor) Leave(n Node) (node Node, ok bool) {
	return n, true
}

// For test only.
func CleanNodeText(node Node) {
	var cleaner nodeTextCleaner
	node.Accept(&cleaner)
}

// nodeTextCleaner clean the text of a node and it's child node.
// For test only.
type nodeTextCleaner struct {
}

// Enter implements Visitor interface.
func (checker *nodeTextCleaner) Enter(in Node) (out Node, skipChildren bool) {
	in.SetText("")
	return in, false
}

// Leave implements Visitor interface.
func (checker *nodeTextCleaner) Leave(in Node) (out Node, ok bool) {
	return in, true
}
func main() {
	parser := parser.New()
	stmt, _, err := parser.Parse("CREATE TABLE foo (name CHAR(50) BINARY);", "", "")
	if err != nil {
		fmt.Println(err.Error())
	}

	var sb strings.Builder
	//err = stmt.Restore(NewRestoreCtx(DefaultRestoreFlags, &sb))
	//fmt.Println(sb.String())
	for _, node := range stmt {
		switch stmt := node.(type) {
		case *CreateTableStmt:
			err = stmt.Restore(NewRestoreCtx(DefaultRestoreFlags, &sb))
		}
	}
	fmt.Println(sb.String())

	fmt.Println("==========")

	//stmt1, err := parser.ParseOneStmt("CREATE TABLE foo (id varchar(50) collate utf8);", "", "")
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//CleanNodeText(stmt)
	//CleanNodeText(stmt1)
	//
	//stmt.Accept(printVisitor{})
	//fmt.Println()
	//fmt.Println()
	//fmt.Println()
	//fmt.Println()
	//stmt1.Accept(printVisitor{})
	//
	//result := reflect.DeepEqual(stmt, stmt1)
	//fmt.Println(result)
}
