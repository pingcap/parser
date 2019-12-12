package test_driver

import (
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/terror"
)

var (
	// ErrTruncated is returned when data has been truncated during conversion.
	ErrTruncated = terror.ClassTypes.New(mysql.WarnDataTruncated, mysql.MySQLErrName[mysql.WarnDataTruncated])
	// ErrOverflow is returned when data is out of range for a field type.
	ErrOverflow = terror.ClassTypes.New(mysql.ErrDataOutOfRange, mysql.MySQLErrName[mysql.ErrDataOutOfRange])
	// ErrBadNumber is return when parsing an invalid binary decimal number.
	ErrBadNumber = terror.ClassTypes.New(mysql.ErrBadNumber, mysql.MySQLErrName[mysql.ErrBadNumber])
)
