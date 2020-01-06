// Copyright 2020 PingCAP, Inc.
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
	"strconv"
	"strings"
	"unicode"

	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/terror"
)

var (
	ErrWarnOptimizerHintUnsupportedHint = terror.ClassParser.New(mysql.ErrWarnOptimizerHintUnsupportedHint, mysql.MySQLErrName[mysql.ErrWarnOptimizerHintUnsupportedHint])
	ErrWarnOptimizerHintInvalidToken    = terror.ClassParser.New(mysql.ErrWarnOptimizerHintInvalidToken, mysql.MySQLErrName[mysql.ErrWarnOptimizerHintInvalidToken])
	ErrWarnMemoryQuotaOverflow          = terror.ClassParser.New(mysql.ErrWarnMemoryQuotaOverflow, mysql.MySQLErrName[mysql.ErrWarnMemoryQuotaOverflow])
	ErrWarnOptimizerHintParseError      = terror.ClassParser.New(mysql.ErrWarnOptimizerHintParseError, mysql.MySQLErrName[mysql.ErrWarnOptimizerHintParseError])
	ErrWarnOptimizerHintInvalidInteger  = terror.ClassParser.New(mysql.ErrWarnOptimizerHintInvalidInteger, mysql.MySQLErrName[mysql.ErrWarnOptimizerHintInvalidInteger])
)

// hintScanner implements the yyhintLexer interface
type hintScanner struct {
	Scanner
}

func (hs *hintScanner) Errorf(format string, args ...interface{}) error {
	inner := hs.Scanner.Errorf(format, args...)
	return ErrWarnOptimizerHintParseError.GenWithStackByArgs(inner)
}

func (hs *hintScanner) Lex(lval *yyhintSymType) int {
	tok, pos, lit := hs.scan()
	var errorTokenType string

	switch tok {
	case intLit:
		n, e := strconv.ParseUint(lit, 10, 64)
		if e != nil {
			hs.AppendError(ErrWarnOptimizerHintInvalidInteger.GenWithStackByArgs(lit))
			return int(unicode.ReplacementChar)
		}
		lval.number = n
		return hintIntLit

	case singleAtIdentifier:
		lval.ident = lit
		return hintSingleAtIdentifier

	case identifier:
		lval.ident = lit
		if tok1, ok := hintTokenMap[strings.ToUpper(lit)]; ok {
			return tok1
		}
		return hintIdentifier

	case stringLit:
		lval.ident = lit
		if hs.sqlMode.HasANSIQuotesMode() && hs.r.s[pos.Offset] == '"' {
			return hintIdentifier
		}
		return hintStringLit

	case bitLit:
		if strings.HasPrefix(lit, "0b") {
			lval.ident = lit
			return hintIdentifier
		}
		errorTokenType = "bit-value literal"

	case hexLit:
		if strings.HasPrefix(lit, "0x") {
			lval.ident = lit
			return hintIdentifier
		}
		errorTokenType = "hexadecimal literal"

	case quotedIdentifier:
		lval.ident = lit
		return hintIdentifier

	case eq:
		return '='

	case floatLit:
		errorTokenType = "floating point number"
	case decLit:
		errorTokenType = "decimal number"

	default:
		if tok <= 0x7f {
			return tok
		}
		errorTokenType = "unknown token"
	}

	hs.AppendError(ErrWarnOptimizerHintInvalidToken.GenWithStackByArgs(errorTokenType, lit, tok))
	return int(unicode.ReplacementChar)
}

type hintParser struct {
	lexer  hintScanner
	result []*ast.TableOptimizerHint

	// the following fields are used by yyParse to reduce allocation.
	cache  []yyhintSymType
	yylval yyhintSymType
	yyVAL  *yyhintSymType
}

// ParseHint parses an optimizer hint (the interior of `/*+ ... */`).
func ParseHint(input string, sqlMode mysql.SQLMode) ([]*ast.TableOptimizerHint, []error) {
	// yyhintDebug = 2

	hp := hintParser{cache: make([]yyhintSymType, 50)}
	hp.lexer.reset(input)
	hp.lexer.SetSQLMode(sqlMode)

	yyhintParse(&hp.lexer, &hp)

	warns, errs := hp.lexer.Errors()
	if len(errs) == 0 {
		errs = warns
	}
	return hp.result, errs
}

func (hp *hintParser) warnUnsupportedHint(name string) {
	warn := ErrWarnOptimizerHintUnsupportedHint.GenWithStackByArgs(name)
	hp.lexer.warns = append(hp.lexer.warns, warn)
}

func (hp *hintParser) lastErrorAsWarn() {
	hp.lexer.lastErrorAsWarn()
}
