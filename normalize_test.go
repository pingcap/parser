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
	"testing"

	"github.com/pingcap/parser"
)

func TestDigestEqForSimpleSQL(t *testing.T) {
	sqlGroups := [][]string{
		{"select * from b where id = 1", "select * from b where id = '1'", "select * from b where id =2"},
		{"select 2 from b, c where b.id =          c.id where c.id > 1", "select 4 from b, c where " +
			"b.id = c.id where c.id > 23"},
	}
	for _, sqlGroup := range sqlGroups {
		var d string
		for _, sql := range sqlGroup {
			dig := parser.Digest(sql)
			if d == "" {
				d = dig
				continue
			}
			if d != dig {
				t.Errorf("digest for %s's digest result %s not eq to previous %s", sql, d, dig)
			}
		}
	}
}

func TestDigestNotEqForSimpleSQL(t *testing.T) {
	sqlGroups := [][]string{
		{"select * from b where id = 1", "select a from b where id = 1", "select * from d where bid =1"},
	}
	for _, sqlGroup := range sqlGroups {
		var d string
		for _, sql := range sqlGroup {
			dig := parser.Digest(sql)
			if d == "" {
				d = dig
				continue
			}
			if d == dig {
				t.Errorf("digest for %s's digest result %s not eq to previous %s", sql, d, dig)
			}
		}
	}
}
