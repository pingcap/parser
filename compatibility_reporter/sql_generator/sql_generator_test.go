package sql_generator

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/pingcap/parser/compatibility_reporter/yacc_parser"
)

func TestA(t *testing.T) {
	b := []byte(`t1: 'a' 'b' t2
    | 'c' t3
    | t2 'f' t3 'g'

t2: 'd'
    | t3 'e'

t3: 'f'
    | 'g' 'h'
    | 'i'`)

	p := yacc_parser.Parse(yacc_parser.Tokenize(bufio.NewReader(bytes.NewBuffer(b))))
	sqlIter := GenerateSQL(p, "t1")
	for sqlIter.HasNext() {
		println(sqlIter.Next())
	}
}

func TestB(t *testing.T) {

}
