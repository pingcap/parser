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

package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/pingcap/errors"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/compatibility_reporter/sql_generator"
	"github.com/pingcap/parser/compatibility_reporter/yacc_parser"
	"github.com/pingcap/parser/terror"
	_ "github.com/pingcap/tidb/types/parser_driver"
)

var (
	output         = flag.String("o", "./report.csv", "Output path of csv format report")
	printAll       = flag.Bool("a", false, "Output all test case, regardless of success or failure")
	mysqlUser      = flag.String("u", "root", "MySQL User for login")
	mysqlPassword  = flag.String("p", "", "Password to use when connecting to MySQL server")
	mysqlHost      = flag.String("h", "127.0.0.1", "Connect to MySQL host")
	mysqlPort      = flag.Int("P", 3306, "Port number to use for MySQL connection")
	productionName = flag.String("n", "", "Production name to test")
	bnfPath        = flag.String("b", "", "BNF file path")
	MySQLVersion   = "None"
)

type caseReport struct {
	Sql        string
	MySQLPass  bool
	MySQLErr   error
	MySQLErrNo uint16
	TiDBPass   bool
	TiDBWarns  []error
	TiDBErr    error
	TiDBErrNo  uint16
}

func mysqlParserTest(mysqlSource *sql.DB, report *caseReport) {
	_, parserErr := mysqlSource.Query(report.Sql)
	if parserErr == nil {
		report.MySQLPass = true
		return
	}
	mysqlErr, success := parserErr.(*mysql.MySQLError)
	if !success {
		panic("MySQL client error:" + parserErr.Error())
	}
	// number 1064 is mysql server errno, it means parser error
	// see: https://dev.mysql.com/doc/refman/8.0/en/server-error-reference.html#error_er_parse_error
	report.MySQLPass = mysqlErr.Number != 1064
	report.MySQLErr = mysqlErr
	report.MySQLErrNo = mysqlErr.Number
}

func tidbParserTest(tidbParser *parser.Parser, report *caseReport) {
	stmtNodes, parserWarns, parserErr := tidbParser.Parse(report.Sql, "", "")
	report.TiDBPass = stmtNodes != nil && len(stmtNodes) > 0 && parserErr == nil
	report.TiDBWarns = parserWarns
	report.TiDBErr = parserErr
	if parserErr != nil {
		if pErr, ok := errors.Unwrap(parserErr).(*terror.Error); ok {
			report.TiDBErrNo = uint16(pErr.Code())
		}
	}
}

func printCsvHead(csvFile *os.File) {

	_, writeErr := csvFile.WriteString(fmt.Sprintf("TiDB Parser Git Hash,%s\n", parser.TiDBParserGitHash))
	if writeErr != nil {
		panic(fmt.Sprintf("file(%s) write failure: %s", *output, writeErr.Error()))
	}
	_, writeErr = csvFile.WriteString(fmt.Sprintf("TiDB Parser Git Branch,%s\n", parser.TiDBParserGitBranch))
	if writeErr != nil {
		panic(fmt.Sprintf("file(%s) write failure: %s", *output, writeErr.Error()))
	}
	_, writeErr = csvFile.WriteString(fmt.Sprintf("MySQL Version,%s\n", MySQLVersion))
	if writeErr != nil {
		panic(fmt.Sprintf("file(%s) write failure: %s", *output, writeErr.Error()))
	}
	_, writeErr = csvFile.WriteString("sql,mysql_pass,mysql_err,tidb_pass,tidb_warns,tidb_err\n")
	if writeErr != nil {
		panic(fmt.Sprintf("file(%s) write failure: %s", *output, writeErr.Error()))
	}
}

func printCsvCaseReport(csvFile *os.File, report *caseReport) {
	if !*printAll &&
		((report.TiDBPass && report.MySQLPass) ||
			(report.TiDBErrNo == report.MySQLErrNo)) {
		return
	}
	var tidbWarns []string
	for _, warn := range report.TiDBWarns {
		tidbWarns = append(tidbWarns, warn.Error())
	}
	tidbWarnsStr := strings.Join(tidbWarns, ";")
	_, writeErr := csvFile.WriteString(fmt.Sprintf("%s,%t,%s,%t,%s,%s\n",
		escapeString(report.Sql),
		report.MySQLPass,
		escapeErrorString(report.MySQLErr),
		report.TiDBPass,
		escapeString(tidbWarnsStr),
		escapeErrorString(report.TiDBErr)))
	if writeErr != nil {
		panic(fmt.Sprintf("file(%s) write failure: %s", *output, writeErr.Error()))
	}
}

func printCsvSummary(csvFile *os.File, totalCases uint64, tidbPassCases uint64, mysqlPassCases uint64, incompatibleCases uint64) {
	_, writeErr := csvFile.WriteString(fmt.Sprintf("totalCases,%d,tidbPassCases,%d,mysqlPassCases,%d,incompatibleCases,%d\n",
		totalCases,
		tidbPassCases,
		mysqlPassCases,
		incompatibleCases,
	))
	if writeErr != nil {
		panic(fmt.Sprintf("file(%s) write failure: %s", *output, writeErr.Error()))
	}
}

func escapeErrorString(err error) string {
	if err == nil {
		return ""
	}
	return escapeString(err.Error())
}

func escapeString(str string) string {
	str = strings.ReplaceAll(str, "\"", "\"\"")
	str = "\"" + str + "\""
	return str
}

func main() {
	flag.Parse()

	bnfs := strings.Split(*bnfPath, " ")
	var allProductions []yacc_parser.Production
	for _, bnf := range bnfs {
		bnfFile, err := os.Open(bnf)
		if err != nil {
			panic(fmt.Sprintf("File '%s' open failure", bnf))
		}
		productions := yacc_parser.Parse(yacc_parser.Tokenize(bufio.NewReader(bnfFile)))
		allProductions = append(allProductions, productions...)
	}

	sqlIter := sql_generator.GenerateSQL(allProductions, *productionName)

	db, dbErr := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/mysql",
		*mysqlUser, *mysqlPassword, *mysqlHost, *mysqlPort))
	if dbErr != nil {
		panic("MySQL client error:" + dbErr.Error())
	}
	defer db.Close()
	row := db.QueryRow("select version()")
	queryErr := row.Scan(&MySQLVersion)
	if queryErr != nil {
		panic("MySQL client error:" + queryErr.Error())
	}

	tidbParser := parser.New()

	csvFile, fileErr := os.OpenFile(*output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if fileErr != nil {
		panic(fmt.Sprintf("file '%s' open failure: %s", *output, fileErr.Error()))
	}
	defer csvFile.Close()
	printCsvHead(csvFile)
	var totalCases uint64 = 0
	var tidbPassCases uint64 = 0
	var mysqlPassCases uint64 = 0
	var incompatibleCases uint64 = 0
	for sqlIter.HasNext() {
		report := caseReport{
			Sql: sqlIter.Next(),
		}
		tidbParserTest(tidbParser, &report)
		mysqlParserTest(db, &report)
		printCsvCaseReport(csvFile, &report)
		totalCases++
		if report.TiDBPass {
			tidbPassCases++
		}
		if report.MySQLPass {
			mysqlPassCases++
		}
		if report.MySQLPass != report.TiDBPass {
			incompatibleCases++
		}
	}
	printCsvSummary(csvFile, totalCases, tidbPassCases, mysqlPassCases, incompatibleCases)
}
