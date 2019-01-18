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

// DigestText generates the normalized statements.
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
	tokens tokenDeque
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

const (
	genericSymbol     = -1
	genericSymbolList = -2
)

func (d *sqlDigester) normalize(sql string) {
	d.lexer.reset(sql)
	for {
		tok, pos, lit := d.lexer.scan()
		if tok == unicode.ReplacementChar && d.lexer.r.eof() {
			break
		}
		if pos.Offset == len(sql) {
			break
		}
		currTok := token{tok, strings.ToLower(lit)}
		if tok == hintEnd {
			d.popUntilHintBegin()
			continue
		}
		if isLit(tok) {
			currTok = d.genericLit(currTok)
		}
		d.tokens = append(d.tokens, currTok)
	}
	d.lexer.reset("")
	for i, token := range d.tokens {
		d.buffer.WriteString(token.lit)
		if i != len(d.tokens)-1 {
			d.buffer.WriteRune(' ')
		}
	}
	d.tokens = d.tokens[:0]
}

func (d *sqlDigester) genericLit(currTok token) token {
	// "?, ?, ?, ?" => "..."
	last2 := d.tokens.back(2)
	if isGenericList(last2) {
		d.tokens.popBack(2)
		currTok.tok = genericSymbolList
		currTok.lit = "..."
		return currTok
	}

	// order by n => order by n
	if currTok.tok == intLit {
		last2 := d.tokens.back(2)
		if isOrderBy(last2) {
			return currTok
		}
	}

	// 2 => ?
	currTok.tok = genericSymbol
	currTok.lit = "?"
	return currTok
}
func (d *sqlDigester) popUntilHintBegin() {
	for {
		token := d.tokens.popBack(1)
		if len(token) == 0 {
			return
		}
		if token[0].tok == hintBegin {
			return
		}
	}
}

type token struct {
	tok int
	lit string
}

type tokenDeque []token

func (s *tokenDeque) pushBack(t token) {
	*s = append(*s, t)
}

func (s *tokenDeque) popBack(n int) (t []token) {
	t = (*s)[len(*s)-n:]
	*s = (*s)[:len(*s)-n]
	return
}

func (s *tokenDeque) back(n int) (t []token) {
	if len(*s)-2 < 0 {
		return
	}
	t = (*s)[len(*s)-2:]
	return
}

func isLit(tok int) (beLit bool) {
	switch tok {
	case intLit, stringLit, decLit, floatLit, bitLit, hexLit:
		beLit = true
	default:
	}
	return
}

func isGenericList(last2 []token) (generic bool) {
	if len(last2) < 2 {
		return false
	}
	if !isComma(last2[1].tok) {
		return false
	}
	switch last2[0].tok {
	case genericSymbol, genericSymbolList:
		generic = true
	default:
	}
	return
}

func isOrderBy(last2 []token) (orderBy bool) {
	if len(last2) < 2 {
		return false
	}
	if last2[1].lit != "by" {
		return false
	}
	orderBy = last2[0].lit == "order"
	return
}

func isComma(tok int) (isComma bool) {
	isComma = tok == 44
	return
}
