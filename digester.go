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
	"crypto/sha256"
	"fmt"
	hash2 "hash"
	"strings"
	"sync"
	"unicode"
)

// DigestHash generates the digest of statements.
func DigestHash(sql string) (result string) {
	d := digesterPool.Get().(*sqlDigester)
	result = d.doDigest(sql)
	digesterPool.Put(d)
	return
}

// DigestText generates the normalized sql query.
func DigestText(sql string) (result string) {
	d := digesterPool.Get().(*sqlDigester)
	result = d.doDigestText(sql)
	digesterPool.Put(d)
	return
}

var digesterPool = sync.Pool{
	New: func() interface{} {
		return &sqlDigester{
			lexer:  NewScanner(""),
			hasher: sha256.New(),
		}
	},
}

// sqlDigester is used to compute DigestHash or DigestText for sql.
type sqlDigester struct {
	buffer bytes.Buffer
	lexer  *Scanner
	hasher hash2.Hash
}

func (d *sqlDigester) doDigest(sql string) (result string) {
	d.normalize(sql)
	d.hasher.Write(d.buffer.Bytes())
	d.buffer.Reset()
	result = fmt.Sprintf("%x", d.hasher.Sum(nil))
	d.hasher.Reset()
	return
}

func (d *sqlDigester) doDigestText(sql string) (result string) {
	d.normalize(sql)
	result = string(d.buffer.Bytes())
	d.lexer.reset("")
	d.buffer.Reset()
	return
}

func (d *sqlDigester) normalize(sql string) {
	d.lexer.reset(sql)
	isHead := true
	for {
		tok, pos, lit := d.lexer.scan()
		if tok == unicode.ReplacementChar && d.lexer.r.eof() {
			break
		}
		if pos.Offset == len(sql) {
			break
		}
		if isHead {
			isHead = false
		} else {
			d.buffer.WriteRune(' ')
		}
		switch tok {
		case intLit, stringLit, decLit, floatLit, bitLit, hexLit:
			d.buffer.WriteRune('?')
		default:
			d.buffer.WriteString(strings.ToLower(lit))
		}
	}
	d.lexer.reset("")
}
