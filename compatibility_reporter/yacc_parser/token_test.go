package yacc_parser

import (
	"bufio"
	"fmt"
	"io"
	"testing"
)

// Implement io.Reader.
type mockReader struct {
	str string
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	if len(m.str) == 0 {
		return 0, io.EOF
	}
	read := copy(p, m.str)
	m.str = m.str[read:]
	return read, nil
}

func newMockReader(str string) *bufio.Reader {
	return bufio.NewReader(&mockReader{str})
}

func assertEq(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("expect: '%v', but got: '%v'", expected, actual)
	}
}

func withTokenizeResult(origin string, visitor func(index int, tkn string)) {
	next := Tokenize(newMockReader(origin))
	for i := 0; ; i++ {
		tkn := next()
		if isEOF(tkn) {
			break
		}
		visitor(i, tkn.toString())
	}
}

func printTokenizeResult(origin string) {
	withTokenizeResult(origin, func(_ int, s string) {
		fmt.Println(s)
	})
}

func assertExpectedTokenResult(t *testing.T, origin string, expected []string) {
	withTokenizeResult(origin, func(idx int, s string) {
		assertEq(t, expected[idx], s)
	})
}

func TestTokenize(t *testing.T) {
	origin := `column_attribute_list: column_attribute_list column_attribute | column_attribute`
	expect := []string{"column_attribute_list", ":", "column_attribute_list", "column_attribute", "|", "column_attribute"}
	assertExpectedTokenResult(t, origin, expect)
}

func TestColonStrToken(t *testing.T) {
	origin := `this: is a test with 'colon appears inside a string :)'`
	expect := []string{"this", ":", "is", "a", "test", "with", "'colon appears inside a string :)'"}
	assertExpectedTokenResult(t, origin, expect)
}

func TestSimpleStr(t *testing.T) {
	origin := `a: 'b' c`
	expect := []string{"a", ":", "'b'", "c"}
	assertExpectedTokenResult(t, origin, expect)

	origin = `a: '"b' "'c"`
	expect = []string{"a", ":", `'"b'`, `"'c"`}
	assertExpectedTokenResult(t, origin, expect)

	origin = `a: 'b"' "c'"`
	expect = []string{"a", ":", `'b"'`, `"c'"`}
	assertExpectedTokenResult(t, origin, expect)
}

func TestA(t *testing.T) {
	origin := `t1: 'a' 'b' t2
    | 'c' t3
    | t2 'f' t3 'g'

t2: 'd'
    | t3 'e'

t3: 'f'
    | 'g' 'h'
	| 'i'`
	expect := []string{`t1`, `:`, `'a'`, `'b'`, `t2`, `|`, `'c'`, `t3`, `|`, `t2`, `'f'`, `t3`, `'g'`, `t2`, `:`, `'d'`, `|`, `t3`, `'e'`, `t3`, `:`, `'f'`, `|`, `'g'`, `'h'`, `|`, `'i'`}
	assertExpectedTokenResult(t, origin, expect)
}
