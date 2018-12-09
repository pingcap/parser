// Copyright 2017 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package ast_test

import (
	. "github.com/pingcap/check"
	"github.com/pingcap/parser"
	. "github.com/pingcap/parser/ast"
	"strings"
)

var _ = Suite(&testDMLSuite{})

type testDMLSuite struct {
}

func (ts *testDMLSuite) TestDMLVisitorCover(c *C) {
	ce := &checkExpr{}

	tableRefsClause := &TableRefsClause{TableRefs: &Join{Left: &TableSource{Source: &TableName{}}, On: &OnCondition{Expr: ce}}}

	stmts := []struct {
		node             Node
		expectedEnterCnt int
		expectedLeaveCnt int
	}{
		{&DeleteStmt{TableRefs: tableRefsClause, Tables: &DeleteTableList{}, Where: ce,
			Order: &OrderByClause{}, Limit: &Limit{Count: ce, Offset: ce}}, 4, 4},
		{&ShowStmt{Table: &TableName{}, Column: &ColumnName{}, Pattern: &PatternLikeExpr{Expr: ce, Pattern: ce}, Where: ce}, 3, 3},
		{&LoadDataStmt{Table: &TableName{}, Columns: []*ColumnName{{}}, FieldsInfo: &FieldsClause{}, LinesInfo: &LinesClause{}}, 0, 0},
		{&Assignment{Column: &ColumnName{}, Expr: ce}, 1, 1},
		{&ByItem{Expr: ce}, 1, 1},
		{&GroupByClause{Items: []*ByItem{{Expr: ce}, {Expr: ce}}}, 2, 2},
		{&HavingClause{Expr: ce}, 1, 1},
		{&Join{Left: &TableSource{Source: &TableName{}}}, 0, 0},
		{&Limit{Count: ce, Offset: ce}, 2, 2},
		{&OnCondition{Expr: ce}, 1, 1},
		{&OrderByClause{Items: []*ByItem{{Expr: ce}, {Expr: ce}}}, 2, 2},
		{&SelectField{Expr: ce, WildCard: &WildCardField{}}, 1, 1},
		{&TableName{}, 0, 0},
		{tableRefsClause, 1, 1},
		{&TableSource{Source: &TableName{}}, 0, 0},
		{&WildCardField{}, 0, 0},

		// TODO: cover childrens
		{&InsertStmt{Table: tableRefsClause}, 1, 1},
		{&UnionStmt{}, 0, 0},
		{&UpdateStmt{TableRefs: tableRefsClause}, 1, 1},
		{&SelectStmt{}, 0, 0},
		{&FieldList{}, 0, 0},
		{&UnionSelectList{}, 0, 0},
		{&WindowSpec{}, 0, 0},
		{&PartitionByClause{}, 0, 0},
		{&FrameClause{}, 0, 0},
		{&FrameBound{}, 0, 0},
	}

	for _, v := range stmts {
		ce.reset()
		v.node.Accept(checkVisitor{})
		c.Check(ce.enterCnt, Equals, v.expectedEnterCnt)
		c.Check(ce.leaveCnt, Equals, v.expectedLeaveCnt)
		v.node.Accept(visitor1{})
	}
}

type tableNameTestCase struct {
	sourceSQL string
	expectSQL string
}

/* ***********************************************************************************************************
 * TableName Stmt Test Case
 */
// only TableName test data
func (tc *testDMLSuite) createTestCase4TableName() []tableNameTestCase {

	return []tableNameTestCase{
		{"CREATE TABLE dbb.`tbb1` (id VARCHAR(128) NOT NULL);", "CREATE TABLE `dbb`.`tbb1` (id VARCHAR(128) NOT NULL);"},
		{"CREATE TABLE `tbb2` (id VARCHAR(128) NOT NULL);", "CREATE TABLE `tbb2` (id VARCHAR(128) NOT NULL);"},
		{"CREATE TABLE tbb3 (id VARCHAR(128) NOT NULL);", "CREATE TABLE `tbb3` (id VARCHAR(128) NOT NULL);"},
		{"CREATE TABLE dbb.`hello-world` (id VARCHAR(128) NOT NULL);", "CREATE TABLE `dbb`.`hello-world` (id VARCHAR(128) NOT NULL);"},
		{"CREATE TABLE `dbb`.`hello-world` (id VARCHAR(128) NOT NULL);", "CREATE TABLE `dbb`.`hello-world` (id VARCHAR(128) NOT NULL);"},
		{"CREATE TABLE `dbb.HelloWorld` (id VARCHAR(128) NOT NULL);", "CREATE TABLE `dbb.HelloWorld` (id VARCHAR(128) NOT NULL);"},
	}

}
func (tc *testDMLSuite) TestTableNameRestore(c *C) {

	parser := parser.New()
	var testNodes []tableNameTestCase
	testNodes = append(testNodes, tc.createTestCase4TableName()...)

	for _, node := range testNodes {

		// String comparison
		stmt, err := parser.ParseOneStmt(node.sourceSQL, "", "")
		comment := Commentf("source %#v", node)
		c.Assert(err, IsNil, comment)

		var sb strings.Builder
		sb.WriteString("CREATE TABLE" + " ")
		err = stmt.(*CreateTableStmt).Table.Restore(&sb)
		sb.WriteString("(id VARCHAR(128) NOT NULL);")
		c.Assert(err, IsNil, comment)
		restoreSql := sb.String()
		comment = Commentf("\n source %#v; \n restore %v", node, restoreSql)

		c.Assert(restoreSql, Equals, node.expectSQL, comment)

		// Ast comparison
		stmt2, err := parser.ParseOneStmt(restoreSql, "", "")
		c.Assert(err, IsNil, comment)
		CleanNodeText(stmt)
		CleanNodeText(stmt2)
		c.Assert(stmt2, DeepEquals, stmt, comment)

	}

}

// add index hints test data
func (tc *testDMLSuite) createTestCase4TableNameIndexHints() []tableNameTestCase {

	return []tableNameTestCase{
		{"select * from t use index (hello)", "SELECT * FROM `t` USE INDEX (`hello`)"},
		{"select * from t use index (hello, world)", "SELECT * FROM `t` USE INDEX (`hello`, `world`)"},
		{"select * from t use index ()", "SELECT * FROM `t` USE INDEX ()"},
		{"select * from t use key ()", "SELECT * FROM `t` USE INDEX ()"},
		{"select * from t ignore key ()", "SELECT * FROM `t` IGNORE INDEX ()"},
		{"select * from t force key ()", "SELECT * FROM `t` FORCE INDEX ()"},
		{"select * from t use index for order by (idx1)", "SELECT * FROM `t` USE INDEX FOR ORDER BY (`idx1`)"},

		{"select * from t use index (hello, world, yes) force key (good)", "SELECT * FROM `t` USE INDEX (`hello`, `world`, `yes`) FORCE INDEX (`good`)"},
		{"select * from t use index (hello, world, yes) use index for order by (good)", "SELECT * FROM `t` USE INDEX (`hello`, `world`, `yes`) USE INDEX FOR ORDER BY (`good`)"},
		{"select * from t ignore key (hello, world, yes) force key (good)", "SELECT * FROM `t` IGNORE INDEX (`hello`, `world`, `yes`) FORCE INDEX (`good`)"},

		{"select * from t use index for group by (idx1) use index for order by (idx2)","SELECT * FROM `t` USE INDEX FOR GROUP BY (`idx1`) USE INDEX FOR ORDER BY (`idx2`)"},
		{"select * from t use index for group by (idx1) ignore key for order by (idx2)","SELECT * FROM `t` USE INDEX FOR GROUP BY (`idx1`) IGNORE INDEX FOR ORDER BY (`idx2`)"},
		{"select * from t use index for group by (idx1) ignore key for group by (idx2)","SELECT * FROM `t` USE INDEX FOR GROUP BY (`idx1`) IGNORE INDEX FOR GROUP BY (`idx2`)"},
		{"select * from t use index for order by (idx1) ignore key for group by (idx2)","SELECT * FROM `t` USE INDEX FOR ORDER BY (`idx1`) IGNORE INDEX FOR GROUP BY (`idx2`)"},

		{"select * from t use index for order by (idx1) ignore key for group by (idx2) use index (idx3)","SELECT * FROM `t` USE INDEX FOR ORDER BY (`idx1`) IGNORE INDEX FOR GROUP BY (`idx2`) USE INDEX (`idx3`)"},
		{"select * from t use index for order by (idx1) ignore key for group by (idx2) use index (idx3)","SELECT * FROM `t` USE INDEX FOR ORDER BY (`idx1`) IGNORE INDEX FOR GROUP BY (`idx2`) USE INDEX (`idx3`)"},

		{"select * from t use index (`foo``bar`) force index (`baz``1`, `xyz`)","SELECT * FROM `t` USE INDEX (`foo``bar`) FORCE INDEX (`baz``1`, `xyz`)"},
		{"select * from t force index (`foo``bar`) ignore index (`baz``1`, xyz)","SELECT * FROM `t` FORCE INDEX (`foo``bar`) IGNORE INDEX (`baz``1`, `xyz`)"},
		{"select * from t ignore index (`foo``bar`) force key (`baz``1`, xyz)","SELECT * FROM `t` IGNORE INDEX (`foo``bar`) FORCE INDEX (`baz``1`, `xyz`)"},
		{"select * from t ignore index (`foo``bar`) ignore key for group by (`baz``1`, xyz)","SELECT * FROM `t` IGNORE INDEX (`foo``bar`) IGNORE INDEX FOR GROUP BY (`baz``1`, `xyz`)"},
		{"select * from t ignore index (`foo``bar`) ignore key for order by (`baz``1`, xyz)","SELECT * FROM `t` IGNORE INDEX (`foo``bar`) IGNORE INDEX FOR ORDER BY (`baz``1`, `xyz`)"},

		{"select * from t use index for group by (`foo``bar`) use index for order by (`baz``1`, `xyz`)","SELECT * FROM `t` USE INDEX FOR GROUP BY (`foo``bar`) USE INDEX FOR ORDER BY (`baz``1`, `xyz`)"},
		{"select * from t use index for group by (`foo``bar`) ignore key for order by (`baz``1`, `xyz`)","SELECT * FROM `t` USE INDEX FOR GROUP BY (`foo``bar`) IGNORE INDEX FOR ORDER BY (`baz``1`, `xyz`)"},
		{"select * from t use index for group by (`foo``bar`) ignore key for group by (`baz``1`, `xyz`)","SELECT * FROM `t` USE INDEX FOR GROUP BY (`foo``bar`) IGNORE INDEX FOR GROUP BY (`baz``1`, `xyz`)"},
		{"select * from t use index for order by (`foo``bar`) ignore key for group by (`baz``1`, `xyz`)","SELECT * FROM `t` USE INDEX FOR ORDER BY (`foo``bar`) IGNORE INDEX FOR GROUP BY (`baz``1`, `xyz`)"},
	}

}
func (tc *testDMLSuite) TestTableNameIndexHintsRestore(c *C) {

	parser := parser.New()
	var testNodes []tableNameTestCase
	testNodes = append(testNodes, tc.createTestCase4TableNameIndexHints()...)

	for _, node := range testNodes {

		// String comparison
		stmt, err := parser.ParseOneStmt(node.sourceSQL, "", "")
		comment := Commentf("source %#v", node)
		c.Assert(err, IsNil, comment)

		var sb strings.Builder
		sb.WriteString("SELECT * FROM" + " ")
		err = stmt.(*SelectStmt).From.TableRefs.Left.(*TableSource).Source.(*TableName).Restore(&sb)
		c.Assert(err, IsNil, comment)
		restoreSql := sb.String()
		comment = Commentf("\n source %#v; \n restore %v", node, restoreSql)

		c.Assert(restoreSql, Equals, node.expectSQL, comment)

		// Ast comparison
		stmt2, err := parser.ParseOneStmt(restoreSql, "", "")
		c.Assert(err, IsNil, comment)
		CleanNodeText(stmt)
		CleanNodeText(stmt2)
		c.Assert(stmt2, DeepEquals, stmt, comment)

	}

}
