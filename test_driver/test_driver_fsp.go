package test_driver

import (
	"strings"

	"github.com/pingcap/errors"
)

const (
	// UnspecifiedFsp is the unspecified fractional seconds part.
	UnspecifiedFsp = int8(-1)
	// MaxFsp is the maximum digit of fractional seconds part.
	MaxFsp = int8(6)
	// MinFsp is the minimum digit of fractional seconds part.
	MinFsp = int8(0)
	// DefaultFsp is the default digit of fractional seconds part.
	// MySQL use 0 as the default Fsp.
	DefaultFsp = int8(0)
)

// CheckFsp checks whether fsp is in valid range.
func CheckFsp(fsp int) (int8, error) {
	if fsp == int(UnspecifiedFsp) {
		return DefaultFsp, nil
	}
	if fsp < int(MinFsp) || fsp > int(MaxFsp) {
		return DefaultFsp, errors.Errorf("Invalid fsp %d", fsp)
	}
	return int8(fsp), nil
}

// alignFrac is used to generate alignment frac, like `100` -> `100000` ,`-100` -> `-100000`
func alignFrac(s string, fsp int) string {
	sl := len(s)
	if sl > 0 && s[0] == '-' {
		sl = sl - 1
	}
	if sl < fsp {
		return s + strings.Repeat("0", fsp-sl)
	}

	return s
}
