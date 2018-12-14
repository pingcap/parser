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

// Mutually exclusive group of `RestoreFlags`:
// [RestoreStringSingleQuotes, RestoreStringDoubleQuotes]
// [RestoreKeyWordUppercase, RestoreKeyWordLowercase]
// [RestoreNameUppercase, RestoreNameLowercase]
// [RestoreNameDoubleQuotes, RestoreNameBackQuotes]
// The flag with the left position in each group has a higher priority.
const (
	RestoreStringSingleQuotes RestoreFlags = 1 << iota
	RestoreStringDoubleQuotes
	RestoreStringEscapeBackslash

	RestoreKeyWordUppercase
	RestoreKeyWordLowercase

	RestoreNameUppercase
	RestoreNameLowercase
	RestoreNameDoubleQuotes
	RestoreNameBackQuotes
)

const (
	DefaultRestoreFlags = RestoreStringSingleQuotes | RestoreKeyWordUppercase | RestoreNameBackQuotes
)

// Has returns weather `rf` has this `flag`.
func (rf RestoreFlags) Has(flag RestoreFlags) bool {
	return rf&flag != 0
}

// HasStringSingleQuotesFlag returns weather `rf` has `RestoreStringSingleQuotes` flag.
func (rf RestoreFlags) HasStringSingleQuotesFlag() bool {
	return rf.Has(RestoreStringSingleQuotes)
}

// HasStringDoubleQuotesFlag returns weather `rf` has `RestoreStringDoubleQuotes` flag.
func (rf RestoreFlags) HasStringDoubleQuotesFlag() bool {
	return rf.Has(RestoreStringDoubleQuotes)
}

// HasStringEscapeBackslashFlag returns weather `rf` has `RestoreStringEscapeBackslash` flag.
func (rf RestoreFlags) HasStringEscapeBackslashFlag() bool {
	return rf.Has(RestoreStringEscapeBackslash)
}

// HasKeyWordUppercaseFlag returns weather `rf` has `RestoreKeyWordUppercase` flag.
func (rf RestoreFlags) HasKeyWordUppercaseFlag() bool {
	return rf.Has(RestoreKeyWordUppercase)
}

// HasKeyWordLowercaseFlag returns weather `rf` has `RestoreKeyWordLowercase` flag.
func (rf RestoreFlags) HasKeyWordLowercaseFlag() bool {
	return rf.Has(RestoreKeyWordLowercase)
}

// HasNameUppercaseFlag returns weather `rf` has `RestoreNameUppercase` flag.
func (rf RestoreFlags) HasNameUppercaseFlag() bool {
	return rf.Has(RestoreNameUppercase)
}

// HasNameLowercaseFlag returns weather `rf` has `RestoreNameLowercase` flag.
func (rf RestoreFlags) HasNameLowercaseFlag() bool {
	return rf.Has(RestoreNameLowercase)
}

// HasNameDoubleQuotesFlag returns weather `rf` has `RestoreNameDoubleQuotes` flag.
func (rf RestoreFlags) HasNameDoubleQuotesFlag() bool {
	return rf.Has(RestoreNameDoubleQuotes)
}

// HasNameBackQuotesFlag returns weather `rf` has `RestoreNameBackQuotes` flag.
func (rf RestoreFlags) HasNameBackQuotesFlag() bool {
	return rf.Has(RestoreNameBackQuotes)
}

// RestoreCtx is `Restore` context to hold flags and writer.
type RestoreCtx struct {
	Flags RestoreFlags
	In    io.Writer
}

// NewRestoreCtx returns a new `RestoreCtx`.
func NewRestoreCtx(flags RestoreFlags, in io.Writer) *RestoreCtx {
	return &RestoreCtx{flags, in}
}

// WriteKeyWord writes the `keyWord` into writer.
// `keyWord` will be converted format(uppercase and lowercase for now) according to `RestoreFlags`.
func (ctx *RestoreCtx) WriteKeyWord(keyWord string) {
	switch {
	case ctx.Flags.HasKeyWordUppercaseFlag():
		keyWord = strings.ToUpper(keyWord)
	case ctx.Flags.HasKeyWordLowercaseFlag():
		keyWord = strings.ToLower(keyWord)
	}
	fmt.Fprint(ctx.In, keyWord)
}

// WriteString writes the string into writer
// `str` may be wrapped in quotes and escaped according to RestoreFlags.
func (ctx *RestoreCtx) WriteString(str string) {
	if ctx.Flags.HasStringEscapeBackslashFlag() {
		str = strings.Replace(str, `\`, `\\`, -1)
	}
	quotes := ""
	switch {
	case ctx.Flags.HasStringSingleQuotesFlag():
		str = strings.Replace(str, `'`, `''`, -1)
		quotes = `'`
	case ctx.Flags.HasStringDoubleQuotesFlag():
		str = strings.Replace(str, `"`, `""`, -1)
		quotes = `"`
	}
	fmt.Fprint(ctx.In, quotes, str, quotes)
}

// WriteName writes the name into writer
// `name` maybe wrapped in quotes and escaped according to RestoreFlags.
func (ctx *RestoreCtx) WriteName(name string) {
	switch {
	case ctx.Flags.HasNameUppercaseFlag():
		name = strings.ToUpper(name)
	case ctx.Flags.HasNameLowercaseFlag():
		name = strings.ToLower(name)
	}
	quotes := ""
	switch {
	case ctx.Flags.HasNameDoubleQuotesFlag():
		name = strings.Replace(name, `"`, `""`, -1)
		quotes = `"`
	case ctx.Flags.HasNameBackQuotesFlag():
		name = strings.Replace(name, "`", "``", -1)
		quotes = "`"
	}
	fmt.Fprint(ctx.In, quotes, name, quotes)
}

// WritePlain writes the plain text into writer without any handling.
func (ctx *RestoreCtx) WritePlain(plainText string) {
	fmt.Fprint(ctx.In, plainText)
}

// WritePlainf write the plain text into writer without any handling.
func (ctx *RestoreCtx) WritePlainf(format string, a ...interface{}) {
	fmt.Fprintf(ctx.In, format, a...)
}
