package parser

import (
	"bytes"
	"crypto/md5"
	"reflect"
	"sync"
	"unicode"
	"unsafe"
)

// Digest generates a digest(or sql-id) for a SQL.
// the purpose of digest is to identity a group of similar SQLs, then we can do other logic base on it.
func Digest(sql string) string {
	d := digesterPool.Get().(*digester)
	sqlLen := len(sql)
	d.lexer.reset(sql)
	for {
		tok, pos, lit := d.lexer.scan()
		if tok == unicode.ReplacementChar && d.lexer.r.eof() {
			break
		}
		if pos.Offset == sqlLen {
			break
		}
		d.buffer.WriteRune(' ')
		switch tok {
		case intLit, stringLit, decLit, floatLit, bitLit, hexLit:
			d.buffer.WriteRune('?')
		default:
			d.buffer.WriteString(lit)
		}
	}
	result := md5.Sum(d.buffer.Bytes())
	d.buffer.Reset()
	digesterPool.Put(d)
	return byte2str(result[:])
}

var digesterPool = sync.Pool{
	New: func() interface{} {
		return &digester{
			lexer: NewScanner(""),
		}
	},
}

type digester struct {
	buffer bytes.Buffer
	lexer  *Scanner
}

func byte2str(b []byte) (s string) {
	if len(b) == 0 {
		return ""
	}
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pstring.Data = pbytes.Data
	pstring.Len = pbytes.Len
	return
}
