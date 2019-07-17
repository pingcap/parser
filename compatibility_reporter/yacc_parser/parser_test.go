package yacc_parser

import (
	"bufio"
	"bytes"
	"fmt"
	. "github.com/pingcap/check"
)


var _ = Suite(&testBNFParserSuite{})

type testBNFParserSuite struct {}

func (s *testBNFParserSuite) TestParser(c *C) {
	next := Tokenize(bufio.NewReader(bytes.NewBufferString(
		`sql_statement: simple_statement_or_begin empty1 ';' opt_end_of_input
		| simple_statement_or_begin END_OF_INPUT
		opt_end_of_input: empty
                | END_OF_INPUT`)))
	p := Parse(next)
	c.Assert(fmt.Sprintf("%v", p), Equals,`[{sql_statement [{[simple_statement_or_begin empty1 ';' opt_end_of_input]} {[simple_statement_or_begin END_OF_INPUT]}]} {opt_end_of_input [{[empty]} {[END_OF_INPUT]}]}]`)
}
