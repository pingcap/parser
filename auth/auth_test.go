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

package auth

import (
	"testing"

	. "github.com/pingcap/check"
)

var _ = Suite(&testAuthSuite{})

type testAuthSuite struct {
}

func TestT(t *testing.T) {
	TestingT(t)
}

func (s *testAuthSuite) TestEscapeAccountName(c *C) {
	c.Assert(EscapeAccountName(""), Equals, "''")
	c.Assert(EscapeAccountName("User"), Equals, "'User'")
	c.Assert(EscapeAccountName("User's"), Equals, "'User''s'")
	c.Assert(EscapeAccountName("User is me"), Equals, "'User is me'")
	c.Assert(EscapeAccountName(`u'v"w\'x\"y@z`+"`a"+`\b\\c`), Equals, "'u''v\"w\\''x\\\"y@z`a\\b\\\\c'") // u'v"\'x\"y@z`a\b\\c -> 'u''v"\''x\"y@z`a\b\\c'
	c.Assert(EscapeAccountName("u'v\"w\\'x\\\"y@z`a\\b\\\\c"), Equals, `'u''v"w\''x\"y@z`+"`"+`a\b\\c'`) // u'v"\'x\"y@z`a\b\\c -> 'u''v"\''x\"y@z`a\b\\c'
	u := UserIdentity{Username: "U & I @ Party", Hostname: "10.%", CurrentUser: false, AuthUsername: "root's friend", AuthHostname: "server"}
	c.Assert(u.String(), Equals, "'U & I @ Party'@'10.%'")
	c.Assert(u.AuthIdentityString(), Equals, "'root''s friend'@'server'")
	u = UserIdentity{Username: "", Hostname: "", CurrentUser: false, AuthUsername: "ceo", AuthHostname: "%"}
	c.Assert(u.String(), Equals, "''@''")
	c.Assert(u.AuthIdentityString(), Equals, "'ceo'@'%'")
	var uNil *UserIdentity = nil
	c.Assert(uNil.String(), Equals, "")
	r := RoleIdentity{Username: "Admin", Hostname: "192.168.%"}
	c.Assert(r.String(), Equals, "'Admin'@'192.168.%'")
}
