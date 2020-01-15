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

package mysql

const (
	// DefaultMySQLState is default state of the mySQL
	DefaultMySQLState = "HY000"
)

// ParserState maps error code to MySQL SQLSTATE value.
// The values are taken from ANSI SQL and ODBC and are more standardized.
var ParserState = map[uint16]string{
	ErrParse:                    "42000",
	ErrInvalidDefault:           "42000",
	ErrWrongFieldTerminators:    "42000",
	ErrUnknownCharacterSet:      "42000",
	ErrSyntax:                   "42000",
	ErrCollationCharsetMismatch: "42000",
	ErrTooBigPrecision:          "42000",
	ErrTooBigDisplaywidth:       "42000",
}
