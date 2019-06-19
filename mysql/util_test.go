package mysql

import (
	. "github.com/pingcap/check"
)

var _ = Suite(&testTypeSuite{})

func (s *testTypeSuite) TestDecimalDefaultValue(c *C) {
	len, _ := GetDefaultFieldLengthAndDecimal(TypeDecimal)
	c.Assert(len, Equals, 10)

	len, _ = GetDefaultFieldLengthAndDecimalForCast(TypeDecimal)
	c.Assert(len, Equals, 10)
}
