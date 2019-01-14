// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

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
