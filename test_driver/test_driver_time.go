package test_driver

import (
	"bytes"
	"fmt"
	"github.com/pingcap/errors"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/terror"
	"math"
	"regexp"
	"strconv"
	gotime "time"
)

// Time format without fractional seconds precision.
const (
	DateFormat = "2006-01-02"
	TimeFormat = "2006-01-02 15:04:05"
	// TimeFSPFormat is time format with fractional seconds precision.
	TimeFSPFormat = "2006-01-02 15:04:05.000000"
)

const (
	// MinYear is the minimum for mysql year type.
	MinYear int16 = 1901
	// MaxYear is the maximum for mysql year type.
	MaxYear int16 = 2155
	// MaxDuration is the maximum for duration.
	MaxDuration int64 = 838*10000 + 59*100 + 59
	// MinTime is the minimum for mysql time type.
	MinTime = -gotime.Duration(838*3600+59*60+59) * gotime.Second
	// MaxTime is the maximum for mysql time type.
	MaxTime = gotime.Duration(838*3600+59*60+59) * gotime.Second
	// ZeroDatetimeStr is the string representation of a zero datetime.
	ZeroDatetimeStr = "0000-00-00 00:00:00"
	// ZeroDateStr is the string representation of a zero date.
	ZeroDateStr = "0000-00-00"

	// TimeMaxHour is the max hour for mysql time type.
	TimeMaxHour = 838
	// TimeMaxMinute is the max minute for mysql time type.
	TimeMaxMinute = 59
	// TimeMaxSecond is the max second for mysql time type.
	TimeMaxSecond = 59
	// TimeMaxValue is the maximum value for mysql time type.
	TimeMaxValue = TimeMaxHour*10000 + TimeMaxMinute*100 + TimeMaxSecond
	// TimeMaxValueSeconds is the maximum second value for mysql time type.
	TimeMaxValueSeconds = TimeMaxHour*3600 + TimeMaxMinute*60 + TimeMaxSecond
)

const (
	// YearIndex is index of 'YEARS-MONTHS DAYS HOURS:MINUTES:SECONDS.MICROSECONDS' expr Format
	YearIndex = 0 + iota
	// MonthIndex is index of 'YEARS-MONTHS DAYS HOURS:MINUTES:SECONDS.MICROSECONDS' expr Format
	MonthIndex
	// DayIndex is index of 'YEARS-MONTHS DAYS HOURS:MINUTES:SECONDS.MICROSECONDS' expr Format
	DayIndex
	// HourIndex is index of 'YEARS-MONTHS DAYS HOURS:MINUTES:SECONDS.MICROSECONDS' expr Format
	HourIndex
	// MinuteIndex is index of 'YEARS-MONTHS DAYS HOURS:MINUTES:SECONDS.MICROSECONDS' expr Format
	MinuteIndex
	// SecondIndex is index of 'YEARS-MONTHS DAYS HOURS:MINUTES:SECONDS.MICROSECONDS' expr Format
	SecondIndex
	// MicrosecondIndex is index of 'YEARS-MONTHS DAYS HOURS:MINUTES:SECONDS.MICROSECONDS' expr Format
	MicrosecondIndex
)

const (
	// YearMonthMaxCnt is max parameters count 'YEARS-MONTHS' expr Format allowed
	YearMonthMaxCnt = 2
	// DayHourMaxCnt is max parameters count 'DAYS HOURS' expr Format allowed
	DayHourMaxCnt = 2
	// DayMinuteMaxCnt is max parameters count 'DAYS HOURS:MINUTES' expr Format allowed
	DayMinuteMaxCnt = 3
	// DaySecondMaxCnt is max parameters count 'DAYS HOURS:MINUTES:SECONDS' expr Format allowed
	DaySecondMaxCnt = 4
	// DayMicrosecondMaxCnt is max parameters count 'DAYS HOURS:MINUTES:SECONDS.MICROSECONDS' expr Format allowed
	DayMicrosecondMaxCnt = 5
	// HourMinuteMaxCnt is max parameters count 'HOURS:MINUTES' expr Format allowed
	HourMinuteMaxCnt = 2
	// HourSecondMaxCnt is max parameters count 'HOURS:MINUTES:SECONDS' expr Format allowed
	HourSecondMaxCnt = 3
	// HourMicrosecondMaxCnt is max parameters count 'HOURS:MINUTES:SECONDS.MICROSECONDS' expr Format allowed
	HourMicrosecondMaxCnt = 4
	// MinuteSecondMaxCnt is max parameters count 'MINUTES:SECONDS' expr Format allowed
	MinuteSecondMaxCnt = 2
	// MinuteMicrosecondMaxCnt is max parameters count 'MINUTES:SECONDS.MICROSECONDS' expr Format allowed
	MinuteMicrosecondMaxCnt = 3
	// SecondMicrosecondMaxCnt is max parameters count 'SECONDS.MICROSECONDS' expr Format allowed
	SecondMicrosecondMaxCnt = 2
	// TimeValueCnt is parameters count 'YEARS-MONTHS DAYS HOURS:MINUTES:SECONDS.MICROSECONDS' expr Format
	TimeValueCnt = 7
)

// Zero values for different types.
var (
	// ZeroDuration is the zero value for Duration type.
	ZeroDuration = Duration{Duration: gotime.Duration(0), Fsp: DefaultFsp}

	// ZeroTime is the zero value for TimeInternal type.
	ZeroTime = MysqlTime{}

	// ZeroDatetime is the zero value for datetime Time.
	ZeroDatetime = Time{
		Time: ZeroTime,
		Type: mysql.TypeDatetime,
		Fsp:  DefaultFsp,
	}

	// ZeroTimestamp is the zero value for timestamp Time.
	ZeroTimestamp = Time{
		Time: ZeroTime,
		Type: mysql.TypeTimestamp,
		Fsp:  DefaultFsp,
	}

	// ZeroDate is the zero value for date Time.
	ZeroDate = Time{
		Time: ZeroTime,
		Type: mysql.TypeDate,
		Fsp:  DefaultFsp,
	}
)

var (
	// MonthNames lists names of months, which are used in builtin time function `monthname`.
	MonthNames = []string{
		"January", "February",
		"March", "April",
		"May", "June",
		"July", "August",
		"September", "October",
		"November", "December",
	}
)

const (
	// GoDurationDay is the gotime.Duration which equals to a Day.
	GoDurationDay = gotime.Hour * 24
	// GoDurationWeek is the gotime.Duration which equals to a Week.
	GoDurationWeek = GoDurationDay * 7
)

// FromDate makes a internal time representation from the given date.
func FromDate(year int, month int, day int, hour int, minute int, second int, microsecond int) MysqlTime {
	return MysqlTime{
		year:        uint16(year),
		month:       uint8(month),
		day:         uint8(day),
		hour:        uint32(hour),
		minute:      uint8(minute),
		second:      uint8(second),
		microsecond: uint32(microsecond),
	}
}

// Clock returns the hour, minute, and second within the day specified by t.
func (t Time) Clock() (hour int, minute int, second int) {
	return t.Time.Hour(), t.Time.Minute(), t.Time.Second()
}

// Time is the struct for handling datetime, timestamp and date.
// TODO: check if need a NewTime function to set Fsp default value?
type Time struct {
	Time MysqlTime
	Type uint8
	// Fsp is short for Fractional Seconds Precision.
	// See http://dev.mysql.com/doc/refman/5.7/en/fractional-seconds.html
	Fsp int8
}

func (t Time) String() string {
	if t.Type == mysql.TypeDate {
		// We control the format, so no error would occur.
		str, err := t.DateFormat("%Y-%m-%d")
		terror.Log(errors.Trace(err))
		return str
	}

	str, err := t.DateFormat("%Y-%m-%d %H:%i:%s")
	terror.Log(errors.Trace(err))
	if t.Fsp > 0 {
		tmp := fmt.Sprintf(".%06d", t.Time.Microsecond())
		str = str + tmp[:1+t.Fsp]
	}

	return str
}

// IsZero returns a boolean indicating whether the time is equal to ZeroTime.
func (t Time) IsZero() bool {
	return compareTime(t.Time, ZeroTime) == 0
}

// compareTime compare two MysqlTime.
// return:
//  0: if a == b
//  1: if a > b
// -1: if a < b
func compareTime(a, b MysqlTime) int {
	ta := datetimeToUint64(a)
	tb := datetimeToUint64(b)

	switch {
	case ta < tb:
		return -1
	case ta > tb:
		return 1
	}

	switch {
	case a.Microsecond() < b.Microsecond():
		return -1
	case a.Microsecond() > b.Microsecond():
		return 1
	}

	return 0
}

// Duration is the type for MySQL TIME type.
type Duration struct {
	gotime.Duration
	// Fsp is short for Fractional Seconds Precision.
	// See http://dev.mysql.com/doc/refman/5.7/en/fractional-seconds.html
	Fsp int8
}

// String returns the time formatted using default TimeFormat and fsp.
func (d Duration) String() string {
	var buf bytes.Buffer

	sign, hours, minutes, seconds, fraction := splitDuration(d.Duration)
	if sign < 0 {
		buf.WriteByte('-')
	}

	fmt.Fprintf(&buf, "%02d:%02d:%02d", hours, minutes, seconds)
	if d.Fsp > 0 {
		buf.WriteString(".")
		buf.WriteString(d.formatFrac(fraction))
	}

	p := buf.String()

	return p
}

func (d Duration) formatFrac(frac int) string {
	s := fmt.Sprintf("%06d", frac)
	return s[0:d.Fsp]
}

func splitDuration(t gotime.Duration) (int, int, int, int, int) {
	sign := 1
	if t < 0 {
		t = -t
		sign = -1
	}

	hours := t / gotime.Hour
	t -= hours * gotime.Hour
	minutes := t / gotime.Minute
	t -= minutes * gotime.Minute
	seconds := t / gotime.Second
	t -= seconds * gotime.Second
	fraction := t / gotime.Microsecond

	return sign, int(hours), int(minutes), int(seconds), int(fraction)
}

// DateFormat returns a textual representation of the time value formatted
// according to layout.
// See http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_date-format
func (t Time) DateFormat(layout string) (string, error) {
	var buf bytes.Buffer
	inPatternMatch := false
	for _, b := range layout {
		if inPatternMatch {
			if err := t.convertDateFormat(b, &buf); err != nil {
				return "", errors.Trace(err)
			}
			inPatternMatch = false
			continue
		}

		// It's not in pattern match now.
		if b == '%' {
			inPatternMatch = true
		} else {
			buf.WriteRune(b)
		}
	}
	return buf.String(), nil
}

var abbrevWeekdayName = []string{
	"Sun", "Mon", "Tue",
	"Wed", "Thu", "Fri", "Sat",
}

func (t Time) convertDateFormat(b rune, buf *bytes.Buffer) error {
	switch b {
	case 'b':
		m := t.Time.Month()
		if m == 0 || m > 12 {
			return errors.Trace(ErrWrongValue.GenWithStackByArgs(TimeStr, m))
		}
		buf.WriteString(MonthNames[m-1][:3])
	case 'M':
		m := t.Time.Month()
		if m == 0 || m > 12 {
			return errors.Trace(ErrWrongValue.GenWithStackByArgs(TimeStr, m))
		}
		buf.WriteString(MonthNames[m-1])
	case 'm':
		buf.WriteString(FormatIntWidthN(t.Time.Month(), 2))
	case 'c':
		buf.WriteString(strconv.FormatInt(int64(t.Time.Month()), 10))
	case 'D':
		buf.WriteString(strconv.FormatInt(int64(t.Time.Day()), 10))
		buf.WriteString(abbrDayOfMonth(t.Time.Day()))
	case 'd':
		buf.WriteString(FormatIntWidthN(t.Time.Day(), 2))
	case 'e':
		buf.WriteString(strconv.FormatInt(int64(t.Time.Day()), 10))
	case 'j':
		fmt.Fprintf(buf, "%03d", t.Time.YearDay())
	case 'H':
		buf.WriteString(FormatIntWidthN(t.Time.Hour(), 2))
	case 'k':
		buf.WriteString(strconv.FormatInt(int64(t.Time.Hour()), 10))
	case 'h', 'I':
		t := t.Time.Hour()
		if t%12 == 0 {
			buf.WriteString("12")
		} else {
			buf.WriteString(FormatIntWidthN(t%12, 2))
		}
	case 'l':
		t := t.Time.Hour()
		if t%12 == 0 {
			buf.WriteString("12")
		} else {
			buf.WriteString(strconv.FormatInt(int64(t%12), 10))
		}
	case 'i':
		buf.WriteString(FormatIntWidthN(t.Time.Minute(), 2))
	case 'p':
		hour := t.Time.Hour()
		if hour/12%2 == 0 {
			buf.WriteString("AM")
		} else {
			buf.WriteString("PM")
		}
	case 'r':
		h := t.Time.Hour()
		h %= 24
		switch {
		case h == 0:
			fmt.Fprintf(buf, "%02d:%02d:%02d AM", 12, t.Time.Minute(), t.Time.Second())
		case h == 12:
			fmt.Fprintf(buf, "%02d:%02d:%02d PM", 12, t.Time.Minute(), t.Time.Second())
		case h < 12:
			fmt.Fprintf(buf, "%02d:%02d:%02d AM", h, t.Time.Minute(), t.Time.Second())
		default:
			fmt.Fprintf(buf, "%02d:%02d:%02d PM", h-12, t.Time.Minute(), t.Time.Second())
		}
	case 'T':
		fmt.Fprintf(buf, "%02d:%02d:%02d", t.Time.Hour(), t.Time.Minute(), t.Time.Second())
	case 'S', 's':
		buf.WriteString(FormatIntWidthN(t.Time.Second(), 2))
	case 'f':
		fmt.Fprintf(buf, "%06d", t.Time.Microsecond())
	case 'U':
		w := t.Time.Week(0)
		buf.WriteString(FormatIntWidthN(w, 2))
	case 'u':
		w := t.Time.Week(1)
		buf.WriteString(FormatIntWidthN(w, 2))
	case 'V':
		w := t.Time.Week(2)
		buf.WriteString(FormatIntWidthN(w, 2))
	case 'v':
		_, w := t.Time.YearWeek(3)
		buf.WriteString(FormatIntWidthN(w, 2))
	case 'a':
		weekday := t.Time.Weekday()
		buf.WriteString(abbrevWeekdayName[weekday])
	case 'W':
		buf.WriteString(t.Time.Weekday().String())
	case 'w':
		buf.WriteString(strconv.FormatInt(int64(t.Time.Weekday()), 10))
	case 'X':
		year, _ := t.Time.YearWeek(2)
		if year < 0 {
			buf.WriteString(strconv.FormatUint(uint64(math.MaxUint32), 10))
		} else {
			buf.WriteString(FormatIntWidthN(year, 4))
		}
	case 'x':
		year, _ := t.Time.YearWeek(3)
		if year < 0 {
			buf.WriteString(strconv.FormatUint(uint64(math.MaxUint32), 10))
		} else {
			buf.WriteString(FormatIntWidthN(year, 4))
		}
	case 'Y':
		buf.WriteString(FormatIntWidthN(t.Time.Year(), 4))
	case 'y':
		str := FormatIntWidthN(t.Time.Year(), 4)
		buf.WriteString(str[2:])
	default:
		buf.WriteRune(b)
	}

	return nil
}

// FormatIntWidthN uses to format int with width. Insufficient digits are filled by 0.
func FormatIntWidthN(num, n int) string {
	numString := strconv.FormatInt(int64(num), 10)
	if len(numString) >= n {
		return numString
	}
	padBytes := make([]byte, n-len(numString))
	for i := range padBytes {
		padBytes[i] = '0'
	}
	return string(padBytes) + numString
}

func abbrDayOfMonth(day int) string {
	var str string
	switch day {
	case 1, 21, 31:
		str = "st"
	case 2, 22:
		str = "nd"
	case 3, 23:
		str = "rd"
	default:
		str = "th"
	}
	return str
}

// numericRegex: it was for any numeric characters
var numericRegex = regexp.MustCompile("[0-9]+")
