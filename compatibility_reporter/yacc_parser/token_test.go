package yacc_parser

import (
	"bufio"
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
func TestTokenize(t *testing.T) {
	next := Tokenize(newMockReader(
		`column_attribute_list: column_attribute_list 
		column_attribute | column_attribute`))
	expect := []string{"column_attribute_list", ":", "column_attribute_list", "column_attribute", "|", "column_attribute"}
	for i := 0; ; i++ {
		tkn := next()
		if isEOF(tkn) {
			break
		}
		assertEq(t, expect[i], tkn.toString())
	}
}
