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

// MySQL error code.
// This value is numeric. It is not portable to other database systems.
const (
	ErrSyntax                  = 1149
	ErrParse                   = 1064
	ErrUnknownCharacterSet     = 1115
	ErrInvalidYearColumnLength = 1818
	ErrWrongArguments          = 1210
	ErrWrongFieldTerminators   = 1083
	ErrTooBigDisplaywidth      = 1439
	ErrTooBigPrecision         = 1426
	ErrUnknownAlterLock        = 1801
	ErrUnknownAlterAlgorithm   = 1800

	ErrInvalidDefault = 1067

	ErrNoParts                              = 1504
	ErrPartitionColumnList                  = 1653
	ErrPartitionRequiresValues              = 1479
	ErrPartitionsMustBeDefined              = 1492
	ErrPartitionWrongNoPart                 = 1484
	ErrPartitionWrongNoSubpart              = 1485
	ErrPartitionWrongValues                 = 1480
	ErrRowSinglePartitionField              = 1658
	ErrSubpartition                         = 1500
	ErrSystemVersioningWrongPartitions      = 4128
	ErrTooManyValues                        = 1657
	ErrWrongPartitionTypeExpectedSystemTime = 4113

	ErrWarnOptimizerHintUnsupportedHint = 8061
	ErrWarnOptimizerHintInvalidToken    = 8062
	ErrWarnMemoryQuotaOverflow          = 8063
	ErrWarnOptimizerHintParseError      = 8064
	ErrWarnOptimizerHintInvalidInteger  = 8065

	ErrUnknownCollation         = 1273
	ErrCollationCharsetMismatch = 1253

	ErrUnknown = 1105
)
