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
	. "github.com/pingcap/parser/ast"
)

var _ = Suite(&testFunctionsSuite{})

type testFunctionsSuite struct {
}

func (ts *testFunctionsSuite) TestFunctionsVisitorCover(c *C) {
	valueExpr := NewValueExpr(42)
	stmts := []Node{
		&AggregateFuncExpr{Args: []ExprNode{valueExpr}},
		&FuncCallExpr{Args: []ExprNode{valueExpr}},
		&FuncCastExpr{Expr: valueExpr},
		&WindowFuncExpr{Spec: WindowSpec{}},
	}

	for _, stmt := range stmts {
		stmt.Accept(visitor{})
		stmt.Accept(visitor1{})
	}
}

func (ts *testFunctionsSuite) TestFuncCallExprRestore(c *C) {
	testCases := []NodeRestoreTestCase{
		{"JSON_ARRAYAGG(attribute)", "JSON_ARRAYAGG(`attribute`)"},
		{"JSON_OBJECTAGG(attribute, value)", "JSON_OBJECTAGG(`attribute`, `value`)"},
		{"ABS(-1024)", "ABS(-1024)"},
		{"ACOS(3.14)", "ACOS(3.14)"},
		{"CONV('a',16,2)", "CONV('a', 16, 2)"},
		{"COS(PI())", "COS(PI())"},
		{"RAND()", "RAND()"},
	}
	extractNodeFunc := func(node Node) Node {
		return node.(*SelectStmt).Fields.Fields[0].Expr
	}
	RunNodeRestoreTest(c, testCases, "select %s", extractNodeFunc)
}

func (ts *testFunctionsSuite) TestFuncCastExprRestore(c *C) {
	testCases := []NodeRestoreTestCase{
		{"CONVERT('Müller' USING utf8mb4)", "CONVERT('Müller' USING utf8mb4)"},
		{"CONVERT('Müller', CHAR(32) CHARACTER SET utf8)", "CONVERT('Müller', CHAR(32) CHARACTER SET utf8)"},
		{"CAST('test' AS CHAR CHARACTER SET utf8)", "CAST('test' AS CHAR CHARACTER SET utf8)"},
		{"BINARY 'New York'", "BINARY 'New York'"},
	}
	extractNodeFunc := func(node Node) Node {
		return node.(*SelectStmt).Fields.Fields[0].Expr
	}
	RunNodeRestoreTest(c, testCases, "select %s", extractNodeFunc)
}

func (ts *testFunctionsSuite) TestAggregateFuncExprRestore(c *C) {
	testCases := []NodeRestoreTestCase{
		{"AVG(test_score)", "AVG(`test_score`)"},
		{"AVG(distinct test_score)", "AVG(DISTINCT `test_score`)"},
		{"BIT_AND(test_score)", "BIT_AND(`test_score`)"},
		{"BIT_OR(test_score)", "BIT_OR(`test_score`)"},
		{"BIT_XOR(test_score)", "BIT_XOR(`test_score`)"},
		{"COUNT(test_score)", "COUNT(`test_score`)"},
		{"COUNT(*)", "COUNT(1)"},
		{"COUNT(DISTINCT scores, results)", "COUNT(DISTINCT `scores`, `results`)"},
		{"MIN(test_score)", "MIN(`test_score`)"},
		{"MIN(DISTINCT test_score)", "MIN(DISTINCT `test_score`)"},
		{"MAX(test_score)", "MAX(`test_score`)"},
		{"MAX(DISTINCT test_score)", "MAX(DISTINCT `test_score`)"},
		{"STD(test_score)", "STD(`test_score`)"},
		{"STDDEV(test_score)", "STDDEV(`test_score`)"},
		{"STDDEV_POP(test_score)", "STDDEV_POP(`test_score`)"},
		{"STDDEV_SAMP(test_score)", "STDDEV_SAMP(`test_score`)"},
		{"SUM(test_score)", "SUM(`test_score`)"},
		{"SUM(DISTINCT test_score)", "SUM(DISTINCT `test_score`)"},
		{"VAR_POP(test_score)", "VAR_POP(`test_score`)"},
		{"VAR_SAMP(test_score)", "VAR_SAMP(`test_score`)"},
		{"VARIANCE(test_score)", "VAR_POP(`test_score`)"},
	}
	extractNodeFunc := func(node Node) Node {
		return node.(*SelectStmt).Fields.Fields[0].Expr
	}
	RunNodeRestoreTest(c, testCases, "select %s", extractNodeFunc)
}
