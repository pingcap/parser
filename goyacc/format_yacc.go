package main

import (
	"bufio"
	"fmt"
	"github.com/cznic/strutil"
	"github.com/cznic/y"
	"go/token"
	"os"
	"path"
)

func Format(filename string) (err error) {
	name, ext := splitFileNameAndExt(filename)
	formattedFilename := fmt.Sprintf("%s_golden%s", name, ext)

	yFmt := SetupYaccFormatter(formattedFilename)
	defer yFmt.Teardown()

	yaccParser, err := y.ProcessFile(token.NewFileSet(), filename, &y.Options{
		AllowConflicts: true,
	})
	if err != nil {
		return err
	}

	yFmt.Printf("%%{\n%s%%}\n", yaccParser.Prologue)
	unionStr := getUnionOrigin(yaccParser)
	if len(unionStr) != 0 {
		yFmt.Printf("%%union%i%s%u\n", unionStr)
	}

}

func splitFileNameAndExt(filename string) (name, ext string) {
	ext = path.Ext(filename)
	return filename[:len(filename)-len(ext)], ext
}

const UnionDefinitionCase = 1

func getUnionOrigin(yaccParser *y.Parser) string {
	for _, def := range yaccParser.Definitions {
		if def.Case != UnionDefinitionCase {
			continue
		}
		return def.Value
	}
	return ""
}

type YaccFormatter struct {
	file *os.File
	out *bufio.Writer
	formatter strutil.Formatter
	err error
}

func SetupYaccFormatter(filename string) *YaccFormatter {
	f := &YaccFormatter{
		file: nil,
		out:  nil,
		err:  nil,
	}
	if f.file, f.err = os.Create(filename); f.err != nil {
		return f
	}
	f.out = bufio.NewWriter(f.file)
	f.formatter = strutil.IndentFormatter(f.out, "\t")
	return f
}

func (y *YaccFormatter) Teardown() {
	if y.err != nil {
		return
	}
	y.err = y.out.Flush()
	if y.err != nil {
		return
	}
	y.err = y.file.Close()
}

func (y *YaccFormatter) Printf(format string, args ...interface{}) {
	if y.err != nil {
		return
	}
	_, y.err = y.formatter.Format(format, args...)
}
