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
	"strconv"
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
	output          string
	printAll        bool
	mysqlUser       string
	mysqlPassword   string
	mysqlHost       string
	mysqlPort       int
	productionName  string
	bnfPath         string
	randomlyGen     bool
	totalOutputCase uint64
	database        string
	exceptErrnoStr  string
	exceptErrno     []uint16
	MySQLVersion    = "None"
)

func parseFlag() {
	flag.StringVar(&output, "o", "./report.csv", "Output path of csv format report")
	flag.BoolVar(&printAll, "a", false, "Output all test case, regardless of success or failure")
	flag.StringVar(&mysqlUser, "u", "root", "MySQL User for login")
	flag.StringVar(&mysqlPassword, "p", "", "Password to use when connecting to MySQL server")
	flag.StringVar(&mysqlHost, "h", "127.0.0.1", "Connect to MySQL host")
	flag.IntVar(&mysqlPort, "P", 3306, "Port number to use for MySQL connection")
	flag.StringVar(&productionName, "n", "", "Production name to test")
	flag.StringVar(&bnfPath, "b", "", "BNF file path")
	flag.BoolVar(&randomlyGen, "R", false, "Generator SQL randomly")
	flag.Uint64Var(&totalOutputCase, "N", 0, "The number of output sql case, set 0 for infinite")
	flag.StringVar(&database, "d", "mysql", "The database selected after connected")
	flag.StringVar(&exceptErrnoStr, "E", "", "Except cases of these error codes")
	flag.Parse()
	if len(productionName) == 0 || len(bnfPath) == 0 {
		flag.Usage()
		os.Exit(0)
	}
	errnoStrs := strings.Split(exceptErrnoStr, ",")
	for _, errnoStr := range errnoStrs {
		if errnoStr == "" {
			continue
		}
		errno, err := strconv.ParseUint(errnoStr, 10, 16)
		if err != nil {
			panic("parameter error:" + errnoStr)
		}
		exceptErrno = append(exceptErrno, uint16(errno))
	}
}

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
	rows, parserErr := mysqlSource.Query(report.Sql)
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	if parserErr == nil {
		report.MySQLPass = true
		return
	}
	mysqlErr, success := parserErr.(*mysql.MySQLError)
	if !success {
		panic("MySQL client error:" + parserErr.Error() + "sql: " + report.Sql)
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
		panic(fmt.Sprintf("file(%s) write failure: %s", output, writeErr.Error()))
	}
	_, writeErr = csvFile.WriteString(fmt.Sprintf("TiDB Parser Git Branch,%s\n", parser.TiDBParserGitBranch))
	if writeErr != nil {
		panic(fmt.Sprintf("file(%s) write failure: %s", output, writeErr.Error()))
	}
	_, writeErr = csvFile.WriteString(fmt.Sprintf("MySQL Version,%s\n", MySQLVersion))
	if writeErr != nil {
		panic(fmt.Sprintf("file(%s) write failure: %s", output, writeErr.Error()))
	}
	_, writeErr = csvFile.WriteString("sql,mysql_pass,mysql_err,tidb_pass,tidb_warns,tidb_err\n")
	if writeErr != nil {
		panic(fmt.Sprintf("file(%s) write failure: %s", output, writeErr.Error()))
	}
}

var outputCaseNum uint64

func printCsvCaseReport(csvFile *os.File, report *caseReport) {
	if !printAll &&
		((report.TiDBPass && report.MySQLPass) ||
			(report.TiDBErrNo == report.MySQLErrNo)) ||
		!report.MySQLPass {
		return
	}
	for _, errno := range exceptErrno {
		if errno == report.MySQLErrNo {
			return
		}
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
	outputCaseNum++
	if writeErr != nil {
		panic(fmt.Sprintf("file(%s) write failure: %s", output, writeErr.Error()))
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
		panic(fmt.Sprintf("file(%s) write failure: %s", output, writeErr.Error()))
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
	parseFlag()
	bnfs := strings.Split(bnfPath, " ")
	var allProductions []yacc_parser.Production
	for _, bnf := range bnfs {
		bnfFile, err := os.Open(bnf)
		if err != nil {
			panic(fmt.Sprintf("File '%s' open failure", bnf))
		}
		productions := yacc_parser.Parse(yacc_parser.Tokenize(bufio.NewReader(bnfFile)))
		allProductions = append(allProductions, productions...)
	}
	var sqlIter sql_generator.SQLIterator
	if randomlyGen {
		sqlIter = sql_generator.GenerateSQLRandomly(allProductions, productionName)
	} else {
		sqlIter = sql_generator.NewSQLEnumIterator(allProductions, productionName)
	}

	db, dbErr := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		mysqlUser, mysqlPassword, mysqlHost, mysqlPort, database))
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

	csvFile, fileErr := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if fileErr != nil {
		panic(fmt.Sprintf("file '%s' open failure: %s", output, fileErr.Error()))
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
		if !(report.MySQLPass && report.TiDBPass) && (report.MySQLErrNo != report.TiDBErrNo) {
			incompatibleCases++
		}
		if totalOutputCase != 0 && outputCaseNum >= totalOutputCase {
			break
		}
	}
	printCsvSummary(csvFile, totalCases, tidbPassCases, mysqlPassCases, incompatibleCases)
}
