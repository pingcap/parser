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

package terror

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"strconv"

	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"github.com/pingcap/parser/mysql"
)

// ErrCode represents a specific error type in a error class.
// Same error code can be used in different error classes.
type ErrCode int

const (
	// Executor error codes.

	// CodeUnknown is for errors of unknown reason.
	CodeUnknown ErrCode = -1
	// CodeExecResultIsEmpty indicates execution result is empty.
	CodeExecResultIsEmpty ErrCode = 3

	// Expression error codes.

	// CodeMissConnectionID indicates connection id is missing.
	CodeMissConnectionID ErrCode = 1

	// Special error codes.

	// CodeResultUndetermined indicates the sql execution result is undetermined.
	CodeResultUndetermined ErrCode = 2
)

// ErrClass represents a class of errors.
type ErrClass int

// Error classes.
const (
	ClassAutoid ErrClass = iota + 1
	ClassDDL
	ClassParser
	ClassTypes
	ClassDomain
	ClassEvaluator
	ClassExecutor
	ClassExpression
	ClassAdmin
	ClassKV
	ClassMeta
	ClassOptimizer
	ClassPerfSchema
	ClassPrivilege
	ClassSchema
	ClassServer
	ClassStructure
	ClassVariable
	ClassXEval
	ClassTable
	ClassGlobal
	ClassMockTikv
	ClassJSON
	ClassTiKV
	ClassSession
	ClassPlugin
	ClassUtil
	// Add more as needed.
)

var errClz2Str = map[ErrClass]string{
	ClassAutoid:     "autoid",
	ClassDDL:        "ddl",
	ClassDomain:     "domain",
	ClassExecutor:   "executor",
	ClassExpression: "expression",
	ClassAdmin:      "admin",
	ClassMeta:       "meta",
	ClassKV:         "kv",
	ClassOptimizer:  "planner",
	ClassParser:     "parser",
	ClassPerfSchema: "perfschema",
	ClassPrivilege:  "privilege",
	ClassSchema:     "schema",
	ClassServer:     "server",
	ClassStructure:  "structure",
	ClassVariable:   "variable",
	ClassTable:      "table",
	ClassTypes:      "types",
	ClassGlobal:     "global",
	ClassMockTikv:   "mocktikv",
	ClassJSON:       "json",
	ClassTiKV:       "tikv",
	ClassSession:    "session",
	ClassPlugin:     "plugin",
	ClassUtil:       "util",
}

// String implements fmt.Stringer interface.
func (ec ErrClass) String() string {
	if s, exists := errClz2Str[ec]; exists {
		return s
	}
	return strconv.Itoa(int(ec))
}

// EqualClass returns true if err is *Error with the same class.
func (ec ErrClass) EqualClass(err error) bool {
	e := errors.Cause(err)
	if e == nil {
		return false
	}
	if te, ok := e.(*BaseError); ok {
		if !ok {
			if c, cok := e.(BaseErrorConvertible); cok {
				te = c.ToBaseError()
				ok = true
			}
		}
		return te.class == ec
	}
	return false
}

// NotEqualClass returns true if err is not *Error with the same class.
func (ec ErrClass) NotEqualClass(err error) bool {
	return !ec.EqualClass(err)
}

// NewBaseError creates an *Error with an error code and an error message.
// Usually used to create base *Error.
// DO NOT CALL ME DIRECTLY, Please use terror.New
func (ec ErrClass) NewBaseError(code ErrCode, message string) *BaseError {
	return &BaseError{
		class:   ec,
		code:    code,
		message: message,
	}
}

// BaseErrorConvertible presents errors can be converted to terror.BaseError
type BaseErrorConvertible interface {
	ToBaseError() *BaseError
}

// Error implements error interface and adds integer Class and Code, so
// errors with different message can be compared.
type BaseError struct {
	class   ErrClass
	code    ErrCode
	message string
	args    []interface{}
	file    string
	line    int
}

// Class returns ErrClass
func (e *BaseError) Class() ErrClass {
	return e.class
}

// Code returns ErrCode
func (e *BaseError) Code() ErrCode {
	return e.code
}

// MarshalJSON implements json.Marshaler interface.
func (e *BaseError) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Class ErrClass `json:"class"`
		Code  ErrCode  `json:"code"`
		Msg   string   `json:"message"`
	}{
		Class: e.class,
		Code:  e.code,
		Msg:   e.GetMsg(),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (e *BaseError) UnmarshalJSON(data []byte) error {
	err := &struct {
		Class ErrClass `json:"class"`
		Code  ErrCode  `json:"code"`
		Msg   string   `json:"message"`
	}{}

	if err := json.Unmarshal(data, &err); err != nil {
		return errors.Trace(err)
	}

	e.class = err.Class
	e.code = err.Code
	e.message = err.Msg
	return nil
}

// Location returns the location where the error is created,
// implements juju/errors locationer interface.
func (e *BaseError) Location() (file string, line int) {
	return e.file, e.line
}

// Error implements error interface.
func (e *BaseError) Error() string {
	return fmt.Sprintf("[%s:%d]%s", e.class, e.code, e.GetMsg())
}

func (e *BaseError) GetMsg() string {
	if len(e.args) > 0 {
		return fmt.Sprintf(e.message, e.args...)
	}
	return e.message
}

func (e *BaseError) SetArgs(args []interface{}) {
	e.args = args
}

func (e *BaseError) SetMessage(format string) {
	e.message = format
}

// Equal checks if err is equal to e.
func (e *BaseError) Equal(err error) bool {
	originErr := errors.Cause(err)
	if originErr == nil {
		return false
	}

	if error(e) == originErr {
		return true
	}
	inErr, ok := originErr.(*BaseError)
	if !ok {
		if cErr, convOK := originErr.(BaseErrorConvertible); convOK {
			inErr = cErr.ToBaseError()
			ok = true
		}
	}
	return ok && e.class == inErr.class && e.code == inErr.code
}

// NotEqual checks if err is not equal to e.
func (e *BaseError) NotEqual(err error) bool {
	return !e.Equal(err)
}

// ErrorEqual returns a boolean indicating whether err1 is equal to err2.
func ErrorEqual(err1, err2 error) bool {
	e1 := errors.Cause(err1)
	e2 := errors.Cause(err2)

	if e1 == e2 {
		return true
	}

	if e1 == nil || e2 == nil {
		return e1 == e2
	}

	te1, ok1 := e1.(BaseErrorConvertible)
	te2, ok2 := e2.(BaseErrorConvertible)
	if ok1 && ok2 {
		return te1.ToBaseError().class == te2.ToBaseError().class && te1.ToBaseError().code == te2.ToBaseError().code
	}

	return e1.Error() == e2.Error()
}

// ErrorNotEqual returns a boolean indicating whether err1 isn't equal to err2.
func ErrorNotEqual(err1, err2 error) bool {
	return !ErrorEqual(err1, err2)
}

// MustNil cleans up and fatals if err is not nil.
func MustNil(err error, closeFuns ...func()) {
	if err != nil {
		for _, f := range closeFuns {
			f()
		}
		log.Fatal("unexpected error", zap.Error(err))
	}
}

// Call executes a function and checks the returned err.
func Call(fn func() error) {
	err := fn()
	if err != nil {
		log.Error("function call errored", zap.Error(err))
	}
}

// Log logs the error if it is not nil.
func Log(err error) {
	if err != nil {
		log.Error("encountered error", zap.Error(err))
	}
}

var (
	_ mysql.SQLErrorConvertible = &ParseError{}
)

// ParseError implements error interface and adds integer Class and Code, so
// errors with different message can be compared.
type ParseError struct {
	*BaseError
}

// NewStd calls New using the standard message for the error code
func NewStd(ec ErrClass, code ErrCode) *ParseError {
	return &ParseError{ec.NewBaseError(code, mysql.ParseErrName[uint16(code)])}
}

// New creates an *Error with an error code and an error message.
// Usually used to create base *Error.
func New(ec ErrClass, code ErrCode, message string) *ParseError {
	return &ParseError{ec.NewBaseError(code, message)}
}

// ToSQLError convert ParseError to mysql.SQLError.
func (e *ParseError) ToSQLError() *mysql.SQLError {
	terr := e.ToBaseError()
	code := getMySQLErrorCode(terr)
	return mysql.NewErrf(code, "%s", terr.GetMsg())
}

func (e *ParseError) ToBaseError() *BaseError {
	return e.BaseError
}

// GenWithStack generates a new *Error with the same class and code, and a new formatted message.
func (e *ParseError) GenWithStack(format string, args ...interface{}) error {
	bErr := *e.BaseError
	err := ParseError{BaseError: &bErr}
	err.message = format
	err.args = args
	return errors.AddStack(&err)
}

// GenWithStackByArgs generates a new *Error with the same class and code, and new arguments.
func (e *ParseError) GenWithStackByArgs(args ...interface{}) error {
	bErr := *e.BaseError
	err := ParseError{BaseError: &bErr}
	err.SetArgs(args)
	return errors.AddStack(&err)
}

// FastGen generates a new *Error with the same class and code, and a new formatted message.
// This will not call runtime.Caller to get file and line.
func (e *ParseError) FastGen(format string, args ...interface{}) error {
	bErr := *e.BaseError
	err := ParseError{BaseError: &bErr}
	err.SetMessage(format)
	err.SetArgs(args)
	return errors.SuspendStack(&err)
}

// FastGen generates a new *Error with the same class and code, and a new arguments.
// This will not call runtime.Caller to get file and line.
func (e *ParseError) FastGenByArgs(args ...interface{}) error {
	bErr := *e.BaseError
	err := ParseError{BaseError: &bErr}
	err.SetArgs(args)
	return errors.SuspendStack(&err)
}

var defaultMySQLErrorCode uint16

func getMySQLErrorCode(e *BaseError) uint16 {
	codeMap, ok := ErrClassToMySQLCodes[e.Class()]
	if !ok {
		log.Warn("Unknown error class", zap.Int("class", int(e.Class())))
		return defaultMySQLErrorCode
	}
	code, ok := codeMap[e.Code()]
	if !ok {
		log.Debug("Unknown error code", zap.Int("class", int(e.Class())), zap.Uint16("code", code))
		return defaultMySQLErrorCode
	}
	return code
}

var (
	// ErrClassToMySQLCodes is the map of ErrClass to code-map.
	ErrClassToMySQLCodes map[ErrClass]map[ErrCode]uint16
)

func init() {
	ErrClassToMySQLCodes = make(map[ErrClass]map[ErrCode]uint16)
	defaultMySQLErrorCode = mysql.ErrUnknown
}
