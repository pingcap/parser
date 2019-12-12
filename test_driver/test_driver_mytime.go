package test_driver

import (
	"fmt"
	gotime "time"

	"github.com/pingcap/errors"
)

// MysqlTime is the internal struct type for Time.
// The order of the attributes is refined to reduce the memory overhead
// considering memory alignment.
type MysqlTime struct {
	// When it's type is Time, HH:MM:SS may be 839:59:59, so use uint32 to avoid overflow.
	hour        uint32 // hour <= 23
	microsecond uint32
	year        uint16 // year <= 9999
	month       uint8  // month <= 12
	day         uint8  // day <= 31
	minute      uint8  // minute <= 59
	second      uint8  // second <= 59
}

// String implements fmt.Stringer.
func (t MysqlTime) String() string {
	return fmt.Sprintf("{%d %d %d %d %d %d %d}", t.year, t.month, t.day, t.hour, t.minute, t.second, t.microsecond)
}

// Year returns the year value.
func (t MysqlTime) Year() int {
	return int(t.year)
}

// Month returns the month value.
func (t MysqlTime) Month() int {
	return int(t.month)
}

// Day returns the day value.
func (t MysqlTime) Day() int {
	return int(t.day)
}

// Hour returns the hour value.
func (t MysqlTime) Hour() int {
	return int(t.hour)
}

// Minute returns the minute value.
func (t MysqlTime) Minute() int {
	return int(t.minute)
}

// Second returns the second value.
func (t MysqlTime) Second() int {
	return int(t.second)
}

// Microsecond returns the microsecond value.
func (t MysqlTime) Microsecond() int {
	return int(t.microsecond)
}

// Weekday returns the Weekday value.
func (t MysqlTime) Weekday() gotime.Weekday {
	// TODO: Consider time_zone variable.
	t1, err := t.GoTime(gotime.Local)
	// allow invalid dates
	if err != nil {
		return t1.Weekday()
	}
	return t1.Weekday()
}

// YearDay returns day in year.
func (t MysqlTime) YearDay() int {
	if t.month == 0 || t.day == 0 {
		return 0
	}
	return calcDaynr(int(t.year), int(t.month), int(t.day)) -
		calcDaynr(int(t.year), 1, 1) + 1
}

// YearWeek return year and week.
func (t MysqlTime) YearWeek(mode int) (int, int) {
	behavior := weekMode(mode) | weekBehaviourYear
	return calcWeek(&t, behavior)
}

// Week returns the week value.
func (t MysqlTime) Week(mode int) int {
	if t.month == 0 || t.day == 0 {
		return 0
	}
	_, week := calcWeek(&t, weekMode(mode))
	return week
}

// GoTime converts MysqlTime to GoTime.
func (t MysqlTime) GoTime(loc *gotime.Location) (gotime.Time, error) {
	// gotime.Time can't represent month 0 or day 0, date contains 0 would be converted to a nearest date,
	// For example, 2006-12-00 00:00:00 would become 2015-11-30 23:59:59.
	tm := gotime.Date(t.Year(), gotime.Month(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Microsecond()*1000, loc)
	year, month, day := tm.Date()
	hour, minute, second := tm.Clock()
	microsec := tm.Nanosecond() / 1000
	// This function will check the result, and return an error if it's not the same with the origin input.
	if year != t.Year() || int(month) != t.Month() || day != t.Day() ||
		hour != t.Hour() || minute != t.Minute() || second != t.Second() ||
		microsec != t.Microsecond() {
		return tm, errors.Trace(ErrWrongValue.GenWithStackByArgs(TimeStr, t))
	}
	return tm, nil
}

// IsLeapYear returns if it's leap year.
func (t MysqlTime) IsLeapYear() bool {
	return isLeapYear(t.year)
}

func isLeapYear(year uint16) bool {
	return (year%4 == 0 && year%100 != 0) || year%400 == 0
}

func calcTimeFromSec(to *MysqlTime, seconds, microseconds int) {
	to.hour = uint32(seconds / 3600)
	seconds = seconds % 3600
	to.minute = uint8(seconds / 60)
	to.second = uint8(seconds % 60)
	to.microsecond = uint32(microseconds)
}

const secondsIn24Hour = 86400

// calcTimeDiff calculates difference between two datetime values as seconds + microseconds.
// t1 and t2 should be TIME/DATE/DATETIME value.
// sign can be +1 or -1, and t2 is preprocessed with sign first.
func calcTimeDiff(t1, t2 MysqlTime, sign int) (seconds, microseconds int, neg bool) {
	days := calcDaynr(t1.Year(), t1.Month(), t1.Day())
	days2 := calcDaynr(t2.Year(), t2.Month(), t2.Day())
	days -= sign * days2

	tmp := (int64(days)*secondsIn24Hour+
		int64(t1.Hour())*3600+int64(t1.Minute())*60+
		int64(t1.Second())-
		int64(sign)*(int64(t2.Hour())*3600+int64(t2.Minute())*60+
			int64(t2.Second())))*
		1e6 +
		int64(t1.Microsecond()) - int64(sign)*int64(t2.Microsecond())

	if tmp < 0 {
		tmp = -tmp
		neg = true
	}
	seconds = int(tmp / 1e6)
	microseconds = int(tmp % 1e6)
	return
}

// datetimeToUint64 converts time value to integer in YYYYMMDDHHMMSS format.
func datetimeToUint64(t MysqlTime) uint64 {
	return dateToUint64(t)*1e6 + timeToUint64(t)
}

// dateToUint64 converts time value to integer in YYYYMMDD format.
func dateToUint64(t MysqlTime) uint64 {
	return uint64(t.Year())*10000 +
		uint64(t.Month())*100 +
		uint64(t.Day())
}

// timeToUint64 converts time value to integer in HHMMSS format.
func timeToUint64(t MysqlTime) uint64 {
	return uint64(t.Hour())*10000 +
		uint64(t.Minute())*100 +
		uint64(t.Second())
}

// calcDaynr calculates days since 0000-00-00.
func calcDaynr(year, month, day int) int {
	if year == 0 && month == 0 {
		return 0
	}

	delsum := 365*year + 31*(month-1) + day
	if month <= 2 {
		year--
	} else {
		delsum -= (month*4 + 23) / 10
	}
	temp := ((year/100 + 1) * 3) / 4
	return delsum + year/4 - temp
}

// calcDaysInYear calculates days in one year, it works with 0 <= year <= 99.
func calcDaysInYear(year int) int {
	if (year&3) == 0 && (year%100 != 0 || (year%400 == 0 && (year != 0))) {
		return 366
	}
	return 365
}

// calcWeekday calculates weekday from daynr, returns 0 for Monday, 1 for Tuesday ...
func calcWeekday(daynr int, sundayFirstDayOfWeek bool) int {
	daynr += 5
	if sundayFirstDayOfWeek {
		daynr++
	}
	return daynr % 7
}

type weekBehaviour uint

const (
	// weekBehaviourMondayFirst set Monday as first day of week; otherwise Sunday is first day of week
	weekBehaviourMondayFirst weekBehaviour = 1 << iota
	// If set, Week is in range 1-53, otherwise Week is in range 0-53.
	// Note that this flag is only relevant if WEEK_JANUARY is not set.
	weekBehaviourYear
	// If not set, Weeks are numbered according to ISO 8601:1988.
	// If set, the week that contains the first 'first-day-of-week' is week 1.
	weekBehaviourFirstWeekday
)

func (v weekBehaviour) test(flag weekBehaviour) bool {
	return (v & flag) != 0
}

func weekMode(mode int) weekBehaviour {
	weekFormat := weekBehaviour(mode & 7)
	if (weekFormat & weekBehaviourMondayFirst) == 0 {
		weekFormat ^= weekBehaviourFirstWeekday
	}
	return weekFormat
}

// calcWeek calculates week and year for the time.
func calcWeek(t *MysqlTime, wb weekBehaviour) (year int, week int) {
	var days int
	daynr := calcDaynr(int(t.year), int(t.month), int(t.day))
	firstDaynr := calcDaynr(int(t.year), 1, 1)
	mondayFirst := wb.test(weekBehaviourMondayFirst)
	weekYear := wb.test(weekBehaviourYear)
	firstWeekday := wb.test(weekBehaviourFirstWeekday)

	weekday := calcWeekday(firstDaynr, !mondayFirst)

	year = int(t.year)

	if t.month == 1 && int(t.day) <= 7-weekday {
		if !weekYear &&
			((firstWeekday && weekday != 0) || (!firstWeekday && weekday >= 4)) {
			week = 0
			return
		}
		weekYear = true
		year--
		days = calcDaysInYear(year)
		firstDaynr -= days
		weekday = (weekday + 53*7 - days) % 7
	}

	if (firstWeekday && weekday != 0) ||
		(!firstWeekday && weekday >= 4) {
		days = daynr - (firstDaynr + 7 - weekday)
	} else {
		days = daynr - (firstDaynr - weekday)
	}

	if weekYear && days >= 52*7 {
		weekday = (weekday + calcDaysInYear(year)) % 7
		if (!firstWeekday && weekday < 4) ||
			(firstWeekday && weekday == 0) {
			year++
			week = 1
			return
		}
	}
	week = days/7 + 1
	return
}
