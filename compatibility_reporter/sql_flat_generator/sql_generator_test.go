package sql_flat_generator

import (
	"bufio"
	"bytes"
	. "github.com/pingcap/check"
	"github.com/pingcap/parser/compatibility_reporter/yacc_parser"
	"testing"
)

func TestT(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&testSQLGeneratorSuite{})

type testSQLGeneratorSuite struct{}

func (s *testSQLGeneratorSuite) TestGenerator(c *C) {
	source := `t1: 'a' 'b' t2
    | 'c' t3
    | t2 'f' t3 'g'

	t2: 'd'
    | t3 'e'

	t3: 'f'
    | 'g' 'h'
    | 'i'`
	expect := []string{
		"a b d",
		"a b f e",
		"a b g h e",
		"a b i e",
		"c f",
		"c g h",
		"c i",
		"d f f g",
		"d f g h g",
		"d f i g",
		"f e f f g",
		"f e f g h g",
		"f e f i g",
		"g h e f f g",
		"g h e f g h g",
		"g h e f i g",
		"i e f f g",
		"i e f g h g",
		"i e f i g"}
	s.runTest(c, source, "t1", expect)
}

func (s *testSQLGeneratorSuite) TestLoopBNF(c *C) {
	source := `t1: 'a' 'b' t2
    | 'c' t3
    | t2 'f' 'g'

	t2: 'd'
    | t3 'e'

	t3: 'f'
    | 'g' 'h'
    | 'i'
	| t1`
	expect := []string{
		"a b d", "a b f e",
		"a b g h e", "a b i e",
		"a b a b d e", "a b a b f e e",
		"a b a b g h e e", "a b a b i e e",
		"a b c f e", "a b c g h e",
		"a b c i e", "a b d f g e",
		"a b f e f g e", "a b g h e f g e",
		"a b i e f g e", "c f",
		"c g h", "c i",
		"c a b d", "c a b f e",
		"c a b g h e", "c a b i e",
		"c c f", "c c g h",
		"c c i", "c d f g",
		"c f e f g", "c g h e f g",
		"c i e f g", "d f g",
		"f e f g", "g h e f g",
		"i e f g", "a b d e f g",
		"a b f e e f g", "a b g h e e f g",
		"a b i e e f g", "c f e f g",
		"c g h e f g", "c i e f g",
		"d f g e f g", "f e f g e f g",
		"g h e f g e f g", "i e f g e f g"}

	s.runTest(c, source, "t1", expect)
}


func (s *testSQLGeneratorSuite) runTest(c *C, source string, productionName string, expect []string) {
	p := yacc_parser.Parse(yacc_parser.Tokenize(bufio.NewReader(bytes.NewBuffer([]byte(source)))))
	sqlSeqIter := NewSQLEnumIterator(p, productionName)
	var output []string
	for sqlSeqIter.HasNext() {
		output = append(output, sqlSeqIter.Next())
	}
	c.Assert(output, DeepEquals, expect)
}