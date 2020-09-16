package main

import (
	"fmt"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	_ "github.com/pingcap/parser/test_driver"
)

func parse(sql string) (*ast.StmtNode, error) {
	p := parser.New()

	stmtNodes, _, err := p.Parse(sql, "", "")
	if err != nil {
		return nil, err
	}

	return &stmtNodes[0], nil
}

func main() {
	//astNode, err := parse("SELECT * FROM t1 UNION TABLE t2")
	//astNode, err := parse("TABLE t1 UNION TABLE t2")
	//astNode, err := parse("TABLE t1 UNION SELECT * FROM t2")
	//astNode, err := parse("SELECT * FROM t1 UNION SELECT * FROM t2")
	astNode, err := parse("TABLE t1 INTO OUTFILE 'a.txt'")
	if err != nil {
		fmt.Printf("parse error: %v\n", err.Error())
		return
	}
	fmt.Printf("%v\n", *astNode)
}
