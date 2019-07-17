package yacc_parser

import (
	"bufio"
	"bytes"
	. "github.com/pingcap/check"
	"testing"
)

func TestT(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&testBNFTokenizerSuite{})

type testBNFTokenizerSuite struct{}

func (s *testBNFTokenizerSuite) TestTokenize(c *C) {
	origin := `column_attribute_list: column_attribute_list column_attribute | column_attribute`
	expect := []string{"column_attribute_list", ":", "column_attribute_list", "column_attribute", "|", "column_attribute"}
	assertExpectedTokenResult(c, origin, expect)
}

func (s *testBNFTokenizerSuite) TestColonStrToken(c *C) {
	origin := `this: is a test with 'colon appears inside a string :)'`
	expect := []string{"this", ":", "is", "a", "test", "with", "'colon appears inside a string :)'"}
	assertExpectedTokenResult(c, origin, expect)
}

func (s *testBNFTokenizerSuite) TestSimpleStr(c *C) {
	origin := `a: 'b' c`
	expect := []string{"a", ":", "'b'", "c"}
	assertExpectedTokenResult(c, origin, expect)

	origin = `a: '"b' "'c"`
	expect = []string{"a", ":", `'"b'`, `"'c"`}
	assertExpectedTokenResult(c, origin, expect)

	origin = `a: 'b"' "c'"`
	expect = []string{"a", ":", `'b"'`, `"c'"`}
	assertExpectedTokenResult(c, origin, expect)
}

func (s *testBNFTokenizerSuite) TestA(c *C) {
	origin := `t1: 'a' 'b' t2
    | 'c' t3
    | t2 'f' t3 'g'

t2: 'd'
    | t3 'e'

t3: 'f'
    | 'g' 'h'
	| 'i'`
	expect := []string{`t1`, `:`, `'a'`, `'b'`, `t2`, `|`, `'c'`, `t3`, `|`, `t2`, `'f'`, `t3`, `'g'`, `t2`, `:`, `'d'`, `|`, `t3`, `'e'`, `t3`, `:`, `'f'`, `|`, `'g'`, `'h'`, `|`, `'i'`}
	assertExpectedTokenResult(c, origin, expect)
}


func assertExpectedTokenResult(c *C, origin string, expected []string) {
	withTokenizeResult(origin, func(idx int, s string) {
		c.Assert(expected[idx], Equals, s)
	})
}

func printTokenizeResult(c *C, origin string, expected []string) {
	withTokenizeResult(origin, func(_ int, s string) {
		c.Logf("%s\n", s)
	})
}

func withTokenizeResult(origin string, visitor func(index int, tkn string)) {
	next := Tokenize(bufio.NewReader(bytes.NewBufferString(origin)))
	for i := 0; ; i++ {
		tkn := next()
		if isEOF(tkn) {
			break
		}
		visitor(i, tkn.toString())
	}
}
