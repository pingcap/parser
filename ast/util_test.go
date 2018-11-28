// Copyright 2017 PingCAP, Inc.
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

package ast

import (
	. "github.com/pingcap/check"
)

var _ = Suite(&testUtilSuite{})

type testUtilSuite struct {
}

func (s *testCacheableSuite) TestSQLSentence(c *C) {
	ss := NewSQLSentence()
	ss.Text("SELECT").Space().TextAround("col1", "`").Space().
		Text("FROM").Space().TextAround("table1", "`").TextAround("WHERE", " ").
		Text("col2=1")
	c.Assert(ss.String(), Equals, "SELECT `col1` FROM `table1` WHERE col2=1")
	ss2 := NewSQLSentence()
	ss2.Text("SELECT").Space().LeftBracket().TextAround("col3", "`").TextAround("col5", "`").
		TextAround("col6", "`").RightBracket().TextAround("FROM", " ").
		TextAround("table2", "`").TextAround("WHERE", " ").Text("col4=").
		Text("(").Sentence(ss).Text(")")
	c.Assert(ss2.String(), Equals,
		"SELECT (`col3`, `col5`, `col6`) FROM `table2` WHERE col4=(SELECT `col1` FROM `table1` WHERE col2=1)")
	ss3 := NewSQLSentence()
	ss3.SentenceAround(ss2, "#")
	c.Assert(ss3.String(), Equals,
		"#SELECT (`col3`, `col5`, `col6`) FROM `table2` WHERE col4=(SELECT `col1` FROM `table1` WHERE col2=1)#")

}
