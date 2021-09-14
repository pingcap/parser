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
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

const (
	encodingBufferSizeRecycleThreshold = 4 * 1024

	encodingLegacy = "utf-8" // utf-8 encoding is compatible with old default behavior.
)

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
	buffer     []byte
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
func NewEncoding(label EncodingLabel) *Encoding {
	if len(label) == 0 {
		return &Encoding{}
	}
	e, name := lookup(label)
	if e != nil && name != encodingLegacy {
		return &Encoding{
			enc:        e,
			name:       name,
			charLength: FindNextCharacterLength(name),
		}
	}
	return &Encoding{name: name}
}

// UpdateEncoding updates to a new Encoding without changing the buffer.
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
func (e *Encoding) Encode(src []byte) (string, bool) {
	return e.transform(e.enc.NewEncoder(), src, false)
}

// Decode decodes the bytes to a string.
func (e *Encoding) Decode(src []byte) (string, bool) {
	return e.transform(e.enc.NewDecoder(), src, true)
}

func (e *Encoding) transform(transformer transform.Transformer, src []byte, isDecoding bool) (string, bool) {
	if len(e.buffer) < len(src) {
		e.buffer = make([]byte, len(src)*2)
	}
	var destOffset, srcOffset int
	ok := true
	for {
		srcEnd := len(src)
		if isDecoding {
			srcEnd = e.nextCharacterOffset(src, srcOffset)
		}
		nDest, nSrc, err := transformer.Transform(e.buffer[destOffset:], src[srcOffset:srcEnd], false)
		destOffset += nDest
		srcOffset += nSrc
		if err == nil {
			if srcOffset >= len(src) {
				result := string(e.buffer[:destOffset])
				if len(e.buffer) > encodingBufferSizeRecycleThreshold {
					// This prevents Encoding from holding too much memory.
					e.buffer = nil
				}
				return result, ok
			}
		} else if err == transform.ErrShortDst {
			newDest := make([]byte, len(e.buffer)*2)
			copy(newDest, e.buffer)
			e.buffer = newDest
		} else {
			e.buffer[destOffset] = byte('?')
			destOffset += 1
			srcOffset += 1
			ok = false
		}
	}
}

func (e *Encoding) nextCharacterOffset(src []byte, srcOffset int) int {
	nextLen := 4
	if e.charLength != nil {
		nextLen = e.charLength(src[srcOffset:])
	}
	srcEnd := srcOffset + nextLen
	if srcEnd > len(src) {
		srcEnd = len(src)
	}
	return srcEnd
}
