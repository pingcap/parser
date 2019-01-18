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

package parser_test

import (
	. "github.com/pingcap/check"
	"github.com/pingcap/parser"
)

var _ = Suite(&testSQLDigestSuite{})

type testSQLDigestSuite struct {
}

func (s *testSQLDigestSuite) TestDigestText(c *C) {
	tests := []struct {
		input  string
		expect string
	}{
		{"SELECT 1", "select ?"},
		{"select * from b where id = 1", "select * from b where id = ?"},
		{"select 1 from b where id in (1, 3, '3', 1, 2, 3, 4)", "select ? from b where id in ( ... )"},
		{"select 1 from b where id in (1, a, 4)", "select ? from b where id in ( ? , a , ? )"},
		{"select 1 from b order by 2", "select ? from b order by 2"},
		{"select /*+ a hint */ 1", "select ?"},
	}
	for _, test := range tests {
		actual := parser.DigestText(test.input)
		c.Assert(actual, Equals, test.expect)
	}
}

func (s *testSQLDigestSuite) TestDigestHashEqForSimpleSQL(c *C) {
	sqlGroups := [][]string{
		{"select * from b where id = 1", "select * from b where id = '1'", "select * from b where id =2"},
		{"select 2 from b, c where b.id = c.id where c.id > 1", "select 4 from b, c where b.id = c.id where c.id > 23"},
		{"Select 3", "select 1"},
	}
	for _, sqlGroup := range sqlGroups {
		var d string
		for _, sql := range sqlGroup {
			dig := parser.DigestHash(sql)
			if d == "" {
				d = dig
				continue
			}
			c.Assert(d, Equals, dig)
		}
	}
}

func (s *testSQLDigestSuite) TestDigestHashNotEqForSimpleSQL(c *C) {
	sqlGroups := [][]string{
		{"select * from b where id = 1", "select a from b where id = 1", "select * from d where bid =1"},
	}
	for _, sqlGroup := range sqlGroups {
		var d string
		for _, sql := range sqlGroup {
			dig := parser.DigestHash(sql)
			if d == "" {
				d = dig
				continue
			}
			c.Assert(d, Not(Equals), dig)
		}
	}
}
