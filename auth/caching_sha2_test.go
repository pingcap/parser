// Copyright 2021 PingCAP, Inc.
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
	"encoding/hex"
	"strings"

	. "github.com/pingcap/check"
)

func (s *testAuthSuite) TestCheckShaPasswordGood(c *C) {
	pwd := "foobar"
	pwhash, _ := hex.DecodeString("24412430303524031A69251C34295C4B35167C7F1E5A7B63091349503974624D34504B5A424679354856336868686F52485A736E4A733368786E427575516C73446469496537")
	r, err := CheckShaPassword(pwhash, pwd)
	c.Assert(err, IsNil)
	c.Assert(r, IsTrue)
}

func (s *testAuthSuite) TestCheckShaPasswordBad(c *C) {
	pwd := "not_foobar"
	pwhash, _ := hex.DecodeString("24412430303524031A69251C34295C4B35167C7F1E5A7B63091349503974624D34504B5A424679354856336868686F52485A736E4A733368786E427575516C73446469496537")
	r, err := CheckShaPassword(pwhash, pwd)
	c.Assert(err, IsNil)
	c.Assert(r, IsFalse)
}

func (s *testAuthSuite) TestCheckShaPasswordShort(c *C) {
	pwd := "not_foobar"
	pwhash, _ := hex.DecodeString("aaaaaaaa")
	_, err := CheckShaPassword(pwhash, pwd)
	c.Assert(err, NotNil)
}

func (s *testAuthSuite) TestCheckShaPasswordDigetTypeIncompatible(c *C) {
	pwd := "not_foobar"
	pwhash, _ := hex.DecodeString("24422430303524031A69251C34295C4B35167C7F1E5A7B63091349503974624D34504B5A424679354856336868686F52485A736E4A733368786E427575516C73446469496537")
	_, err := CheckShaPassword(pwhash, pwd)
	c.Assert(err, NotNil)
}

func (s *testAuthSuite) TestCheckShaPasswordIterationsInvalid(c *C) {
	pwd := "not_foobar"
	pwhash, _ := hex.DecodeString("24412430304124031A69251C34295C4B35167C7F1E5A7B63091349503974624D34504B5A424679354856336868686F52485A736E4A733368786E427575516C73446469496537")
	_, err := CheckShaPassword(pwhash, pwd)
	c.Assert(err, NotNil)
}

// The output from NewSha2Password is not stable as the hash is based on the genrated salt.
// This is why CheckShaPassword is used here.
func (s *testAuthSuite) TestNewSha2Password(c *C) {
	pwd := "testpwd"
	pwhash := NewSha2Password(pwd)
	r, err := CheckShaPassword([]byte(pwhash), pwd)
	c.Assert(err, IsNil)
	c.Assert(r, IsTrue)

	for r := range pwhash {
		c.Assert(pwhash[r], Less, uint8(128))
		c.Assert(pwhash[r], Not(Equals), 0)  // NUL
		c.Assert(pwhash[r], Not(Equals), 36) // '$'
	}
}

func (s *testAuthSuite) TestGenerateScramble(c *C) {
	pwd := []byte("abc")
	nonce, _ := hex.DecodeString("6d642b676464321c561d476e094c316d05180b16")
	storedScramble, _ := hex.DecodeString("3455aae998ae2959d7170229375c9626735bf545d2a828a7d45f94f9b2cf19f4")
	scramble, err := GenerateScramble(pwd, nonce)
	c.Assert(err, IsNil)
	c.Assert(scramble, BytesEquals, storedScramble)

	_, err = GenerateScramble([]byte(strings.Repeat("x", MAX_PASSWORD_LENGTH+1)), nonce)
	c.Assert(err, NotNil)

	_, err = GenerateScramble(pwd, append(nonce, byte('A')))
	c.Assert(err, NotNil)

	_, err = GenerateScramble(pwd, nonce[:SALT_LENGTH-1])
	c.Assert(err, NotNil)
}
