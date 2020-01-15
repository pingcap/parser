// Copyright 2015 PingCAP, Inc.
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

package terror_test

import (
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	. "github.com/pingcap/check"
	"github.com/pingcap/errors"
	"github.com/pingcap/parser/terror"
)

// Global error instances.
var (
	ErrCritical           = terror.New(terror.ClassGlobal, terror.CodeExecResultIsEmpty, "critical error %v")
	ErrResultUndetermined = terror.New(terror.ClassGlobal, terror.CodeResultUndetermined, "execution result undetermined")
)

func TestT(t *testing.T) {
	CustomVerboseFlag = true
	TestingT(t)
}

var _ = Suite(&testTErrorSuite{})

type testTErrorSuite struct {
}

func (s *testTErrorSuite) TestErrCode(c *C) {
	c.Assert(terror.CodeMissConnectionID, Equals, terror.ErrCode(1))
	c.Assert(terror.CodeResultUndetermined, Equals, terror.ErrCode(2))
}

func (s *testTErrorSuite) TestTError(c *C) {
	c.Assert(terror.ClassParser.String(), Not(Equals), "")
	c.Assert(terror.ClassOptimizer.String(), Not(Equals), "")
	c.Assert(terror.ClassKV.String(), Not(Equals), "")
	c.Assert(terror.ClassServer.String(), Not(Equals), "")

	parserErr := terror.ClassParser.NewBaseError(terror.ErrCode(1), "error 1")
	c.Assert(parserErr.Error(), Not(Equals), "")
	c.Assert(terror.ClassParser.EqualClass(parserErr), IsTrue)
	c.Assert(terror.ClassParser.NotEqualClass(parserErr), IsFalse)

	c.Assert(terror.ClassOptimizer.EqualClass(parserErr), IsFalse)
	optimizerErr := terror.ClassOptimizer.NewBaseError(terror.ErrCode(2), "abc")
	c.Assert(terror.ClassOptimizer.EqualClass(errors.New("abc")), IsFalse)
	c.Assert(terror.ClassOptimizer.EqualClass(nil), IsFalse)
	perr := terror.ParseError{BaseError: optimizerErr}
	c.Assert(optimizerErr.Equal(perr.GenWithStack("def")), IsTrue)
	c.Assert(optimizerErr.Equal(nil), IsFalse)
	c.Assert(optimizerErr.Equal(errors.New("abc")), IsFalse)

	// Test case for FastGen.
	c.Assert(optimizerErr.Equal(perr.FastGen("def")), IsTrue)
	c.Assert(optimizerErr.Equal(perr.FastGen("def: %s", "def")), IsTrue)
	kvErr := terror.ClassKV.NewBaseError(1062, "key already exist")
	perr = terror.ParseError{BaseError: kvErr}
	e := perr.FastGen("Duplicate entry '%d' for key 'PRIMARY'", 1)
	c.Assert(e.Error(), Equals, "[kv:1062]Duplicate entry '1' for key 'PRIMARY'")

	err := errors.Trace(ErrCritical.GenWithStackByArgs("test"))
	c.Assert(ErrCritical.Equal(err), IsTrue)

	err = errors.Trace(ErrCritical)
	c.Assert(ErrCritical.Equal(err), IsTrue)
}

func (s *testTErrorSuite) TestJson(c *C) {
	prevTErr := terror.ClassTable.NewBaseError(terror.CodeExecResultIsEmpty, "json test")
	buf, err := json.Marshal(prevTErr)
	c.Assert(err, IsNil)
	var curTErr terror.BaseError
	err = json.Unmarshal(buf, &curTErr)
	c.Assert(err, IsNil)
	isEqual := prevTErr.Equal(&curTErr)
	c.Assert(isEqual, IsTrue)
}

var predefinedErr = terror.New(terror.ClassExecutor, terror.ErrCode(123), "predefiend error")

func example() error {
	err := call()
	return errors.Trace(err)
}

func call() error {
	return predefinedErr.GenWithStack("error message:%s", "abc")
}

func (s *testTErrorSuite) TestTraceAndLocation(c *C) {
	err := example()
	stack := errors.ErrorStack(err)
	lines := strings.Split(stack, "\n")
	var sysStack = 0
	for _, line := range lines {
		if strings.Contains(line, runtime.GOROOT()) {
			sysStack++
		}
	}
	c.Assert(len(lines)-(2*sysStack), Equals, 15)
	var containTerr bool
	for _, v := range lines {
		if strings.Contains(v, "error_test.go") {
			containTerr = true
			break
		}
	}
	c.Assert(containTerr, IsTrue)
}

func (s *testTErrorSuite) TestErrorEqual(c *C) {
	e1 := errors.New("test error")
	c.Assert(e1, NotNil)

	e2 := errors.Trace(e1)
	c.Assert(e2, NotNil)

	e3 := errors.Trace(e2)
	c.Assert(e3, NotNil)

	c.Assert(errors.Cause(e2), Equals, e1)
	c.Assert(errors.Cause(e3), Equals, e1)
	c.Assert(errors.Cause(e2), Equals, errors.Cause(e3))

	e4 := errors.New("test error")
	c.Assert(errors.Cause(e4), Not(Equals), e1)

	e5 := errors.Errorf("test error")
	c.Assert(errors.Cause(e5), Not(Equals), e1)

	c.Assert(terror.ErrorEqual(e1, e2), IsTrue)
	c.Assert(terror.ErrorEqual(e1, e3), IsTrue)
	c.Assert(terror.ErrorEqual(e1, e4), IsTrue)
	c.Assert(terror.ErrorEqual(e1, e5), IsTrue)

	var e6 error

	c.Assert(terror.ErrorEqual(nil, nil), IsTrue)
	c.Assert(terror.ErrorNotEqual(e1, e6), IsTrue)
	code1 := terror.ErrCode(1)
	code2 := terror.ErrCode(2)
	te1 := terror.New(terror.ClassParser, code1, "abc")
	te2 := terror.New(terror.ClassParser, code1, "def")
	te3 := terror.New(terror.ClassKV, code1, "abc")
	te4 := terror.New(terror.ClassKV, code2, "abc")
	c.Assert(terror.ErrorEqual(te1, te2), IsTrue)
	c.Assert(terror.ErrorEqual(te1, te3), IsFalse)
	c.Assert(terror.ErrorEqual(te3, te4), IsFalse)
}
