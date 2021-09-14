// Copyright 2015 PingCAP, Inc.
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
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
)

// Lookup returns the encoding with the specified label, and its canonical
// name. It returns nil and the empty string if label is not one of the
// standard encodings for HTML. Matching is case-insensitive and ignores
// leading and trailing whitespace.
func Lookup(label string) (e encoding.Encoding, name string) {
	label = strings.ToLower(strings.Trim(label, "\t\n\r\f "))
	return lookup(Formatted(label))
}

func lookup(label EncodingLabel) (e encoding.Encoding, name string) {
	enc := encodings[string(label)]
	return enc.e, enc.name
}

var encodings = map[string]struct {
	e    encoding.Encoding
	name string
}{
	"unicode-1-1-utf-8":   {encoding.Nop, "utf-8"},
	"utf-8":               {encoding.Nop, "utf-8"},
	"utf8":                {encoding.Nop, "utf-8"},
	"utf8mb4":             {encoding.Nop, "utf-8"},
	"binary":              {encoding.Nop, "binary"},
	"866":                 {charmap.CodePage866, "ibm866"},
	"cp866":               {charmap.CodePage866, "ibm866"},
	"csibm866":            {charmap.CodePage866, "ibm866"},
	"ibm866":              {charmap.CodePage866, "ibm866"},
	"csisolatin2":         {charmap.ISO8859_2, "iso-8859-2"},
	"iso-8859-2":          {charmap.ISO8859_2, "iso-8859-2"},
	"iso-ir-101":          {charmap.ISO8859_2, "iso-8859-2"},
	"iso8859-2":           {charmap.ISO8859_2, "iso-8859-2"},
	"iso88592":            {charmap.ISO8859_2, "iso-8859-2"},
	"iso_8859-2":          {charmap.ISO8859_2, "iso-8859-2"},
	"iso_8859-2:1987":     {charmap.ISO8859_2, "iso-8859-2"},
	"l2":                  {charmap.ISO8859_2, "iso-8859-2"},
	"latin2":              {charmap.ISO8859_2, "iso-8859-2"},
	"csisolatin3":         {charmap.ISO8859_3, "iso-8859-3"},
	"iso-8859-3":          {charmap.ISO8859_3, "iso-8859-3"},
	"iso-ir-109":          {charmap.ISO8859_3, "iso-8859-3"},
	"iso8859-3":           {charmap.ISO8859_3, "iso-8859-3"},
	"iso88593":            {charmap.ISO8859_3, "iso-8859-3"},
	"iso_8859-3":          {charmap.ISO8859_3, "iso-8859-3"},
	"iso_8859-3:1988":     {charmap.ISO8859_3, "iso-8859-3"},
	"l3":                  {charmap.ISO8859_3, "iso-8859-3"},
	"latin3":              {charmap.ISO8859_3, "iso-8859-3"},
	"csisolatin4":         {charmap.ISO8859_4, "iso-8859-4"},
	"iso-8859-4":          {charmap.ISO8859_4, "iso-8859-4"},
	"iso-ir-110":          {charmap.ISO8859_4, "iso-8859-4"},
	"iso8859-4":           {charmap.ISO8859_4, "iso-8859-4"},
	"iso88594":            {charmap.ISO8859_4, "iso-8859-4"},
	"iso_8859-4":          {charmap.ISO8859_4, "iso-8859-4"},
	"iso_8859-4:1988":     {charmap.ISO8859_4, "iso-8859-4"},
	"l4":                  {charmap.ISO8859_4, "iso-8859-4"},
	"latin4":              {charmap.ISO8859_4, "iso-8859-4"},
	"csisolatincyrillic":  {charmap.ISO8859_5, "iso-8859-5"},
	"cyrillic":            {charmap.ISO8859_5, "iso-8859-5"},
	"iso-8859-5":          {charmap.ISO8859_5, "iso-8859-5"},
	"iso-ir-144":          {charmap.ISO8859_5, "iso-8859-5"},
	"iso8859-5":           {charmap.ISO8859_5, "iso-8859-5"},
	"iso88595":            {charmap.ISO8859_5, "iso-8859-5"},
	"iso_8859-5":          {charmap.ISO8859_5, "iso-8859-5"},
	"iso_8859-5:1988":     {charmap.ISO8859_5, "iso-8859-5"},
	"arabic":              {charmap.ISO8859_6, "iso-8859-6"},
	"asmo-708":            {charmap.ISO8859_6, "iso-8859-6"},
	"csiso88596e":         {charmap.ISO8859_6, "iso-8859-6"},
	"csiso88596i":         {charmap.ISO8859_6, "iso-8859-6"},
	"csisolatinarabic":    {charmap.ISO8859_6, "iso-8859-6"},
	"ecma-114":            {charmap.ISO8859_6, "iso-8859-6"},
	"iso-8859-6":          {charmap.ISO8859_6, "iso-8859-6"},
	"iso-8859-6-e":        {charmap.ISO8859_6, "iso-8859-6"},
	"iso-8859-6-i":        {charmap.ISO8859_6, "iso-8859-6"},
	"iso-ir-127":          {charmap.ISO8859_6, "iso-8859-6"},
	"iso8859-6":           {charmap.ISO8859_6, "iso-8859-6"},
	"iso88596":            {charmap.ISO8859_6, "iso-8859-6"},
	"iso_8859-6":          {charmap.ISO8859_6, "iso-8859-6"},
	"iso_8859-6:1987":     {charmap.ISO8859_6, "iso-8859-6"},
	"csisolatingreek":     {charmap.ISO8859_7, "iso-8859-7"},
	"ecma-118":            {charmap.ISO8859_7, "iso-8859-7"},
	"elot_928":            {charmap.ISO8859_7, "iso-8859-7"},
	"greek":               {charmap.ISO8859_7, "iso-8859-7"},
	"greek8":              {charmap.ISO8859_7, "iso-8859-7"},
	"iso-8859-7":          {charmap.ISO8859_7, "iso-8859-7"},
	"iso-ir-126":          {charmap.ISO8859_7, "iso-8859-7"},
	"iso8859-7":           {charmap.ISO8859_7, "iso-8859-7"},
	"iso88597":            {charmap.ISO8859_7, "iso-8859-7"},
	"iso_8859-7":          {charmap.ISO8859_7, "iso-8859-7"},
	"iso_8859-7:1987":     {charmap.ISO8859_7, "iso-8859-7"},
	"sun_eu_greek":        {charmap.ISO8859_7, "iso-8859-7"},
	"csiso88598e":         {charmap.ISO8859_8, "iso-8859-8"},
	"csisolatinhebrew":    {charmap.ISO8859_8, "iso-8859-8"},
	"hebrew":              {charmap.ISO8859_8, "iso-8859-8"},
	"iso-8859-8":          {charmap.ISO8859_8, "iso-8859-8"},
	"iso-8859-8-e":        {charmap.ISO8859_8, "iso-8859-8"},
	"iso-ir-138":          {charmap.ISO8859_8, "iso-8859-8"},
	"iso8859-8":           {charmap.ISO8859_8, "iso-8859-8"},
	"iso88598":            {charmap.ISO8859_8, "iso-8859-8"},
	"iso_8859-8":          {charmap.ISO8859_8, "iso-8859-8"},
	"iso_8859-8:1988":     {charmap.ISO8859_8, "iso-8859-8"},
	"visual":              {charmap.ISO8859_8, "iso-8859-8"},
	"csiso88598i":         {charmap.ISO8859_8, "iso-8859-8-i"},
	"iso-8859-8-i":        {charmap.ISO8859_8, "iso-8859-8-i"},
	"logical":             {charmap.ISO8859_8, "iso-8859-8-i"},
	"csisolatin6":         {charmap.ISO8859_10, "iso-8859-10"},
	"iso-8859-10":         {charmap.ISO8859_10, "iso-8859-10"},
	"iso-ir-157":          {charmap.ISO8859_10, "iso-8859-10"},
	"iso8859-10":          {charmap.ISO8859_10, "iso-8859-10"},
	"iso885910":           {charmap.ISO8859_10, "iso-8859-10"},
	"l6":                  {charmap.ISO8859_10, "iso-8859-10"},
	"latin6":              {charmap.ISO8859_10, "iso-8859-10"},
	"iso-8859-13":         {charmap.ISO8859_13, "iso-8859-13"},
	"iso8859-13":          {charmap.ISO8859_13, "iso-8859-13"},
	"iso885913":           {charmap.ISO8859_13, "iso-8859-13"},
	"iso-8859-14":         {charmap.ISO8859_14, "iso-8859-14"},
	"iso8859-14":          {charmap.ISO8859_14, "iso-8859-14"},
	"iso885914":           {charmap.ISO8859_14, "iso-8859-14"},
	"csisolatin9":         {charmap.ISO8859_15, "iso-8859-15"},
	"iso-8859-15":         {charmap.ISO8859_15, "iso-8859-15"},
	"iso8859-15":          {charmap.ISO8859_15, "iso-8859-15"},
	"iso885915":           {charmap.ISO8859_15, "iso-8859-15"},
	"iso_8859-15":         {charmap.ISO8859_15, "iso-8859-15"},
	"l9":                  {charmap.ISO8859_15, "iso-8859-15"},
	"iso-8859-16":         {charmap.ISO8859_16, "iso-8859-16"},
	"cskoi8r":             {charmap.KOI8R, "koi8-r"},
	"koi":                 {charmap.KOI8R, "koi8-r"},
	"koi8":                {charmap.KOI8R, "koi8-r"},
	"koi8-r":              {charmap.KOI8R, "koi8-r"},
	"koi8_r":              {charmap.KOI8R, "koi8-r"},
	"koi8-u":              {charmap.KOI8U, "koi8-u"},
	"csmacintosh":         {charmap.Macintosh, "macintosh"},
	"mac":                 {charmap.Macintosh, "macintosh"},
	"macintosh":           {charmap.Macintosh, "macintosh"},
	"x-mac-roman":         {charmap.Macintosh, "macintosh"},
	"dos-874":             {charmap.Windows874, "windows-874"},
	"iso-8859-11":         {charmap.Windows874, "windows-874"},
	"iso8859-11":          {charmap.Windows874, "windows-874"},
	"iso885911":           {charmap.Windows874, "windows-874"},
	"tis-620":             {charmap.Windows874, "windows-874"},
	"windows-874":         {charmap.Windows874, "windows-874"},
	"cp1250":              {charmap.Windows1250, "windows-1250"},
	"windows-1250":        {charmap.Windows1250, "windows-1250"},
	"x-cp1250":            {charmap.Windows1250, "windows-1250"},
	"cp1251":              {charmap.Windows1251, "windows-1251"},
	"windows-1251":        {charmap.Windows1251, "windows-1251"},
	"x-cp1251":            {charmap.Windows1251, "windows-1251"},
	"ansi_x3.4-1968":      {charmap.Windows1252, "windows-1252"},
	"ascii":               {charmap.Windows1252, "windows-1252"},
	"cp1252":              {charmap.Windows1252, "windows-1252"},
	"cp819":               {charmap.Windows1252, "windows-1252"},
	"csisolatin1":         {charmap.Windows1252, "windows-1252"},
	"ibm819":              {charmap.Windows1252, "windows-1252"},
	"iso-8859-1":          {charmap.Windows1252, "windows-1252"},
	"iso-ir-100":          {charmap.Windows1252, "windows-1252"},
	"iso8859-1":           {charmap.Windows1252, "windows-1252"},
	"iso88591":            {charmap.Windows1252, "windows-1252"},
	"iso_8859-1":          {charmap.Windows1252, "windows-1252"},
	"iso_8859-1:1987":     {charmap.Windows1252, "windows-1252"},
	"l1":                  {charmap.Windows1252, "windows-1252"},
	"latin1":              {charmap.Windows1252, "windows-1252"},
	"us-ascii":            {charmap.Windows1252, "windows-1252"},
	"windows-1252":        {charmap.Windows1252, "windows-1252"},
	"x-cp1252":            {charmap.Windows1252, "windows-1252"},
	"cp1253":              {charmap.Windows1253, "windows-1253"},
	"windows-1253":        {charmap.Windows1253, "windows-1253"},
	"x-cp1253":            {charmap.Windows1253, "windows-1253"},
	"cp1254":              {charmap.Windows1254, "windows-1254"},
	"csisolatin5":         {charmap.Windows1254, "windows-1254"},
	"iso-8859-9":          {charmap.Windows1254, "windows-1254"},
	"iso-ir-148":          {charmap.Windows1254, "windows-1254"},
	"iso8859-9":           {charmap.Windows1254, "windows-1254"},
	"iso88599":            {charmap.Windows1254, "windows-1254"},
	"iso_8859-9":          {charmap.Windows1254, "windows-1254"},
	"iso_8859-9:1989":     {charmap.Windows1254, "windows-1254"},
	"l5":                  {charmap.Windows1254, "windows-1254"},
	"latin5":              {charmap.Windows1254, "windows-1254"},
	"windows-1254":        {charmap.Windows1254, "windows-1254"},
	"x-cp1254":            {charmap.Windows1254, "windows-1254"},
	"cp1255":              {charmap.Windows1255, "windows-1255"},
	"windows-1255":        {charmap.Windows1255, "windows-1255"},
	"x-cp1255":            {charmap.Windows1255, "windows-1255"},
	"cp1256":              {charmap.Windows1256, "windows-1256"},
	"windows-1256":        {charmap.Windows1256, "windows-1256"},
	"x-cp1256":            {charmap.Windows1256, "windows-1256"},
	"cp1257":              {charmap.Windows1257, "windows-1257"},
	"windows-1257":        {charmap.Windows1257, "windows-1257"},
	"x-cp1257":            {charmap.Windows1257, "windows-1257"},
	"cp1258":              {charmap.Windows1258, "windows-1258"},
	"windows-1258":        {charmap.Windows1258, "windows-1258"},
	"x-cp1258":            {charmap.Windows1258, "windows-1258"},
	"x-mac-cyrillic":      {charmap.MacintoshCyrillic, "x-mac-cyrillic"},
	"x-mac-ukrainian":     {charmap.MacintoshCyrillic, "x-mac-cyrillic"},
	"chinese":             {simplifiedchinese.GBK, "gbk"},
	"csgb2312":            {simplifiedchinese.GBK, "gbk"},
	"csiso58gb231280":     {simplifiedchinese.GBK, "gbk"},
	"gb2312":              {simplifiedchinese.GBK, "gbk"},
	"gb_2312":             {simplifiedchinese.GBK, "gbk"},
	"gb_2312-80":          {simplifiedchinese.GBK, "gbk"},
	"gbk":                 {simplifiedchinese.GBK, "gbk"},
	"iso-ir-58":           {simplifiedchinese.GBK, "gbk"},
	"x-gbk":               {simplifiedchinese.GBK, "gbk"},
	"gb18030":             {simplifiedchinese.GB18030, "gb18030"},
	"hz-gb-2312":          {simplifiedchinese.HZGB2312, "hz-gb-2312"},
	"big5":                {traditionalchinese.Big5, "big5"},
	"big5-hkscs":          {traditionalchinese.Big5, "big5"},
	"cn-big5":             {traditionalchinese.Big5, "big5"},
	"csbig5":              {traditionalchinese.Big5, "big5"},
	"x-x-big5":            {traditionalchinese.Big5, "big5"},
	"cseucpkdfmtjapanese": {japanese.EUCJP, "euc-jp"},
	"euc-jp":              {japanese.EUCJP, "euc-jp"},
	"x-euc-jp":            {japanese.EUCJP, "euc-jp"},
	"csiso2022jp":         {japanese.ISO2022JP, "iso-2022-jp"},
	"iso-2022-jp":         {japanese.ISO2022JP, "iso-2022-jp"},
	"csshiftjis":          {japanese.ShiftJIS, "shift_jis"},
	"ms_kanji":            {japanese.ShiftJIS, "shift_jis"},
	"shift-jis":           {japanese.ShiftJIS, "shift_jis"},
	"shift_jis":           {japanese.ShiftJIS, "shift_jis"},
	"sjis":                {japanese.ShiftJIS, "shift_jis"},
	"windows-31j":         {japanese.ShiftJIS, "shift_jis"},
	"x-sjis":              {japanese.ShiftJIS, "shift_jis"},
	"cseuckr":             {korean.EUCKR, "euc-kr"},
	"csksc56011987":       {korean.EUCKR, "euc-kr"},
	"euc-kr":              {korean.EUCKR, "euc-kr"},
	"iso-ir-149":          {korean.EUCKR, "euc-kr"},
	"korean":              {korean.EUCKR, "euc-kr"},
	"ks_c_5601-1987":      {korean.EUCKR, "euc-kr"},
	"ks_c_5601-1989":      {korean.EUCKR, "euc-kr"},
	"ksc5601":             {korean.EUCKR, "euc-kr"},
	"ksc_5601":            {korean.EUCKR, "euc-kr"},
	"windows-949":         {korean.EUCKR, "euc-kr"},
	"csiso2022kr":         {encoding.Replacement, "replacement"},
	"iso-2022-kr":         {encoding.Replacement, "replacement"},
	"iso-2022-cn":         {encoding.Replacement, "replacement"},
	"iso-2022-cn-ext":     {encoding.Replacement, "replacement"},
	"utf-16be":            {unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM), "utf-16be"},
	"utf-16":              {unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), "utf-16le"},
	"utf-16le":            {unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), "utf-16le"},
	"x-user-defined":      {charmap.XUserDefined, "x-user-defined"},
}

// FindNextCharacterLength is used in lexer.peek() to determine the next character length.
func FindNextCharacterLength(label string) func([]byte) int {
	if f, ok := encodingNextCharacterLength[label]; ok {
		return f
	}
	return nil
}

var encodingNextCharacterLength = map[string]func([]byte) int{
	// https://en.wikipedia.org/wiki/GBK_(character_encoding)#Layout_diagram
	"gbk": func(bs []byte) int {
		if len(bs) == 0 || bs[0] < 0x80 {
			// A byte in the range 00–7F is a single byte that means the same thing as it does in ASCII.
			return 1
		}
		return 2
	},
	"utf-8": func(bs []byte) int {
		if len(bs) == 0 || bs[0] < 0x80 {
			return 1
		} else if bs[0] < 0xe0 {
			return 2
		} else if bs[0] < 0xf0 {
			return 3
		}
		return 4
	},
}
