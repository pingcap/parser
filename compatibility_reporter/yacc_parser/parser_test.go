package yacc_parser

import (
	"fmt"
	"testing"
)

func TestParser(t *testing.T) {
	next := Tokenize(newMockReader(
		`sql_statement: simple_statement_or_begin empty1 ';' opt_end_of_input
		| simple_statement_or_begin END_OF_INPUT
		opt_end_of_input: empty
                | END_OF_INPUT`))
	p := Parse(next)
	if fmt.Sprintf("%v", p) != `[{sql_statement [{[simple_statement_or_begin empty1 ';' opt_end_of_input]} {[simple_statement_or_begin END_OF_INPUT]}]} {opt_end_of_input [{[empty]} {[END_OF_INPUT]}]}]` {
		t.Error("TestParser failed")
	}
}
