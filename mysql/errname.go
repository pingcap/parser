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

// ParseErrName maps error code to MySQL error messages.
var ParseErrName = map[uint16]string{
	ErrSyntax:                  "You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use",
	ErrParse:                   "%s %s",
	ErrUnknownCharacterSet:     "Unknown character set: '%-.64s'",
	ErrInvalidYearColumnLength: "Supports only YEAR or YEAR(4) column",
	ErrWrongArguments:          "Incorrect arguments to %s",
	ErrWrongFieldTerminators:   "Field separator argument is not what is expected; check the manual",
	ErrTooBigDisplaywidth:      "Display width out of range for column '%-.192s' (max = %d)",
	ErrTooBigPrecision:         "Too big precision %d specified for column '%-.192s'. Maximum is %d.",
	ErrUnknownAlterLock:        "Unknown LOCK type '%s'",
	ErrUnknownAlterAlgorithm:   "Unknown ALGORITHM '%s'",
	ErrInvalidDefault:          "Invalid default value for '%-.192s'",

	ErrNoParts:                              "Number of %-.64s = 0 is not an allowed value",
	ErrPartitionColumnList:                  "Inconsistency in usage of column lists for partitioning",
	ErrPartitionRequiresValues:              "Syntax : %-.64s PARTITIONING requires definition of VALUES %-.64s for each partition",
	ErrPartitionsMustBeDefined:              "For %-.64s partitions each partition must be defined",
	ErrPartitionWrongNoPart:                 "Wrong number of partitions defined, mismatch with previous setting",
	ErrPartitionWrongNoSubpart:              "Wrong number of subpartitions defined, mismatch with previous setting",
	ErrPartitionWrongValues:                 "Only %-.64s PARTITIONING can use VALUES %-.64s in partition definition",
	ErrRowSinglePartitionField:              "Row expressions in VALUES IN only allowed for multi-field column partitioning",
	ErrSubpartition:                         "It is only possible to mix RANGE/LIST partitioning with HASH/KEY partitioning for subpartitioning",
	ErrSystemVersioningWrongPartitions:      "Wrong Partitions: must have at least one HISTORY and exactly one last CURRENT",
	ErrTooManyValues:                        "Cannot have more than one value for this type of %-.64s partitioning",
	ErrWrongPartitionTypeExpectedSystemTime: "Wrong partitioning type, expected type: `SYSTEM_TIME`",

	ErrUnknownCollation:         "Unknown collation: '%-.64s'",
	ErrCollationCharsetMismatch: "COLLATION '%s' is not valid for CHARACTER SET '%s'",

	ErrWarnOptimizerHintInvalidInteger:  "integer value is out of range in '%s'",
	ErrWarnOptimizerHintUnsupportedHint: "Optimizer hint %s is not supported by TiDB and is ignored",
	ErrWarnOptimizerHintInvalidToken:    "Cannot use %s '%s' (tok = %d) in an optimizer hint",
	ErrWarnMemoryQuotaOverflow:          "Max value of MEMORY_QUOTA is %d bytes, ignore this invalid limit",
	ErrWarnOptimizerHintParseError:      "Optimizer hint syntax error at %v",

	ErrUnknown: "Unknown error",
}
