// Copyright 2018 PingCAP, Inc.
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

package ast

import (
	"fmt"
	"io"
	"strings"
)

// IsReadOnly checks whether the input ast is readOnly.
func IsReadOnly(node Node) bool {
	switch st := node.(type) {
	case *SelectStmt:
		if st.LockTp == SelectLockForUpdate {
			return false
		}

		checker := readOnlyChecker{
			readOnly: true,
		}

		node.Accept(&checker)
		return checker.readOnly
	case *ExplainStmt, *DoStmt:
		return true
	default:
		return false
	}
}

// readOnlyChecker checks whether a query's ast is readonly, if it satisfied
// 1. selectstmt;
// 2. need not to set var;
// it is readonly statement.
type readOnlyChecker struct {
	readOnly bool
}

// Enter implements Visitor interface.
func (checker *readOnlyChecker) Enter(in Node) (out Node, skipChildren bool) {
	switch node := in.(type) {
	case *VariableExpr:
		// like func rewriteVariable(), this stands for SetVar.
		if !node.IsSystem && node.Value != nil {
			checker.readOnly = false
			return in, true
		}
	}
	return in, false
}

// Leave implements Visitor interface.
func (checker *readOnlyChecker) Leave(in Node) (out Node, ok bool) {
	return in, checker.readOnly
}

//RestoreFlag mark the Restore format
type RestoreFlags uint64

const (
	RestoreStringSingleQuotes RestoreFlags = 1 << iota
	RestoreStringDoubleQuotes
	RestoreStringEscapeBackslash

	RestoreKeyWordUppercase
	RestoreKeyWordLowercase

	RestoreNameUppercase
	RestoreNameLowercase
	RestoreNameOriginal
	RestoreNameSingleQuotes
	RestoreNameDoubleQuotes
	RestoreNameBackQuote
	RestoreNameEscapeBackQuote
)

const (
	DefaultRestoreFlags = RestoreStringSingleQuotes | RestoreKeyWordUppercase | RestoreNameOriginal |
		RestoreNameBackQuote | RestoreNameEscapeBackQuote
)

// Has return weather `rf` has this flag
func (rf RestoreFlags) Has(flag RestoreFlags) bool {
	return rf&flag != 0
}

// RestoreCtx is Restore context to hold flags and writer
type RestoreCtx struct {
	Flags RestoreFlags
	In    io.Writer
}

// NewRestoreCtx return a new RestoreCtx
func NewRestoreCtx(flags RestoreFlags, in io.Writer) *RestoreCtx {
	return &RestoreCtx{flags, in}
}

// WriteKeyWord write the keyword into writer
func (ctx *RestoreCtx) WriteKeyWord(keyWord string) {
	switch {
	case ctx.Flags.Has(RestoreKeyWordUppercase):
		keyWord = strings.ToUpper(keyWord)
	case ctx.Flags.Has(RestoreKeyWordLowercase):
		keyWord = strings.ToLower(keyWord)
	}
	_, _ = fmt.Fprint(ctx.In, keyWord)
}

// WriteKeyWord write the string into writer
func (ctx *RestoreCtx) WriteString(str string) {
	if ctx.Flags.Has(RestoreStringEscapeBackslash) {
		str = strings.Replace(str, "\\", "\\\\", -1)
	}
	quotes := ""
	switch {
	case ctx.Flags.Has(RestoreStringSingleQuotes):
		quotes = "'"
	case ctx.Flags.Has(RestoreStringDoubleQuotes):
		quotes = "\""
	}
	_, _ = fmt.Fprint(ctx.In, quotes, str, quotes)
}

// WriteName write the name into writer
func (ctx *RestoreCtx) WriteName(name string) {
	if ctx.Flags.Has(RestoreNameEscapeBackQuote) {
		name = strings.Replace(name, "`", "``", -1)
	}
	switch {
	case ctx.Flags.Has(RestoreNameUppercase):
		name = strings.ToUpper(name)
	case ctx.Flags.Has(RestoreNameLowercase):
		name = strings.ToLower(name)
	}
	quotes := ""
	switch {
	case ctx.Flags.Has(RestoreNameSingleQuotes):
		quotes = "'"
	case ctx.Flags.Has(RestoreNameDoubleQuotes):
		quotes = "\""
	case ctx.Flags.Has(RestoreNameBackQuote):
		quotes = "`"
	}
	_, _ = fmt.Fprint(ctx.In, quotes, name, quotes)
}

// WriteName write the plain text into writer without any handling
func (ctx *RestoreCtx) WritePlain(plainText string) {
	_, _ = fmt.Fprint(ctx.In, plainText)
}
