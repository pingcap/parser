// Copyright 2018 PingCAP, Inc.
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

package ast

import (
	"strings"
)

// TextElement is an element of a SQLSentence.
type TextElement struct {
	// Next is the next element in a SQLSentence.
	Next *TextElement

	// Next is the previous element in a SQLSentence.
	Prev *TextElement

	// The text stored with this element.
	Text string
}

type sqlSentenceStatus int8

const (
	sqlSentenceNormal sqlSentenceStatus = iota
	sqlSentenceNearBrackets
	sqlSentenceInBrackets
)

// SQLSentence express a sql text.
// In order to avoid copying when string splicing occurs,
// SQLSentence uses a linked list structure to store strings.
type SQLSentence struct {
	root   TextElement
	status sqlSentenceStatus
}

// NewSQLSentence returns a empty SQLSentence.
func NewSQLSentence() *SQLSentence {
	s := new(SQLSentence)
	s.Reset()
	return s
}

// Reset makes the SQLSentence empty.
func (s *SQLSentence) Reset() {
	s.root.Next = &s.root
	s.root.Prev = &s.root
	s.status = sqlSentenceNormal
}

// Text will append text to SQLSentence.
func (s *SQLSentence) Text(text string) *SQLSentence {
	s.TextAround(text, "")
	return s
}

// TextAround will add `around` around `text` then append to SQLSentence.
// Example: s.TextAround("xx", "tt") will append "ttxxtt" to SQLSentence.
func (s *SQLSentence) TextAround(text string, around string) *SQLSentence {
	s.insertComma()
	s.InsertText(text, s.root.Prev, around)
	return s
}

// Space will append " " to SQLSentence.
func (s *SQLSentence) Space() *SQLSentence {
	s.InsertText(" ", s.root.Prev, "")
	return s
}

// Sentence will append sentence to SQLSentence.
func (s *SQLSentence) Sentence(sentence *SQLSentence) *SQLSentence {
	s.SentenceAround(sentence, "")
	return s
}

// SentenceAround will add `around` around `sentence` then append to SQLSentence.
// Just like TextAround
func (s *SQLSentence) SentenceAround(sentence *SQLSentence, around string) *SQLSentence {
	s.insertComma()
	s.InsertSentence(sentence, s.root.Prev, around)
	return s
}

// InsertText inserts `text` surrounded by `around` after at
func (s *SQLSentence) InsertText(text string, at *TextElement, around string) {
	if len(around) == 0 {
		s.insertText(text, at)
		return
	}
	s.insertText(around, at)
	s.insertText(text, at)
	s.insertText(around, at)
}

// InsertSentence inserts `sentence` surrounded by `around` after at
func (s *SQLSentence) InsertSentence(sentence *SQLSentence, at *TextElement, around string) {
	if len(around) == 0 {
		s.insertSentence(sentence, at)
		return
	}
	s.insertText(around, at)
	s.insertSentence(sentence, at)
	s.insertText(around, at)
	sentence.Reset()
}

// LeftBracket append a "(" to SQLSentence,
// and after that all append operations(Text, TextAround, Sentence, SentenceAround) will automatically append a comma.
func (s *SQLSentence) LeftBracket() *SQLSentence {
	s.Text("(")
	s.status = sqlSentenceNearBrackets
	return s
}

// RightBracket append a ")" to SQLSentence,
// and no longer automatically add commas.
func (s *SQLSentence) RightBracket() *SQLSentence {
	s.status = sqlSentenceNormal
	s.Text(")")
	return s
}

// String returns the sql text from SQLSentence
func (s *SQLSentence) String() string {
	var b strings.Builder
	p := &s.root
	for p.Next != &s.root {
		p = p.Next
		b.WriteString(p.Text)
	}
	return b.String()
}

func (s *SQLSentence) insertText(text string, at *TextElement) {
	e := &TextElement{Text: text}
	n := at.Next
	at.Next = e
	e.Prev = at
	e.Next = n
	n.Prev = e
}

func (s *SQLSentence) insertSentence(sentence *SQLSentence, at *TextElement) {
	n := at.Next
	at.Next = sentence.root.Next
	sentence.root.Next.Prev = at
	sentence.root.Prev.Next = n
	n.Prev = sentence.root.Prev
}

func (s *SQLSentence) insertComma() {
	if s.status == sqlSentenceInBrackets {
		s.insertText(", ", s.root.Prev)
		return
	}
	if s.status == sqlSentenceNearBrackets {
		s.status = sqlSentenceInBrackets
	}
}
