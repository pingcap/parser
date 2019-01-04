package ast_test

import (
	"fmt"
	. "github.com/pingcap/check"
	"github.com/pingcap/parser"
	. "github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/util/restore"
	"strings"
)

// CleanNodeText set the text of node and all child node empty.
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

type NodeRestoreTestCase struct {
	sourceSQL string
	expectSQL string
}

func RunNodeRestoreTest(c *C, nodeTestCases []NodeRestoreTestCase, template string, extractNodeFunc func(node Node) Node) {
	parser := parser.New()
	parser.EnableWindowFunc(true)
	for _, testCase := range nodeTestCases {
		sourceSQL := fmt.Sprintf(template, testCase.sourceSQL)
		expectSQL := fmt.Sprintf(template, testCase.expectSQL)
		stmt, err := parser.ParseOneStmt(sourceSQL, "", "")
		comment := Commentf("source %#v", testCase)
		c.Assert(err, IsNil, comment)
		var sb strings.Builder
		err = extractNodeFunc(stmt).Restore(restore.NewRestoreCtx(restore.DefaultRestoreFlags, &sb))
		c.Assert(err, IsNil, comment)
		restoreSql := fmt.Sprintf(template, sb.String())
		comment = Commentf("source %#v; restore %v", testCase, restoreSql)
		c.Assert(restoreSql, Equals, expectSQL, comment)
		stmt2, err := parser.ParseOneStmt(restoreSql, "", "")
		c.Assert(err, IsNil, comment)
		CleanNodeText(stmt)
		CleanNodeText(stmt2)
		c.Assert(stmt2, DeepEquals, stmt, comment)
	}
}
