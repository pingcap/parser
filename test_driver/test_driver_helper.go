package test_driver

import (
	"math"
	"strings"
	"unicode"

	"github.com/pingcap/errors"
)

func isSpace(c byte) bool {
	return c == ' ' || c == '\t'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func myMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func myMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

const (
	maxUint    = uint64(math.MaxUint64)
	uintCutOff = maxUint/uint64(10) + 1
	intCutOff  = uint64(math.MaxInt64) + 1
)

// strToInt converts a string to an integer in best effort.
func strToInt(str string) (int64, error) {
	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return 0, ErrTruncated
	}
	negative := false
	i := 0
	if str[i] == '-' {
		negative = true
		i++
	} else if str[i] == '+' {
		i++
	}

	var (
		err    error
		hasNum = false
	)
	r := uint64(0)
	for ; i < len(str); i++ {
		if !unicode.IsDigit(rune(str[i])) {
			err = ErrTruncated
			break
		}
		hasNum = true
		if r >= uintCutOff {
			r = 0
			err = errors.Trace(ErrBadNumber)
			break
		}
		r = r * uint64(10)

		r1 := r + uint64(str[i]-'0')
		if r1 < r || r1 > maxUint {
			r = 0
			err = errors.Trace(ErrBadNumber)
			break
		}
		r = r1
	}
	if !hasNum {
		err = ErrTruncated
	}

	if !negative && r >= intCutOff {
		return math.MaxInt64, errors.Trace(ErrBadNumber)
	}

	if negative && r > intCutOff {
		return math.MinInt64, errors.Trace(ErrBadNumber)
	}

	if negative {
		r = -r
	}
	return int64(r), err
}

func Abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

// uintSizeTable is used as a table to do comparison to get uint length is faster than doing loop on division with 10
var uintSizeTable = [21]uint64{
	0, // redundant 0 here, so to make function StrLenOfUint64Fast to count from 1 and return i directly
	9, 99, 999, 9999, 99999,
	999999, 9999999, 99999999, 999999999, 9999999999,
	99999999999, 999999999999, 9999999999999, 99999999999999, 999999999999999,
	9999999999999999, 99999999999999999, 999999999999999999, 9999999999999999999,
	math.MaxUint64,
} // math.MaxUint64 is 18446744073709551615 and it has 20 digits

// StrLenOfUint64Fast efficiently calculate the string character lengths of an uint64 as input
func StrLenOfUint64Fast(x uint64) int {
	for i := 1; ; i++ {
		if x <= uintSizeTable[i] {
			return i
		}
	}
}

// StrLenOfInt64Fast efficiently calculate the string character lengths of an int64 as input
func StrLenOfInt64Fast(x int64) int {
	size := 0
	if x < 0 {
		size = 1 // add "-" sign on the length count
	}
	return size + StrLenOfUint64Fast(uint64(Abs(x)))
}
