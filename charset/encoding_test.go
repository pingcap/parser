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

package charset_test

import (
	. "github.com/pingcap/check"
	"github.com/pingcap/parser/charset"
	"golang.org/x/text/transform"
)

var _ = Suite(&testEncodingSuite{})

type testEncodingSuite struct {
}

func (s *testEncodingSuite) TestEncoding(c *C) {
	enc := charset.NewEncoding("gbk")
	c.Assert(enc.Name(), Equals, "gbk")
	c.Assert(enc.Enabled(), IsTrue)
	enc.UpdateEncoding("utf-8")
	c.Assert(enc.Name(), Equals, "utf-8")
	enc.UpdateEncoding("gbk")
	c.Assert(enc.Name(), Equals, "gbk")
	c.Assert(enc.Enabled(), IsTrue)

	txt := "一二三四"
	e, _ := charset.Lookup("gbk")
	gbkEncodedTxt, _, err := transform.String(e.NewEncoder(), txt)
	c.Assert(err, IsNil)
	result, ok := enc.Decode([]byte(gbkEncodedTxt))
	c.Assert(ok, IsTrue)
	c.Assert(result, Equals, txt)

	gbkEncodedTxt2, ok := enc.Encode([]byte(txt))
	c.Assert(ok, IsTrue)
	c.Assert(gbkEncodedTxt, Equals, gbkEncodedTxt2)
	result, ok = enc.Decode([]byte(gbkEncodedTxt2))
	c.Assert(ok, IsTrue)
	c.Assert(result, Equals, txt)

	invalidGBK := []byte{0xE4, 0xB8, 0x80, 0xE4, 0xBA, 0x8C, 0xE4, 0xB8, 0x89}
	result, ok = enc.Decode(invalidGBK)
	c.Assert(ok, IsFalse)
	// MySQL reports "涓?簩涓" and throws a warning "Invalid gbk character string: '80E4BA'"
	// However, in GBK decoder provided by 'golang.org/x/text', '0x80' is decoded to '€' normally.
	c.Assert(result, Equals, "涓€?簩涓?")
}
