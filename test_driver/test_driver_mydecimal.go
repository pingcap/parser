package test_driver

import (
	"github.com/pingcap/errors"
	"github.com/pingcap/parser/terror"
	"math"
)

// RoundMode is the type for round mode.
type RoundMode int32

// constant values.
const (
	ten0 = 1
	ten1 = 10
	ten2 = 100
	ten3 = 1000
	ten4 = 10000
	ten5 = 100000
	ten6 = 1000000
	ten7 = 10000000
	ten8 = 100000000
	ten9 = 1000000000

	maxWordBufLen = 9 // A MyDecimal holds 9 words.
	digitsPerWord = 9 // A word holds 9 digits.
	wordSize      = 4 // A word is 4 bytes int32.
	digMask       = ten8
	wordBase      = ten9
	wordMax       = wordBase - 1

	// ModeHalfEven rounds normally.
	ModeHalfEven RoundMode = 5
	// Truncate just truncates the decimal.
	ModeTruncate RoundMode = 10
	// Ceiling is not supported now.
	modeCeiling RoundMode = 0
)

var (
	wordBufLen = 9
	mod9       = [128]int8{
		0, 1, 2, 3, 4, 5, 6, 7, 8,
		0, 1, 2, 3, 4, 5, 6, 7, 8,
		0, 1, 2, 3, 4, 5, 6, 7, 8,
		0, 1, 2, 3, 4, 5, 6, 7, 8,
		0, 1, 2, 3, 4, 5, 6, 7, 8,
		0, 1, 2, 3, 4, 5, 6, 7, 8,
		0, 1, 2, 3, 4, 5, 6, 7, 8,
		0, 1, 2, 3, 4, 5, 6, 7, 8,
		0, 1, 2, 3, 4, 5, 6, 7, 8,
		0, 1, 2, 3, 4, 5, 6, 7, 8,
		0, 1, 2, 3, 4, 5, 6, 7, 8,
		0, 1, 2, 3, 4, 5, 6, 7, 8,
		0, 1, 2, 3, 4, 5, 6, 7, 8,
		0, 1, 2, 3, 4, 5, 6, 7, 8,
		0, 1,
	}
	div9 = [128]int{
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 1, 1, 1, 1,
		2, 2, 2, 2, 2, 2, 2, 2, 2,
		3, 3, 3, 3, 3, 3, 3, 3, 3,
		4, 4, 4, 4, 4, 4, 4, 4, 4,
		5, 5, 5, 5, 5, 5, 5, 5, 5,
		6, 6, 6, 6, 6, 6, 6, 6, 6,
		7, 7, 7, 7, 7, 7, 7, 7, 7,
		8, 8, 8, 8, 8, 8, 8, 8, 8,
		9, 9, 9, 9, 9, 9, 9, 9, 9,
		10, 10, 10, 10, 10, 10, 10, 10, 10,
		11, 11, 11, 11, 11, 11, 11, 11, 11,
		12, 12, 12, 12, 12, 12, 12, 12, 12,
		13, 13, 13, 13, 13, 13, 13, 13, 13,
		14, 14,
	}
	powers10  = [10]int32{ten0, ten1, ten2, ten3, ten4, ten5, ten6, ten7, ten8, ten9}
	dig2bytes = [10]int{0, 1, 1, 2, 2, 3, 3, 4, 4, 4}
	fracMax   = [8]int32{
		900000000,
		990000000,
		999000000,
		999900000,
		999990000,
		999999000,
		999999900,
		999999990,
	}
	zeroMyDecimal = MyDecimal{}
)

// get the zero of MyDecimal with the specified result fraction digits
func zeroMyDecimalWithFrac(frac int8) MyDecimal {
	zero := MyDecimal{}
	zero.digitsFrac = frac
	zero.resultFrac = frac
	return zero
}

// add adds a and b and carry, returns the sum and new carry.
func add(a, b, carry int32) (int32, int32) {
	sum := a + b + carry
	if sum >= wordBase {
		carry = 1
		sum -= wordBase
	} else {
		carry = 0
	}
	return sum, carry
}

// sub subtracts b and carry from a, returns the diff and new carry.
func sub(a, b, carry int32) (int32, int32) {
	diff := a - b - carry
	if diff < 0 {
		carry = 1
		diff += wordBase
	} else {
		carry = 0
	}
	return diff, carry
}

// fixWordCntError limits word count in wordBufLen, and returns overflow or truncate error.
func fixWordCntError(wordsInt, wordsFrac int) (newWordsInt int, newWordsFrac int, err error) {
	if wordsInt+wordsFrac > wordBufLen {
		if wordsInt > wordBufLen {
			return wordBufLen, 0, ErrOverflow
		}
		return wordsInt, wordBufLen - wordsInt, ErrTruncated
	}
	return wordsInt, wordsFrac, nil
}

/*
  countLeadingZeroes returns the number of leading zeroes that can be removed from fraction.

  @param   i    start index
  @param   word value to compare against list of powers of 10
*/
func countLeadingZeroes(i int, word int32) int {
	leading := 0
	for word < powers10[i] {
		i--
		leading++
	}
	return leading
}

/*
  countTrailingZeros returns the number of trailing zeroes that can be removed from fraction.

  @param   i    start index
  @param   word  value to compare against list of powers of 10
*/
func countTrailingZeroes(i int, word int32) int {
	trailing := 0
	for word%powers10[i] == 0 {
		i++
		trailing++
	}
	return trailing
}

func digitsToWords(digits int) int {
	if digits+digitsPerWord-1 >= 0 && digits+digitsPerWord-1 < 128 {
		return div9[digits+digitsPerWord-1]
	}
	return (digits + digitsPerWord - 1) / digitsPerWord
}

// MyDecimal represents a decimal value.
type MyDecimal struct {
	digitsInt int8 // the number of *decimal* digits before the point.

	digitsFrac int8 // the number of decimal digits after the point.

	resultFrac int8 // result fraction digits.

	negative bool

	//  wordBuf is an array of int32 words.
	// A word is an int32 value can hold 9 digits.(0 <= word < wordBase)
	wordBuf [maxWordBufLen]int32
}

// String returns the decimal string representation rounded to resultFrac.
func (d *MyDecimal) String() string {
	tmp := *d
	err := tmp.Round(&tmp, int(tmp.resultFrac), ModeHalfEven)
	terror.Log(errors.Trace(err))
	return string(tmp.ToString())
}

func (d *MyDecimal) stringSize() int {
	// sign, zero integer and dot.
	return int(d.digitsInt + d.digitsFrac + 3)
}

func (d *MyDecimal) removeLeadingZeros() (wordIdx int, digitsInt int) {
	digitsInt = int(d.digitsInt)
	i := ((digitsInt - 1) % digitsPerWord) + 1
	for digitsInt > 0 && d.wordBuf[wordIdx] == 0 {
		digitsInt -= i
		i = digitsPerWord
		wordIdx++
	}
	if digitsInt > 0 {
		digitsInt -= countLeadingZeroes((digitsInt-1)%digitsPerWord, d.wordBuf[wordIdx])
	} else {
		digitsInt = 0
	}
	return
}

// ToString converts decimal to its printable string representation without rounding.
//
//  RETURN VALUE
//
//      str       - result string
//      errCode   - eDecOK/eDecTruncate/eDecOverflow
//
func (d *MyDecimal) ToString() (str []byte) {
	str = make([]byte, d.stringSize())
	digitsFrac := int(d.digitsFrac)
	wordStartIdx, digitsInt := d.removeLeadingZeros()
	if digitsInt+digitsFrac == 0 {
		digitsInt = 1
		wordStartIdx = 0
	}

	digitsIntLen := digitsInt
	if digitsIntLen == 0 {
		digitsIntLen = 1
	}
	digitsFracLen := digitsFrac
	length := digitsIntLen + digitsFracLen
	if d.negative {
		length++
	}
	if digitsFrac > 0 {
		length++
	}
	str = str[:length]
	strIdx := 0
	if d.negative {
		str[strIdx] = '-'
		strIdx++
	}
	var fill int
	if digitsFrac > 0 {
		fracIdx := strIdx + digitsIntLen
		fill = digitsFracLen - digitsFrac
		wordIdx := wordStartIdx + digitsToWords(digitsInt)
		str[fracIdx] = '.'
		fracIdx++
		for ; digitsFrac > 0; digitsFrac -= digitsPerWord {
			x := d.wordBuf[wordIdx]
			wordIdx++
			for i := myMin(digitsFrac, digitsPerWord); i > 0; i-- {
				y := x / digMask
				str[fracIdx] = byte(y) + '0'
				fracIdx++
				x -= y * digMask
				x *= 10
			}
		}
		for ; fill > 0; fill-- {
			str[fracIdx] = '0'
			fracIdx++
		}
	}
	fill = digitsIntLen - digitsInt
	if digitsInt == 0 {
		fill-- /* symbol 0 before digital point */
	}
	for ; fill > 0; fill-- {
		str[strIdx] = '0'
		strIdx++
	}
	if digitsInt > 0 {
		strIdx += digitsInt
		wordIdx := wordStartIdx + digitsToWords(digitsInt)
		for ; digitsInt > 0; digitsInt -= digitsPerWord {
			wordIdx--
			x := d.wordBuf[wordIdx]
			for i := myMin(digitsInt, digitsPerWord); i > 0; i-- {
				y := x / 10
				strIdx--
				str[strIdx] = '0' + byte(x-y*10)
				x = y
			}
		}
	} else {
		str[strIdx] = '0'
	}
	return
}

// FromString parses decimal from string.
func (d *MyDecimal) FromString(str []byte) error {
	for i := 0; i < len(str); i++ {
		if !isSpace(str[i]) {
			str = str[i:]
			break
		}
	}
	if len(str) == 0 {
		*d = zeroMyDecimal
		return ErrBadNumber
	}
	switch str[0] {
	case '-':
		d.negative = true
		fallthrough
	case '+':
		str = str[1:]
	}
	var strIdx int
	for strIdx < len(str) && isDigit(str[strIdx]) {
		strIdx++
	}
	digitsInt := strIdx
	var digitsFrac int
	var endIdx int
	if strIdx < len(str) && str[strIdx] == '.' {
		endIdx = strIdx + 1
		for endIdx < len(str) && isDigit(str[endIdx]) {
			endIdx++
		}
		digitsFrac = endIdx - strIdx - 1
	} else {
		digitsFrac = 0
		endIdx = strIdx
	}
	if digitsInt+digitsFrac == 0 {
		*d = zeroMyDecimal
		return ErrBadNumber
	}
	wordsInt := digitsToWords(digitsInt)
	wordsFrac := digitsToWords(digitsFrac)
	wordsInt, wordsFrac, err := fixWordCntError(wordsInt, wordsFrac)
	if err != nil {
		digitsFrac = wordsFrac * digitsPerWord
		if err == ErrOverflow {
			digitsInt = wordsInt * digitsPerWord
		}
	}
	d.digitsInt = int8(digitsInt)
	d.digitsFrac = int8(digitsFrac)
	wordIdx := wordsInt
	strIdxTmp := strIdx
	var word int32
	var innerIdx int
	for digitsInt > 0 {
		digitsInt--
		strIdx--
		word += int32(str[strIdx]-'0') * powers10[innerIdx]
		innerIdx++
		if innerIdx == digitsPerWord {
			wordIdx--
			d.wordBuf[wordIdx] = word
			word = 0
			innerIdx = 0
		}
	}
	if innerIdx != 0 {
		wordIdx--
		d.wordBuf[wordIdx] = word
	}

	wordIdx = wordsInt
	strIdx = strIdxTmp
	word = 0
	innerIdx = 0
	for digitsFrac > 0 {
		digitsFrac--
		strIdx++
		word = int32(str[strIdx]-'0') + word*10
		innerIdx++
		if innerIdx == digitsPerWord {
			d.wordBuf[wordIdx] = word
			wordIdx++
			word = 0
			innerIdx = 0
		}
	}
	if innerIdx != 0 {
		d.wordBuf[wordIdx] = word * powers10[digitsPerWord-innerIdx]
	}
	if endIdx+1 <= len(str) && (str[endIdx] == 'e' || str[endIdx] == 'E') {
		exponent, err1 := strToInt(string(str[endIdx+1:]))
		if err1 != nil {
			err = errors.Cause(err1)
			if err != ErrTruncated {
				*d = zeroMyDecimal
			}
		}
		if exponent > math.MaxInt32/2 {
			negative := d.negative
			maxDecimal(wordBufLen*digitsPerWord, 0, d)
			d.negative = negative
			err = ErrOverflow
		}
		if exponent < math.MinInt32/2 && err != ErrOverflow {
			*d = zeroMyDecimal
			err = ErrTruncated
		}
		if err != ErrOverflow {
			shiftErr := d.Shift(int(exponent))
			if shiftErr != nil {
				if shiftErr == ErrOverflow {
					negative := d.negative
					maxDecimal(wordBufLen*digitsPerWord, 0, d)
					d.negative = negative
				}
				err = shiftErr
			}
		}
	}
	allZero := true
	for i := 0; i < wordBufLen; i++ {
		if d.wordBuf[i] != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		d.negative = false
	}
	d.resultFrac = d.digitsFrac
	return err
}

// Shift shifts decimal digits in given number (with rounding if it need), shift > 0 means shift to left shift,
// shift < 0 means right shift. In fact it is multiplying on 10^shift.
//
// RETURN
//   eDecOK          OK
//   eDecOverflow    operation lead to overflow, number is untoched
//   eDecTruncated   number was rounded to fit into buffer
//
func (d *MyDecimal) Shift(shift int) error {
	var err error
	if shift == 0 {
		return nil
	}
	var (
		// digitBegin is index of first non zero digit (all indexes from 0).
		digitBegin int
		// digitEnd is index of position after last decimal digit.
		digitEnd int
		// point is index of digit position just after point.
		point = digitsToWords(int(d.digitsInt)) * digitsPerWord
		// new point position.
		newPoint = point + shift
		// number of digits in result.
		digitsInt, digitsFrac int
		newFront              int
	)
	digitBegin, digitEnd = d.digitBounds()
	if digitBegin == digitEnd {
		*d = zeroMyDecimal
		return nil
	}

	digitsInt = newPoint - digitBegin
	if digitsInt < 0 {
		digitsInt = 0
	}
	digitsFrac = digitEnd - newPoint
	if digitsFrac < 0 {
		digitsFrac = 0
	}
	wordsInt := digitsToWords(digitsInt)
	wordsFrac := digitsToWords(digitsFrac)
	newLen := wordsInt + wordsFrac
	if newLen > wordBufLen {
		lack := newLen - wordBufLen
		if wordsFrac < lack {
			return ErrOverflow
		}
		/* cut off fraction part to allow new number to fit in our buffer */
		err = ErrTruncated
		wordsFrac -= lack
		diff := digitsFrac - wordsFrac*digitsPerWord
		err1 := d.Round(d, digitEnd-point-diff, ModeHalfEven)
		if err1 != nil {
			return errors.Trace(err1)
		}
		digitEnd -= diff
		digitsFrac = wordsFrac * digitsPerWord
		if digitEnd <= digitBegin {
			/*
			   We lost all digits (they will be shifted out of buffer), so we can
			   just return 0.
			*/
			*d = zeroMyDecimal
			return ErrTruncated
		}
	}

	if shift%digitsPerWord != 0 {
		var lMiniShift, rMiniShift, miniShift int
		var doLeft bool
		/*
		   Calculate left/right shift to align decimal digits inside our bug
		   digits correctly.
		*/
		if shift > 0 {
			lMiniShift = shift % digitsPerWord
			rMiniShift = digitsPerWord - lMiniShift
			doLeft = lMiniShift <= digitBegin
		} else {
			rMiniShift = (-shift) % digitsPerWord
			lMiniShift = digitsPerWord - rMiniShift
			doLeft = (digitsPerWord*wordBufLen - digitEnd) < rMiniShift
		}
		if doLeft {
			d.doMiniLeftShift(lMiniShift, digitBegin, digitEnd)
			miniShift = -lMiniShift
		} else {
			d.doMiniRightShift(rMiniShift, digitBegin, digitEnd)
			miniShift = rMiniShift
		}
		newPoint += miniShift
		/*
		   If number is shifted and correctly aligned in buffer we can finish.
		*/
		if shift+miniShift == 0 && (newPoint-digitsInt) < digitsPerWord {
			d.digitsInt = int8(digitsInt)
			d.digitsFrac = int8(digitsFrac)
			return err /* already shifted as it should be */
		}
		digitBegin += miniShift
		digitEnd += miniShift
	}

	/* if new 'decimal front' is in first digit, we do not need move digits */
	newFront = newPoint - digitsInt
	if newFront >= digitsPerWord || newFront < 0 {
		/* need to move digits */
		var wordShift int
		if newFront > 0 {
			/* move left */
			wordShift = newFront / digitsPerWord
			to := digitBegin/digitsPerWord - wordShift
			barier := (digitEnd-1)/digitsPerWord - wordShift
			for ; to <= barier; to++ {
				d.wordBuf[to] = d.wordBuf[to+wordShift]
			}
			for barier += wordShift; to <= barier; to++ {
				d.wordBuf[to] = 0
			}
			wordShift = -wordShift
		} else {
			/* move right */
			wordShift = (1 - newFront) / digitsPerWord
			to := (digitEnd-1)/digitsPerWord + wordShift
			barier := digitBegin/digitsPerWord + wordShift
			for ; to >= barier; to-- {
				d.wordBuf[to] = d.wordBuf[to-wordShift]
			}
			for barier -= wordShift; to >= barier; to-- {
				d.wordBuf[to] = 0
			}
		}
		digitShift := wordShift * digitsPerWord
		digitBegin += digitShift
		digitEnd += digitShift
		newPoint += digitShift
	}
	/*
	   If there are gaps then fill them with 0.

	   Only one of following 'for' loops will work because wordIdxBegin <= wordIdxEnd.
	*/
	wordIdxBegin := digitBegin / digitsPerWord
	wordIdxEnd := (digitEnd - 1) / digitsPerWord
	wordIdxNewPoint := 0

	/* We don't want negative new_point below */
	if newPoint != 0 {
		wordIdxNewPoint = (newPoint - 1) / digitsPerWord
	}
	if wordIdxNewPoint > wordIdxEnd {
		for wordIdxNewPoint > wordIdxEnd {
			d.wordBuf[wordIdxNewPoint] = 0
			wordIdxNewPoint--
		}
	} else {
		for ; wordIdxNewPoint < wordIdxBegin; wordIdxNewPoint++ {
			d.wordBuf[wordIdxNewPoint] = 0
		}
	}
	d.digitsInt = int8(digitsInt)
	d.digitsFrac = int8(digitsFrac)
	return err
}

/*
  digitBounds returns bounds of decimal digits in the number.

      start - index (from 0 ) of first decimal digits.
      end   - index of position just after last decimal digit.
*/
func (d *MyDecimal) digitBounds() (start, end int) {
	var i int
	bufBeg := 0
	bufLen := digitsToWords(int(d.digitsInt)) + digitsToWords(int(d.digitsFrac))
	bufEnd := bufLen - 1

	/* find non-zero digit from number beginning */
	for bufBeg < bufLen && d.wordBuf[bufBeg] == 0 {
		bufBeg++
	}
	if bufBeg >= bufLen {
		return 0, 0
	}

	/* find non-zero decimal digit from number beginning */
	if bufBeg == 0 && d.digitsInt > 0 {
		i = (int(d.digitsInt) - 1) % digitsPerWord
		start = digitsPerWord - i - 1
	} else {
		i = digitsPerWord - 1
		start = bufBeg * digitsPerWord
	}
	if bufBeg < bufLen {
		start += countLeadingZeroes(i, d.wordBuf[bufBeg])
	}

	/* find non-zero digit at the end */
	for bufEnd > bufBeg && d.wordBuf[bufEnd] == 0 {
		bufEnd--
	}
	/* find non-zero decimal digit from the end */
	if bufEnd == bufLen-1 && d.digitsFrac > 0 {
		i = (int(d.digitsFrac)-1)%digitsPerWord + 1
		end = bufEnd*digitsPerWord + i
		i = digitsPerWord - i + 1
	} else {
		end = (bufEnd + 1) * digitsPerWord
		i = 1
	}
	end -= countTrailingZeroes(i, d.wordBuf[bufEnd])
	return start, end
}

/*
  doMiniLeftShift does left shift for alignment of data in buffer.

    shift   number of decimal digits on which it should be shifted
    beg/end bounds of decimal digits (see digitsBounds())

  NOTE
    Result fitting in the buffer should be garanted.
    'shift' have to be from 1 to digitsPerWord-1 (inclusive)
*/
func (d *MyDecimal) doMiniLeftShift(shift, beg, end int) {
	bufFrom := beg / digitsPerWord
	bufEnd := (end - 1) / digitsPerWord
	cShift := digitsPerWord - shift
	if beg%digitsPerWord < shift {
		d.wordBuf[bufFrom-1] = d.wordBuf[bufFrom] / powers10[cShift]
	}
	for bufFrom < bufEnd {
		d.wordBuf[bufFrom] = (d.wordBuf[bufFrom]%powers10[cShift])*powers10[shift] + d.wordBuf[bufFrom+1]/powers10[cShift]
		bufFrom++
	}
	d.wordBuf[bufFrom] = (d.wordBuf[bufFrom] % powers10[cShift]) * powers10[shift]
}

/*
  doMiniRightShift does right shift for alignment of data in buffer.

    shift   number of decimal digits on which it should be shifted
    beg/end bounds of decimal digits (see digitsBounds())

  NOTE
    Result fitting in the buffer should be garanted.
    'shift' have to be from 1 to digitsPerWord-1 (inclusive)
*/
func (d *MyDecimal) doMiniRightShift(shift, beg, end int) {
	bufFrom := (end - 1) / digitsPerWord
	bufEnd := beg / digitsPerWord
	cShift := digitsPerWord - shift
	if digitsPerWord-((end-1)%digitsPerWord+1) < shift {
		d.wordBuf[bufFrom+1] = (d.wordBuf[bufFrom] % powers10[shift]) * powers10[cShift]
	}
	for bufFrom > bufEnd {
		d.wordBuf[bufFrom] = d.wordBuf[bufFrom]/powers10[shift] + (d.wordBuf[bufFrom-1]%powers10[shift])*powers10[cShift]
		bufFrom--
	}
	d.wordBuf[bufFrom] = d.wordBuf[bufFrom] / powers10[shift]
}

// Round rounds the decimal to "frac" digits.
//
//    to			- result buffer. d == to is allowed
//    frac			- to what position after fraction point to round. can be negative!
//    roundMode		- round to nearest even or truncate
// 			ModeHalfEven rounds normally.
// 			Truncate just truncates the decimal.
//
// NOTES
//  scale can be negative !
//  one TRUNCATED error (line XXX below) isn't treated very logical :(
//
// RETURN VALUE
//  eDecOK/eDecTruncated
func (d *MyDecimal) Round(to *MyDecimal, frac int, roundMode RoundMode) (err error) {
	// wordsFracTo is the number of fraction words in buffer.
	wordsFracTo := (frac + 1) / digitsPerWord
	if frac > 0 {
		wordsFracTo = digitsToWords(frac)
	}
	wordsFrac := digitsToWords(int(d.digitsFrac))
	wordsInt := digitsToWords(int(d.digitsInt))

	roundDigit := int32(roundMode)
	/* TODO - fix this code as it won't work for CEILING mode */

	if wordsInt+wordsFracTo > wordBufLen {
		wordsFracTo = wordBufLen - wordsInt
		frac = wordsFracTo * digitsPerWord
		err = ErrTruncated
	}
	if int(d.digitsInt)+frac < 0 {
		*to = zeroMyDecimal
		return nil
	}
	if to != d {
		copy(to.wordBuf[:], d.wordBuf[:])
		to.negative = d.negative
		to.digitsInt = int8(myMin(wordsInt, wordBufLen) * digitsPerWord)
	}
	if wordsFracTo > wordsFrac {
		idx := wordsInt + wordsFrac
		for wordsFracTo > wordsFrac {
			wordsFracTo--
			to.wordBuf[idx] = 0
			idx++
		}
		to.digitsFrac = int8(frac)
		to.resultFrac = to.digitsFrac
		return
	}
	if frac >= int(d.digitsFrac) {
		to.digitsFrac = int8(frac)
		to.resultFrac = to.digitsFrac
		return
	}

	// Do increment.
	toIdx := wordsInt + wordsFracTo - 1
	if frac == wordsFracTo*digitsPerWord {
		doInc := false
		switch roundMode {
		// Notice: No support for ceiling mode now.
		case modeCeiling:
			// If any word after scale is not zero, do increment.
			// e.g ceiling 3.0001 to scale 1, gets 3.1
			idx := toIdx + (wordsFrac - wordsFracTo)
			for idx > toIdx {
				if d.wordBuf[idx] != 0 {
					doInc = true
					break
				}
				idx--
			}
		case ModeHalfEven:
			digAfterScale := d.wordBuf[toIdx+1] / digMask // the first digit after scale.
			// If first digit after scale is 5 and round even, do increment if digit at scale is odd.
			doInc = (digAfterScale > 5) || (digAfterScale == 5)
		case ModeTruncate:
			// Never round, just truncate.
			doInc = false
		}
		if doInc {
			if toIdx >= 0 {
				to.wordBuf[toIdx]++
			} else {
				toIdx++
				to.wordBuf[toIdx] = wordBase
			}
		} else if wordsInt+wordsFracTo == 0 {
			*to = zeroMyDecimal
			return nil
		}
	} else {
		/* TODO - fix this code as it won't work for CEILING mode */
		pos := wordsFracTo*digitsPerWord - frac - 1
		shiftedNumber := to.wordBuf[toIdx] / powers10[pos]
		digAfterScale := shiftedNumber % 10
		if digAfterScale > roundDigit || (roundDigit == 5 && digAfterScale == 5) {
			shiftedNumber += 10
		}
		to.wordBuf[toIdx] = powers10[pos] * (shiftedNumber - digAfterScale)
	}
	/*
	   In case we're rounding e.g. 1.5e9 to 2.0e9, the decimal words inside
	   the buffer are as follows.

	   Before <1, 5e8>
	   After  <2, 5e8>

	   Hence we need to set the 2nd field to 0.
	   The same holds if we round 1.5e-9 to 2e-9.
	*/
	if wordsFracTo < wordsFrac {
		idx := wordsInt + wordsFracTo
		if frac == 0 && wordsInt == 0 {
			idx = 1
		}
		for idx < wordBufLen {
			to.wordBuf[idx] = 0
			idx++
		}
	}

	// Handle carry.
	var carry int32
	if to.wordBuf[toIdx] >= wordBase {
		carry = 1
		to.wordBuf[toIdx] -= wordBase
		for carry == 1 && toIdx > 0 {
			toIdx--
			to.wordBuf[toIdx], carry = add(to.wordBuf[toIdx], 0, carry)
		}
		if carry > 0 {
			if wordsInt+wordsFracTo >= wordBufLen {
				wordsFracTo--
				frac = wordsFracTo * digitsPerWord
				err = ErrTruncated
			}
			for toIdx = wordsInt + myMax(wordsFracTo, 0); toIdx > 0; toIdx-- {
				if toIdx < wordBufLen {
					to.wordBuf[toIdx] = to.wordBuf[toIdx-1]
				} else {
					err = ErrOverflow
				}
			}
			to.wordBuf[toIdx] = 1
			/* We cannot have more than 9 * 9 = 81 digits. */
			if int(to.digitsInt) < digitsPerWord*wordBufLen {
				to.digitsInt++
			} else {
				err = ErrOverflow
			}
		}
	} else {
		for {
			if to.wordBuf[toIdx] != 0 {
				break
			}
			if toIdx == 0 {
				/* making 'zero' with the proper scale */
				idx := wordsFracTo + 1
				to.digitsInt = 1
				to.digitsFrac = int8(myMax(frac, 0))
				to.negative = false
				for toIdx < idx {
					to.wordBuf[toIdx] = 0
					toIdx++
				}
				to.resultFrac = to.digitsFrac
				return nil
			}
			toIdx--
		}
	}
	/* Here we check 999.9 -> 1000 case when we need to increase intDigCnt */
	firstDig := mod9[to.digitsInt]
	if firstDig > 0 && to.wordBuf[toIdx] >= powers10[firstDig] {
		to.digitsInt++
	}
	if frac < 0 {
		frac = 0
	}
	to.digitsFrac = int8(frac)
	to.resultFrac = to.digitsFrac
	return
}

func maxDecimal(precision, frac int, to *MyDecimal) {
	digitsInt := precision - frac
	to.negative = false
	to.digitsInt = int8(digitsInt)
	idx := 0
	if digitsInt > 0 {
		firstWordDigits := digitsInt % digitsPerWord
		if firstWordDigits > 0 {
			to.wordBuf[idx] = powers10[firstWordDigits] - 1 /* get 9 99 999 ... */
			idx++
		}
		for digitsInt /= digitsPerWord; digitsInt > 0; digitsInt-- {
			to.wordBuf[idx] = wordMax
			idx++
		}
	}
	to.digitsFrac = int8(frac)
	if frac > 0 {
		lastDigits := frac % digitsPerWord
		for frac /= digitsPerWord; frac > 0; frac-- {
			to.wordBuf[idx] = wordMax
			idx++
		}
		if lastDigits > 0 {
			to.wordBuf[idx] = fracMax[lastDigits-1]
		}
	}
}
