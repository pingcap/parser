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

func (ts *testDMLSuite) TestWindowFuncExprRestore(c *C) {
	testCases := []NodeRestoreTestCase{
		{"RANK() OVER w", "RANK() OVER (`w`)"},
		{"RANK() OVER (PARTITION BY a)", "RANK() OVER (PARTITION BY `a`)"},
		{"MAX(DISTINCT a) OVER (PARTITION BY a)", "MAX(DISTINCT `a`) OVER (PARTITION BY `a`)"},
		{"MAX(DISTINCTROW a) OVER (PARTITION BY a)", "MAX(DISTINCT `a`) OVER (PARTITION BY `a`)"},
		{"MAX(DISTINCT ALL a) OVER (PARTITION BY a)", "MAX(DISTINCT `a`) OVER (PARTITION BY `a`)"},
		{"MAX(ALL a) OVER (PARTITION BY a)", "MAX(`a`) OVER (PARTITION BY `a`)"},
		{"FIRST_VALUE(val) IGNORE NULLS OVER w", "FIRST_VALUE(`val`) IGNORE NULLS OVER (`w`)"},
		{"FIRST_VALUE(val) RESPECT NULLS OVER w", "FIRST_VALUE(`val`) OVER (`w`)"},
		{"NTH_VALUE(val, 233) FROM LAST IGNORE NULLS OVER w", "NTH_VALUE(`val`, 233) FROM LAST IGNORE NULLS OVER (`w`)"},
		{"NTH_VALUE(val, 233) FROM FIRST IGNORE NULLS OVER w", "NTH_VALUE(`val`, 233) IGNORE NULLS OVER (`w`)"},
	}
	extractNodeFunc := func(node Node) Node {
		return node.(*SelectStmt).Fields.Fields[0].Expr
	}
	RunNodeRestoreTest(c, testCases, "select %s from t", extractNodeFunc)
}
