package test_driver

import (
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/terror"
)

// const strings for ErrWrongValue
const (
	TimeStr = "time"
)

var (
	// ErrTruncated is returned when data has been truncated during conversion.
	ErrTruncated = terror.ClassTypes.New(mysql.WarnDataTruncated, mysql.MySQLErrName[mysql.WarnDataTruncated])
	// ErrOverflow is returned when data is out of range for a field type.
	ErrOverflow = terror.ClassTypes.New(mysql.ErrDataOutOfRange, mysql.MySQLErrName[mysql.ErrDataOutOfRange])
	// ErrDivByZero is return when do division by 0.
	ErrDivByZero = terror.ClassTypes.New(mysql.ErrDivisionByZero, mysql.MySQLErrName[mysql.ErrDivisionByZero])
	// ErrTooBigDisplayWidth is return when display width out of range for column.
	ErrTooBigDisplayWidth = terror.ClassTypes.New(mysql.ErrTooBigDisplaywidth, mysql.MySQLErrName[mysql.ErrTooBigDisplaywidth])
	// ErrTooBigFieldLength is return when column length too big for column.
	ErrTooBigFieldLength = terror.ClassTypes.New(mysql.ErrTooBigFieldlength, mysql.MySQLErrName[mysql.ErrTooBigFieldlength])
	// ErrTooBigSet is returned when too many strings for column.
	ErrTooBigSet = terror.ClassTypes.New(mysql.ErrTooBigSet, mysql.MySQLErrName[mysql.ErrTooBigSet])
	// ErrTooBigScale is returned when type DECIMAL/NUMERIC scale is bigger than mysql.MaxDecimalScale.
	ErrTooBigScale = terror.ClassTypes.New(mysql.ErrTooBigScale, mysql.MySQLErrName[mysql.ErrTooBigScale])
	// ErrTooBigPrecision is returned when type DECIMAL/NUMERIC precision is bigger than mysql.MaxDecimalWidth
	ErrTooBigPrecision = terror.ClassTypes.New(mysql.ErrTooBigPrecision, mysql.MySQLErrName[mysql.ErrTooBigPrecision])
	// ErrBadNumber is return when parsing an invalid binary decimal number.
	ErrBadNumber = terror.ClassTypes.New(mysql.ErrBadNumber, mysql.MySQLErrName[mysql.ErrBadNumber])
	// ErrInvalidFieldSize is returned when the precision of a column is out of range.
	ErrInvalidFieldSize = terror.ClassTypes.New(mysql.ErrInvalidFieldSize, mysql.MySQLErrName[mysql.ErrInvalidFieldSize])
	// ErrMBiggerThanD is returned when precision less than the scale.
	ErrMBiggerThanD = terror.ClassTypes.New(mysql.ErrMBiggerThanD, mysql.MySQLErrName[mysql.ErrMBiggerThanD])
	// ErrWarnDataOutOfRange is returned when the value in a numeric column that is outside the permissible range of the column data type.
	// See https://dev.mysql.com/doc/refman/5.5/en/out-of-range-and-overflow.html for details
	ErrWarnDataOutOfRange = terror.ClassTypes.New(mysql.ErrWarnDataOutOfRange, mysql.MySQLErrName[mysql.ErrWarnDataOutOfRange])
	// ErrDuplicatedValueInType is returned when enum column has duplicated value.
	ErrDuplicatedValueInType = terror.ClassTypes.New(mysql.ErrDuplicatedValueInType, mysql.MySQLErrName[mysql.ErrDuplicatedValueInType])
	// ErrDatetimeFunctionOverflow is returned when the calculation in datetime function cause overflow.
	ErrDatetimeFunctionOverflow = terror.ClassTypes.New(mysql.ErrDatetimeFunctionOverflow, mysql.MySQLErrName[mysql.ErrDatetimeFunctionOverflow])
	// ErrCastAsSignedOverflow is returned when positive out-of-range integer, and convert to it's negative complement.
	ErrCastAsSignedOverflow = terror.ClassTypes.New(mysql.ErrCastAsSignedOverflow, mysql.MySQLErrName[mysql.ErrCastAsSignedOverflow])
	// ErrCastNegIntAsUnsigned is returned when a negative integer be casted to an unsigned int.
	ErrCastNegIntAsUnsigned = terror.ClassTypes.New(mysql.ErrCastNegIntAsUnsigned, mysql.MySQLErrName[mysql.ErrCastNegIntAsUnsigned])
	// ErrInvalidYearFormat is returned when the input is not a valid year format.
	ErrInvalidYearFormat = terror.ClassTypes.New(mysql.ErrInvalidYearFormat, mysql.MySQLErrName[mysql.ErrInvalidYearFormat])
	// ErrInvalidYear is returned when the input value is not a valid year.
	ErrInvalidYear = terror.ClassTypes.New(mysql.ErrInvalidYear, mysql.MySQLErrName[mysql.ErrInvalidYear])
	// ErrTruncatedWrongVal is returned when data has been truncated during conversion.
	ErrTruncatedWrongVal = terror.ClassTypes.New(mysql.ErrTruncatedWrongValue, mysql.MySQLErrName[mysql.ErrTruncatedWrongValue])
	// ErrInvalidWeekModeFormat is returned when the week mode is wrong.
	ErrInvalidWeekModeFormat = terror.ClassTypes.New(mysql.ErrInvalidWeekModeFormat, mysql.MySQLErrName[mysql.ErrInvalidWeekModeFormat])
	// ErrWrongValue is returned when the input value is in wrong format.
	ErrWrongValue = terror.ClassTypes.New(mysql.ErrWrongValue, mysql.MySQLErrName[mysql.ErrWrongValue])
)

var (
	// ErrInvalidJSONText means invalid JSON text.
	ErrInvalidJSONText = terror.ClassJSON.New(mysql.ErrInvalidJSONText, mysql.MySQLErrName[mysql.ErrInvalidJSONText])
	// ErrInvalidJSONPath means invalid JSON path.
	ErrInvalidJSONPath = terror.ClassJSON.New(mysql.ErrInvalidJSONPath, mysql.MySQLErrName[mysql.ErrInvalidJSONPath])
	// ErrInvalidJSONData means invalid JSON data.
	ErrInvalidJSONData = terror.ClassJSON.New(mysql.ErrInvalidJSONData, mysql.MySQLErrName[mysql.ErrInvalidJSONData])
	// ErrInvalidJSONPathWildcard means invalid JSON path that contain wildcard characters.
	ErrInvalidJSONPathWildcard = terror.ClassJSON.New(mysql.ErrInvalidJSONPathWildcard, mysql.MySQLErrName[mysql.ErrInvalidJSONPathWildcard])
	// ErrInvalidJSONContainsPathType means invalid JSON contains path type.
	ErrInvalidJSONContainsPathType = terror.ClassJSON.New(mysql.ErrInvalidJSONContainsPathType, mysql.MySQLErrName[mysql.ErrInvalidJSONContainsPathType])
	// ErrInvalidJSONPathArrayCell means invalid JSON path for an array cell.
	ErrInvalidJSONPathArrayCell = terror.ClassJSON.New(mysql.ErrInvalidJSONPathArrayCell, mysql.MySQLErrName[mysql.ErrInvalidJSONPathArrayCell])
)

func init() {
	typesMySQLErrCodes := map[terror.ErrCode]uint16{
		mysql.ErrInvalidDefault:           mysql.ErrInvalidDefault,
		mysql.ErrDataTooLong:              mysql.ErrDataTooLong,
		mysql.ErrIllegalValueForType:      mysql.ErrIllegalValueForType,
		mysql.WarnDataTruncated:           mysql.WarnDataTruncated,
		mysql.ErrDataOutOfRange:           mysql.ErrDataOutOfRange,
		mysql.ErrDivisionByZero:           mysql.ErrDivisionByZero,
		mysql.ErrTooBigDisplaywidth:       mysql.ErrTooBigDisplaywidth,
		mysql.ErrTooBigFieldlength:        mysql.ErrTooBigFieldlength,
		mysql.ErrTooBigSet:                mysql.ErrTooBigSet,
		mysql.ErrTooBigScale:              mysql.ErrTooBigScale,
		mysql.ErrTooBigPrecision:          mysql.ErrTooBigPrecision,
		mysql.ErrBadNumber:                mysql.ErrBadNumber,
		mysql.ErrInvalidFieldSize:         mysql.ErrInvalidFieldSize,
		mysql.ErrMBiggerThanD:             mysql.ErrMBiggerThanD,
		mysql.ErrWarnDataOutOfRange:       mysql.ErrWarnDataOutOfRange,
		mysql.ErrDuplicatedValueInType:    mysql.ErrDuplicatedValueInType,
		mysql.ErrDatetimeFunctionOverflow: mysql.ErrDatetimeFunctionOverflow,
		mysql.ErrCastAsSignedOverflow:     mysql.ErrCastAsSignedOverflow,
		mysql.ErrCastNegIntAsUnsigned:     mysql.ErrCastNegIntAsUnsigned,
		mysql.ErrInvalidYearFormat:        mysql.ErrInvalidYearFormat,
		mysql.ErrInvalidYear:              mysql.ErrInvalidYear,
		mysql.ErrTruncatedWrongValue:      mysql.ErrTruncatedWrongValue,
		mysql.ErrInvalidTimeFormat:        mysql.ErrInvalidTimeFormat,
		mysql.ErrInvalidWeekModeFormat:    mysql.ErrInvalidWeekModeFormat,
		mysql.ErrWrongValue:               mysql.ErrWrongValue,
	}
	terror.ErrClassToMySQLCodes[terror.ClassTypes] = typesMySQLErrCodes
}
