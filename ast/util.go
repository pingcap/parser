// Copyright 2017 PingCAP, Inc.
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

import "strings"

// TextElement is an element of a SQLSentence.
type TextElement struct {
	//
	Next *TextElement

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

type SQLSentence struct {
	root   TextElement
	status sqlSentenceStatus
}

func NewSQLSentence() *SQLSentence {
	s := new(SQLSentence)
	s.Reset()
	return s
}

func (s *SQLSentence) Reset() {
	s.root.Next = &s.root
	s.root.Prev = &s.root
	s.status = sqlSentenceNormal
}

func (s *SQLSentence) Text(text string) *SQLSentence {
	s.TextAround(text, "")
	return s
}

func (s *SQLSentence) TextAround(text string, around string) *SQLSentence {
	s.insertComma()
	s.InsertText(text, s.root.Prev, around)
	return s
}

func (s *SQLSentence) Space() *SQLSentence {
	s.Text(" ")
	return s
}

func (s *SQLSentence) Sentence(sentence *SQLSentence) *SQLSentence {
	s.insertComma()
	s.SentenceAround(sentence, "")
	return s
}

func (s *SQLSentence) SentenceAround(sentence *SQLSentence, around string) *SQLSentence {
	s.InsertSentence(sentence, s.root.Prev, around)
	return s
}

func (s *SQLSentence) InsertText(text string, at *TextElement, around string) {
	if len(around) == 0 {
		s.insertText(text, at)
		return
	}
	s.insertText(around, at)
	s.insertText(text, at)
	s.insertText(around, at)
}

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

func (s *SQLSentence) LeftBracket() *SQLSentence {
	s.Text("(")
	s.status = sqlSentenceNearBrackets
	return s
}

func (s *SQLSentence) RightBracket() *SQLSentence {
	s.status = sqlSentenceNormal
	s.Text(")")
	return s
}

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
