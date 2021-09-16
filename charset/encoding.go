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

package charset

import (
	"fmt"
	"strings"

	"github.com/cznic/mathutil"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/terror"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

const encodingLegacy = "utf-8" // utf-8 encoding is compatible with old default behavior.

var errInvalidCharacterString = terror.ClassParser.NewStd(mysql.ErrInvalidCharacterString)

type EncodingLabel string

// Format trim and change the label to lowercase.
func Format(label string) EncodingLabel {
	return EncodingLabel(strings.ToLower(strings.Trim(label, "\t\n\r\f ")))
}

// Formatted is used when the label is already trimmed and it is lowercase.
func Formatted(label string) EncodingLabel {
	return EncodingLabel(label)
}

// Encoding provide a interface to encode/decode a string with specific encoding.
type Encoding struct {
	enc        encoding.Encoding
	name       string
	charLength func([]byte) int
}

// Enabled indicates whether the non-utf8 encoding is used.
func (e *Encoding) Enabled() bool {
	return e.enc != nil && e.charLength != nil
}

// Name returns the name of the current encoding.
func (e *Encoding) Name() string {
	return e.name
}

// NewEncoding creates a new Encoding.
func NewEncoding(label string) *Encoding {
	if len(label) == 0 {
		return &Encoding{}
	}
	e, name := Lookup(label)
	if e != nil && name != encodingLegacy {
		return &Encoding{
			enc:        e,
			name:       name,
			charLength: FindNextCharacterLength(name),
		}
	}
	return &Encoding{name: name}
}

// UpdateEncoding updates to a new Encoding.
func (e *Encoding) UpdateEncoding(label EncodingLabel) {
	enc, name := lookup(label)
	e.name = name
	if enc != nil && name != encodingLegacy {
		e.enc = enc
		e.charLength = FindNextCharacterLength(name)
	} else {
		e.enc = nil
		e.charLength = nil
	}
}

// Encode encodes the bytes to a string.
func (e *Encoding) Encode(dest, src []byte) ([]byte, error) {
	return e.transform(e.enc.NewEncoder(), dest, src, false)
}

// Decode decodes the bytes to a string. Please copy the result after calling this.
func (e *Encoding) Decode(dest, src []byte) ([]byte, error) {
	return e.transform(e.enc.NewDecoder(), dest, src, true)
}

func (e *Encoding) transform(transformer transform.Transformer, dest, src []byte, isDecoding bool) ([]byte, error) {
	if len(dest) < len(src) {
		dest = make([]byte, len(src)*2)
	}
	var destOffset, srcOffset int
	var encodingErr error
	for {
		srcEnd, nextLen := len(src), 1
		if isDecoding {
			if e.charLength != nil {
				nextLen = e.charLength(src[srcOffset:])
			}
			srcEnd = srcOffset + nextLen
			if srcEnd > len(src) {
				srcEnd = len(src)
			}
		}
		nDest, nSrc, err := transformer.Transform(dest[destOffset:], src[srcOffset:srcEnd], false)
		if err == transform.ErrShortDst {
			// The destination buffer is too small. Enlarge the capacity.
			newDest := make([]byte, len(dest)*2)
			copy(newDest, dest)
			dest = newDest
		} else if err != nil || isDecoding && e.startWithReplacementChar(dest[destOffset:destOffset+nDest]) {
			if encodingErr == nil {
				cutEnd := mathutil.Min(srcOffset+nextLen, len(src))
				invalidBytes := fmt.Sprintf("%X", string(src[srcOffset:cutEnd]))
				encodingErr = errInvalidCharacterString.GenWithStackByArgs(e.name, invalidBytes)
			}
			// Append '?' to the destination buffer.
			dest[destOffset] = byte('?')
			nDest = 1
			nSrc = nextLen // skip the source bytes that cannot be decoded normally.
		}
		destOffset += nDest
		srcOffset += nSrc
		// The source bytes are exhausted.
		if srcOffset >= len(src) {
			return dest[:destOffset], encodingErr
		}
	}
}

// startWithReplacementChar indicates the first bytes are 0xEFBFBD.
func (e *Encoding) startWithReplacementChar(dst []byte) bool {
	return len(dst) >= 3 && dst[0] == 0xEF && dst[1] == 0xBF && dst[2] == 0xBD
}
