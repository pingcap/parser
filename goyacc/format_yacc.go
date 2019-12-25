package main

import (
	"bufio"
	"fmt"
	parser "github.com/cznic/parser/yacc"
	"github.com/cznic/strutil"
	"github.com/pingcap/errors"
	"github.com/pingcap/parser/format"
	gofmt "go/format"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
)

func Format(inputFilename string) (err error) {
	name, ext := splitFileNameAndExt(inputFilename)
	outFilename := fmt.Sprintf("%s_golden%s", name, ext)
	yFmt := &OutputFormatter{}
	if err = yFmt.Setup(outFilename); err != nil {
		return err
	}
	defer func() {
		err = yFmt.Teardown()
	}()

	spec, err := parseFileToSpec(inputFilename)
	if err != nil {
		return err
	}

	if err = printDefinitions(yFmt, spec.Defs); err != nil {
		return err
	}

	if err = printRules(yFmt, spec.Rules); err != nil {
		return err
	}
	panic("implement me!")
}

func splitFileNameAndExt(filename string) (name, ext string) {
	ext = path.Ext(filename)
	return filename[:len(filename)-len(ext)], ext
}

func parseFileToSpec(inputFilename string) (*parser.Specification, error) {
	src, err := ioutil.ReadFile(inputFilename)
	if err != nil {
		return nil, err
	}
	return parser.Parse(token.NewFileSet(), inputFilename, src)
}

// Definition represents data reduced by productions:
//
//	Definition:
//	        START IDENTIFIER
//	|       UNION                      // Case 1
//	|       LCURL RCURL                // Case 2
//	|       ReservedWord Tag NameList  // Case 3
//	|       ReservedWord Tag           // Case 4
//	|       ERROR_VERBOSE              // Case 5
const (
	StartIdentifierCase = iota
	UnionDefinitionCase
	LCURLRCURLCase
	ReservedWordTagNameListCase
	ReservedWordTagCase
)

func printDefinitions(formatter format.Formatter, definitions []*parser.Definition) error {
	for _, def := range definitions {
		var err error
		switch def.Case {
		case StartIdentifierCase:
			err = handleStart(formatter, def)
		case UnionDefinitionCase:
			err = handleUnion(formatter, def)
		case LCURLRCURLCase:
			err = handleProlog(formatter, def)
		case ReservedWordTagNameListCase:
			err = handleReservedWordTagNameList(formatter, def)
		}
		if err != nil {
			return err
		}
	}
	_, err := formatter.Format("\n%%%%\n")
	return err
}

func handleStart(f format.Formatter, definition *parser.Definition) error {
	if err := Ensure(definition).
		and(definition.Token2).
		and(definition.Token2).NotNil(); err != nil {
		return err
	}
	cmt1 := strings.Join(definition.Token.Comments, "\n")
	cmt2 := strings.Join(definition.Token2.Comments, "\n")
	_, err := f.Format("\n%s%s\t%s%s", cmt1, definition.Token.Val, cmt2, definition.Token2.Val)
	return err
}

func handleUnion(f format.Formatter, definition *parser.Definition) error {
	if err := Ensure(definition).
		and(definition.Value).NotNil(); err != nil {
		return err
	}
	if len(definition.Value) != 0 {
		_, err := f.Format("%%union%i%s%u\n", definition.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleProlog(f format.Formatter, definition *parser.Definition) error {
	if err := Ensure(definition).
		and(definition.Value).NotNil(); err != nil {
		return err
	}
	_, err := f.Format("%%{%s%%}\n\n", definition.Value)
	return err
}

func handleReservedWordTagNameList(f format.Formatter, def *parser.Definition) error {
	if err := Ensure(def).
		and(def.ReservedWord).
		and(def.ReservedWord.Token).NotNil(); err != nil {
		return err
	}
	comment := getTokenComment(def.ReservedWord.Token, divCommentLayout)
	directive := def.ReservedWord.Token.Val

	hasTag := def.Tag != nil
	var wordAfterDirective string
	if hasTag {
		wordAfterDirective = joinTag(def.Tag)
 	} else {
 		wordAfterDirective = joinNames(def.Nlist)
	}

	if _, err := f.Format("%s%s%s%i", comment, directive, wordAfterDirective); err != nil {
		return err
	}
	if hasTag {
		if _, err := f.Format("\n"); err != nil {
			return err
		}
		if err := printNameListVertical(f, def.Nlist); err != nil {
			return err
		}
	}
	_, err := f.Format("%u\n")
	return err
}

func joinTag(tag *parser.Tag) string {
	var sb strings.Builder
	sb.WriteString("\t")
	if tag.Token != nil {
		sb.WriteString(tag.Token.Val)
	}
	if tag.Token2 != nil {
		sb.WriteString(tag.Token2.Val)
	}
	if tag.Token3 != nil {
		sb.WriteString(tag.Token3.Val)
	}
	return sb.String()
}

type commentLayout int8
const (
	spanCommentLayout commentLayout = iota
	divCommentLayout
)

func getTokenComment(token *parser.Token, layout commentLayout) string {
	if len(token.Comments) == 0 {
		return ""
	}
	var splitter, beforeComment string
	switch layout {
	case spanCommentLayout:
		splitter, beforeComment = " ", ""
	case divCommentLayout:
		splitter, beforeComment = "\n", "\n"
	default:
		panic(errors.WithStack(errors.Errorf("unsupported commentLayout: %v", layout)))
	}

	var sb strings.Builder
	sb.WriteString(beforeComment)
	for _, comment := range token.Comments {
		sb.WriteString(comment)
		sb.WriteString(splitter)
	}
	return sb.String()
}

func printNameListVertical(f format.Formatter, names NameArr) (err error) {
	rest := names
	for len(rest) != 0 {
		var processing NameArr
		processing, rest = rest[:1], rest[1:]

		var noComments NameArr
		noComments, rest = rest.span(noComment)
		processing = append(processing, noComments...)

		maxCharLength := processing.findMaxLength()
		for _, name := range processing {
			if err := printSingleName(f, name, maxCharLength); err != nil {
				return err
			}
		}
	}
	return nil
}

func joinNames(names NameArr) string {
	var sb strings.Builder
	for _, name := range names {
		sb.WriteString(" ")
		sb.WriteString(getTokenComment(name.Token, spanCommentLayout))
		sb.WriteString(name.Token.Val)
	}
	return sb.String()
}

func printSingleName(f format.Formatter, name *parser.Name, maxCharLength int) error {
	if hasComments(name) {
		_, err := f.Format("\n%s\n", strings.Join(name.Token.Comments, "\n"))
		if err != nil {
			return err
		}
	}
	if name.LiteralStringOpt != nil && name.LiteralStringOpt.Token != nil {
		strLit := fmt.Sprintf(" %s", name.LiteralStringOpt.Token.Val)
		_, err := f.Format("%-*s%s\n", maxCharLength, name.Token.Val, strLit)
		return err
	} else {
		_, err := f.Format("%s\n", name.Token.Val)
		return err
	}
}

type NameArr []*parser.Name

func (ns NameArr) span(pred func(*parser.Name) bool) (NameArr, NameArr) {
	first := ns.takeWhile(pred)
	second := ns[len(first):]
	return first, second
}

func (ns NameArr) takeWhile(pred func(*parser.Name) bool) NameArr {
	for i, def := range ns {
		if pred(def) {
			continue
		}
		return ns[:i]
	}
	return ns
}

func (ns NameArr) findMaxLength() int {
	maxLen := -1
	for _, s := range ns {
		if len(s.Token.Val) > maxLen {
			maxLen = len(s.Token.Val)
		}
	}
	return maxLen
}

func hasComments(n *parser.Name) bool {
	return len(n.Token.Comments) != 0
}

func noComment(n *parser.Name) bool {
	return !hasComments(n)
}

type RuleArr []*parser.Rule

func printRules(f format.Formatter, rules RuleArr) (err error) {
	var lastRuleName string
	for _, rule := range rules {
		if rule.Name.Val == lastRuleName {
			_, err = f.Format("\n|\t%i")
		} else {
			cmt := getTokenComment(rule.Name, divCommentLayout)
			_, err = f.Format("\n\n%s%s:%i\n", cmt, rule.Name.Val)
		}
		if err != nil {
			return err
		}
		lastRuleName = rule.Name.Val

		if err = printRuleBody(f, rule); err != nil {
			return err
		}
		if _, err = f.Format("%u"); err != nil {
			return err
		}
	}
	_, err = f.Format("\n%%%%\n")
	return err
}

func printRuleBody(f format.Formatter, rule *parser.Rule) error {
	emptyBody := true
	for i, body := range rule.Body {
		switch b := body.(type) {
		case string, int:
			if bInt, ok := b.(int); ok {
				b = fmt.Sprintf("'%c'", bInt)
			}
			term := fmt.Sprintf(" %s", b)
			if i == 0 {
				term = term[1:]
			}

			if _, err := f.Format("%s", term); err != nil {
				return err
			}
			emptyBody = false
		case *parser.Action:
			if rule.Precedence != nil {
				if err := handlePrecedenceBeforeAction(f, rule.Precedence); err != nil {
					return err
				}
				emptyBody = false
			}

			if !emptyBody {
				if _, err := f.Format("\n"); err != nil {
					return err
				}
			}

			goSnippet, err := formatGoSnippet(b.Values)
			if err != nil {
				return err
			}

			if _, err := f.Format("{%i"); err != nil {
				return err
			}
			if _, err := f.Format(goSnippet); err != nil {
				return err
			}
			if _, err := f.Format("%u\n}"); err != nil {
				return err
			}
		}
	}
	return nil
}

func handlePrecedenceBeforeAction(f format.Formatter, p *parser.Precedence) error {
	if err := Ensure(p.Token).
		and(p.Token2).NotNil(); err != nil {
		return err
	}
	cmt := getTokenComment(p.Token, spanCommentLayout)
	_, err := f.Format("%s%s %s", cmt, p.Token.Val, p.Token2.Val)
	return err
}

func formatGoSnippet(actVal []*parser.ActionValue) (string, error) {
	tran := &SpecialActionValTransformer{
		store: map[string]string{},
	}
	goSnippet := collectGoSnippet(tran, actVal)
	formatted, err := gofmt.Source([]byte(goSnippet))
	if err != nil {
		return "", err
	}
	formattedSnippet := tran.restore(string(formatted))
	return strings.TrimRight(formattedSnippet, "\n"), nil
}

func collectGoSnippet(tran *SpecialActionValTransformer, actionValArr []*parser.ActionValue) string {
	var sb strings.Builder
	for _, value := range actionValArr {
		trimTab := removeLineBeginBlanks(value.Src)
		sb.WriteString(tran.transform(trimTab))
	}
	snipWithPar := strings.TrimSpace(sb.String())
	if strings.HasPrefix(snipWithPar, "{") && strings.HasSuffix(snipWithPar, "}") {
		return snipWithPar[1:len(snipWithPar)-1]
	}
	return ""
}

var lineBeginBlankRegex = regexp.MustCompile("(?m)^[\t ]+")

func removeLineBeginBlanks(src string) string {
	return lineBeginBlankRegex.ReplaceAllString(src, "")
}

type SpecialActionValTransformer struct {
	store map[string]string
}

const yaccFmtVar = "_yaccfmt_var_"
var yaccFmtVarRegex = regexp.MustCompile("_yaccfmt_var_[0-9]{1,5}")

func (s *SpecialActionValTransformer) transform(val string) string {
	if strings.HasPrefix(val, "$") {
		generated := fmt.Sprintf("%s%d", yaccFmtVar, len(s.store))
		s.store[generated] = val
		return generated
	}
	return val
}

func (s *SpecialActionValTransformer) restore(src string) string {
	return yaccFmtVarRegex.ReplaceAllStringFunc(src, func(matched string) string {
		origin, ok := s.store[matched]
		if !ok {
			panic(errors.WithStack(errors.Errorf("mismatch in SpecialActionValTransformer")).Error())
		}
		return origin
	})
}

type OutputFormatter struct {
	file *os.File
	readBytes []byte
	out *bufio.Writer
	formatter strutil.Formatter
}

func (y *OutputFormatter) Setup(filename string) (err error) {
	if y.file, err = os.Create(filename); err != nil {
		return
	}
	y.out = bufio.NewWriter(y.file)
	y.formatter = strutil.IndentFormatter(y.out, "\t")
	return
}

func (y *OutputFormatter) Teardown() error {
	if y.out != nil {
		if err := y.out.Flush(); err != nil {
			return err
		}
	}
	if y.file != nil {
		if err := y.file.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (y *OutputFormatter) Format(format string, args ...interface{}) (int, error) {
	return y.formatter.Format(format, args...)
}

func (y *OutputFormatter) Write(bytes []byte) (int, error) {
	return y.formatter.Write(bytes)
}

type NotNilAssert struct {
	idx int
	err error
}

func (n *NotNilAssert) and(target interface{}) *NotNilAssert {
	if n.err != nil {
		return n
	}
	if target == nil {
		n.err = errors.WithStack(errors.Errorf("encounter nil, index: %d", n.idx))
	}
	n.idx += 1
	return n
}

func (n *NotNilAssert) NotNil() error {
	return n.err
}

func Ensure(target interface{}) *NotNilAssert {
	return (&NotNilAssert{}).and(target)
}
