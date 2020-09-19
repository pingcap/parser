package main

import (
	"fmt"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	_ "github.com/pingcap/parser/test_driver"
)

func parse(sql string) (*ast.StmtNode, error) {
	p := parser.New()

	stmts, _, err := p.Parse(sql, "", "")
	if err != nil {
		return nil, err
	}
	return &stmts[0], nil
}

func main() {
	//astNode, err := parse("Select * from t1 UNION TABLE t2")
	astNode, err := parse("VALUES ROW(1,2), ROW(4,5) UNION SELECT * FROM t")
	if err != nil {
		fmt.Printf("parser error: %v\n", err.Error())
		return
	}
	fmt.Printf("%v\n", *astNode)
}
