package mysql

import (
	. "github.com/pingcap/check"
)

var _ = Suite(&testUtilSuite{})

type testUtilSuite struct{}

func (s *testUtilSuite) TestDecimalDefaultValue(c *C) {
	len, _ := GetDefaultFieldLengthAndDecimal(TypeNewDecimal)
	c.Assert(len, Equals, 10)

	len, _ = GetDefaultFieldLengthAndDecimalForCast(TypeNewDecimal)
	c.Assert(len, Equals, 10)
}
