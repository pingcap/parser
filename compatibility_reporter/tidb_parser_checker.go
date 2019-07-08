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
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pingcap/parser"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"os"
	"strings"
)

var (
	printAll  = flag.Bool("a", false, "Print all test results, regardless of success or failure.")
	printWarn = flag.Bool("w", false, "Print error and warn test results.")
)

type caseReport struct {
	Sql   string  `json:"sql"`
	Pass  bool    `json:"pass"`
	Warns []error `json:"warns"`
	Err   error   `json:"err"`
}

func main() {
	flag.Parse()
	reader := bufio.NewReader(os.Stdin)
	p := parser.New()
	for true {
		sql, scanErr := reader.ReadString('\n')
		if scanErr != nil {
			fmt.Println(scanErr)
			break
		}
		sql = strings.TrimSpace(sql)
		stmtNodes, parserWarns, parserErr := p.Parse(sql, "", "")
		report := caseReport{
			Sql:   sql,
			Pass:  stmtNodes != nil && len(stmtNodes) > 0 && parserErr == nil,
			Warns: parserWarns,
			Err:   parserErr,
		}
		if !report.Pass || *printAll || (report.Warns != nil && *printWarn) {
			jsonBytes, jsonErr := json.Marshal(report)
			if jsonErr != nil {
				panic("Json serialize failure: " + jsonErr.Error())
			}
			fmt.Println(string(jsonBytes))
		}
	}
}
