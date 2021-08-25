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
	enc := encodings[label]
	return enc.e, enc.name
}

func LookupCharLength(label string) func([]byte) int {
	label = strings.ToLower(strings.Trim(label, "\t\n\r\f "))
	enc := encodings[label]
	return enc.charLength
}

var encodings = map[string]struct {
	e          encoding.Encoding
	name       string
	charLength func([]byte) int
}{
	"unicode-1-1-utf-8":   {encoding.Nop, "utf-8", defaultLength},
	"utf-8":               {encoding.Nop, "utf-8", defaultLength},
	"utf8":                {encoding.Nop, "utf-8", defaultLength},
	"utf8mb4":             {encoding.Nop, "utf-8", defaultLength},
	"binary":              {encoding.Nop, "binary", defaultLength},
	"866":                 {charmap.CodePage866, "ibm866", defaultLength},
	"cp866":               {charmap.CodePage866, "ibm866", defaultLength},
	"csibm866":            {charmap.CodePage866, "ibm866", defaultLength},
	"ibm866":              {charmap.CodePage866, "ibm866", defaultLength},
	"csisolatin2":         {charmap.ISO8859_2, "iso-8859-2", defaultLength},
	"iso-8859-2":          {charmap.ISO8859_2, "iso-8859-2", defaultLength},
	"iso-ir-101":          {charmap.ISO8859_2, "iso-8859-2", defaultLength},
	"iso8859-2":           {charmap.ISO8859_2, "iso-8859-2", defaultLength},
	"iso88592":            {charmap.ISO8859_2, "iso-8859-2", defaultLength},
	"iso_8859-2":          {charmap.ISO8859_2, "iso-8859-2", defaultLength},
	"iso_8859-2:1987":     {charmap.ISO8859_2, "iso-8859-2", defaultLength},
	"l2":                  {charmap.ISO8859_2, "iso-8859-2", defaultLength},
	"latin2":              {charmap.ISO8859_2, "iso-8859-2", defaultLength},
	"csisolatin3":         {charmap.ISO8859_3, "iso-8859-3", defaultLength},
	"iso-8859-3":          {charmap.ISO8859_3, "iso-8859-3", defaultLength},
	"iso-ir-109":          {charmap.ISO8859_3, "iso-8859-3", defaultLength},
	"iso8859-3":           {charmap.ISO8859_3, "iso-8859-3", defaultLength},
	"iso88593":            {charmap.ISO8859_3, "iso-8859-3", defaultLength},
	"iso_8859-3":          {charmap.ISO8859_3, "iso-8859-3", defaultLength},
	"iso_8859-3:1988":     {charmap.ISO8859_3, "iso-8859-3", defaultLength},
	"l3":                  {charmap.ISO8859_3, "iso-8859-3", defaultLength},
	"latin3":              {charmap.ISO8859_3, "iso-8859-3", defaultLength},
	"csisolatin4":         {charmap.ISO8859_4, "iso-8859-4", defaultLength},
	"iso-8859-4":          {charmap.ISO8859_4, "iso-8859-4", defaultLength},
	"iso-ir-110":          {charmap.ISO8859_4, "iso-8859-4", defaultLength},
	"iso8859-4":           {charmap.ISO8859_4, "iso-8859-4", defaultLength},
	"iso88594":            {charmap.ISO8859_4, "iso-8859-4", defaultLength},
	"iso_8859-4":          {charmap.ISO8859_4, "iso-8859-4", defaultLength},
	"iso_8859-4:1988":     {charmap.ISO8859_4, "iso-8859-4", defaultLength},
	"l4":                  {charmap.ISO8859_4, "iso-8859-4", defaultLength},
	"latin4":              {charmap.ISO8859_4, "iso-8859-4", defaultLength},
	"csisolatincyrillic":  {charmap.ISO8859_5, "iso-8859-5", defaultLength},
	"cyrillic":            {charmap.ISO8859_5, "iso-8859-5", defaultLength},
	"iso-8859-5":          {charmap.ISO8859_5, "iso-8859-5", defaultLength},
	"iso-ir-144":          {charmap.ISO8859_5, "iso-8859-5", defaultLength},
	"iso8859-5":           {charmap.ISO8859_5, "iso-8859-5", defaultLength},
	"iso88595":            {charmap.ISO8859_5, "iso-8859-5", defaultLength},
	"iso_8859-5":          {charmap.ISO8859_5, "iso-8859-5", defaultLength},
	"iso_8859-5:1988":     {charmap.ISO8859_5, "iso-8859-5", defaultLength},
	"arabic":              {charmap.ISO8859_6, "iso-8859-6", defaultLength},
	"asmo-708":            {charmap.ISO8859_6, "iso-8859-6", defaultLength},
	"csiso88596e":         {charmap.ISO8859_6, "iso-8859-6", defaultLength},
	"csiso88596i":         {charmap.ISO8859_6, "iso-8859-6", defaultLength},
	"csisolatinarabic":    {charmap.ISO8859_6, "iso-8859-6", defaultLength},
	"ecma-114":            {charmap.ISO8859_6, "iso-8859-6", defaultLength},
	"iso-8859-6":          {charmap.ISO8859_6, "iso-8859-6", defaultLength},
	"iso-8859-6-e":        {charmap.ISO8859_6, "iso-8859-6", defaultLength},
	"iso-8859-6-i":        {charmap.ISO8859_6, "iso-8859-6", defaultLength},
	"iso-ir-127":          {charmap.ISO8859_6, "iso-8859-6", defaultLength},
	"iso8859-6":           {charmap.ISO8859_6, "iso-8859-6", defaultLength},
	"iso88596":            {charmap.ISO8859_6, "iso-8859-6", defaultLength},
	"iso_8859-6":          {charmap.ISO8859_6, "iso-8859-6", defaultLength},
	"iso_8859-6:1987":     {charmap.ISO8859_6, "iso-8859-6", defaultLength},
	"csisolatingreek":     {charmap.ISO8859_7, "iso-8859-7", defaultLength},
	"ecma-118":            {charmap.ISO8859_7, "iso-8859-7", defaultLength},
	"elot_928":            {charmap.ISO8859_7, "iso-8859-7", defaultLength},
	"greek":               {charmap.ISO8859_7, "iso-8859-7", defaultLength},
	"greek8":              {charmap.ISO8859_7, "iso-8859-7", defaultLength},
	"iso-8859-7":          {charmap.ISO8859_7, "iso-8859-7", defaultLength},
	"iso-ir-126":          {charmap.ISO8859_7, "iso-8859-7", defaultLength},
	"iso8859-7":           {charmap.ISO8859_7, "iso-8859-7", defaultLength},
	"iso88597":            {charmap.ISO8859_7, "iso-8859-7", defaultLength},
	"iso_8859-7":          {charmap.ISO8859_7, "iso-8859-7", defaultLength},
	"iso_8859-7:1987":     {charmap.ISO8859_7, "iso-8859-7", defaultLength},
	"sun_eu_greek":        {charmap.ISO8859_7, "iso-8859-7", defaultLength},
	"csiso88598e":         {charmap.ISO8859_8, "iso-8859-8", defaultLength},
	"csisolatinhebrew":    {charmap.ISO8859_8, "iso-8859-8", defaultLength},
	"hebrew":              {charmap.ISO8859_8, "iso-8859-8", defaultLength},
	"iso-8859-8":          {charmap.ISO8859_8, "iso-8859-8", defaultLength},
	"iso-8859-8-e":        {charmap.ISO8859_8, "iso-8859-8", defaultLength},
	"iso-ir-138":          {charmap.ISO8859_8, "iso-8859-8", defaultLength},
	"iso8859-8":           {charmap.ISO8859_8, "iso-8859-8", defaultLength},
	"iso88598":            {charmap.ISO8859_8, "iso-8859-8", defaultLength},
	"iso_8859-8":          {charmap.ISO8859_8, "iso-8859-8", defaultLength},
	"iso_8859-8:1988":     {charmap.ISO8859_8, "iso-8859-8", defaultLength},
	"visual":              {charmap.ISO8859_8, "iso-8859-8", defaultLength},
	"csiso88598i":         {charmap.ISO8859_8, "iso-8859-8-i", defaultLength},
	"iso-8859-8-i":        {charmap.ISO8859_8, "iso-8859-8-i", defaultLength},
	"logical":             {charmap.ISO8859_8, "iso-8859-8-i", defaultLength},
	"csisolatin6":         {charmap.ISO8859_10, "iso-8859-10", defaultLength},
	"iso-8859-10":         {charmap.ISO8859_10, "iso-8859-10", defaultLength},
	"iso-ir-157":          {charmap.ISO8859_10, "iso-8859-10", defaultLength},
	"iso8859-10":          {charmap.ISO8859_10, "iso-8859-10", defaultLength},
	"iso885910":           {charmap.ISO8859_10, "iso-8859-10", defaultLength},
	"l6":                  {charmap.ISO8859_10, "iso-8859-10", defaultLength},
	"latin6":              {charmap.ISO8859_10, "iso-8859-10", defaultLength},
	"iso-8859-13":         {charmap.ISO8859_13, "iso-8859-13", defaultLength},
	"iso8859-13":          {charmap.ISO8859_13, "iso-8859-13", defaultLength},
	"iso885913":           {charmap.ISO8859_13, "iso-8859-13", defaultLength},
	"iso-8859-14":         {charmap.ISO8859_14, "iso-8859-14", defaultLength},
	"iso8859-14":          {charmap.ISO8859_14, "iso-8859-14", defaultLength},
	"iso885914":           {charmap.ISO8859_14, "iso-8859-14", defaultLength},
	"csisolatin9":         {charmap.ISO8859_15, "iso-8859-15", defaultLength},
	"iso-8859-15":         {charmap.ISO8859_15, "iso-8859-15", defaultLength},
	"iso8859-15":          {charmap.ISO8859_15, "iso-8859-15", defaultLength},
	"iso885915":           {charmap.ISO8859_15, "iso-8859-15", defaultLength},
	"iso_8859-15":         {charmap.ISO8859_15, "iso-8859-15", defaultLength},
	"l9":                  {charmap.ISO8859_15, "iso-8859-15", defaultLength},
	"iso-8859-16":         {charmap.ISO8859_16, "iso-8859-16", defaultLength},
	"cskoi8r":             {charmap.KOI8R, "koi8-r", defaultLength},
	"koi":                 {charmap.KOI8R, "koi8-r", defaultLength},
	"koi8":                {charmap.KOI8R, "koi8-r", defaultLength},
	"koi8-r":              {charmap.KOI8R, "koi8-r", defaultLength},
	"koi8_r":              {charmap.KOI8R, "koi8-r", defaultLength},
	"koi8-u":              {charmap.KOI8U, "koi8-u", defaultLength},
	"csmacintosh":         {charmap.Macintosh, "macintosh", defaultLength},
	"mac":                 {charmap.Macintosh, "macintosh", defaultLength},
	"macintosh":           {charmap.Macintosh, "macintosh", defaultLength},
	"x-mac-roman":         {charmap.Macintosh, "macintosh", defaultLength},
	"dos-874":             {charmap.Windows874, "windows-874", defaultLength},
	"iso-8859-11":         {charmap.Windows874, "windows-874", defaultLength},
	"iso8859-11":          {charmap.Windows874, "windows-874", defaultLength},
	"iso885911":           {charmap.Windows874, "windows-874", defaultLength},
	"tis-620":             {charmap.Windows874, "windows-874", defaultLength},
	"windows-874":         {charmap.Windows874, "windows-874", defaultLength},
	"cp1250":              {charmap.Windows1250, "windows-1250", defaultLength},
	"windows-1250":        {charmap.Windows1250, "windows-1250", defaultLength},
	"x-cp1250":            {charmap.Windows1250, "windows-1250", defaultLength},
	"cp1251":              {charmap.Windows1251, "windows-1251", defaultLength},
	"windows-1251":        {charmap.Windows1251, "windows-1251", defaultLength},
	"x-cp1251":            {charmap.Windows1251, "windows-1251", defaultLength},
	"ansi_x3.4-1968":      {charmap.Windows1252, "windows-1252", defaultLength},
	"ascii":               {charmap.Windows1252, "windows-1252", defaultLength},
	"cp1252":              {charmap.Windows1252, "windows-1252", defaultLength},
	"cp819":               {charmap.Windows1252, "windows-1252", defaultLength},
	"csisolatin1":         {charmap.Windows1252, "windows-1252", defaultLength},
	"ibm819":              {charmap.Windows1252, "windows-1252", defaultLength},
	"iso-8859-1":          {charmap.Windows1252, "windows-1252", defaultLength},
	"iso-ir-100":          {charmap.Windows1252, "windows-1252", defaultLength},
	"iso8859-1":           {charmap.Windows1252, "windows-1252", defaultLength},
	"iso88591":            {charmap.Windows1252, "windows-1252", defaultLength},
	"iso_8859-1":          {charmap.Windows1252, "windows-1252", defaultLength},
	"iso_8859-1:1987":     {charmap.Windows1252, "windows-1252", defaultLength},
	"l1":                  {charmap.Windows1252, "windows-1252", defaultLength},
	"latin1":              {charmap.Windows1252, "windows-1252", defaultLength},
	"us-ascii":            {charmap.Windows1252, "windows-1252", defaultLength},
	"windows-1252":        {charmap.Windows1252, "windows-1252", defaultLength},
	"x-cp1252":            {charmap.Windows1252, "windows-1252", defaultLength},
	"cp1253":              {charmap.Windows1253, "windows-1253", defaultLength},
	"windows-1253":        {charmap.Windows1253, "windows-1253", defaultLength},
	"x-cp1253":            {charmap.Windows1253, "windows-1253", defaultLength},
	"cp1254":              {charmap.Windows1254, "windows-1254", defaultLength},
	"csisolatin5":         {charmap.Windows1254, "windows-1254", defaultLength},
	"iso-8859-9":          {charmap.Windows1254, "windows-1254", defaultLength},
	"iso-ir-148":          {charmap.Windows1254, "windows-1254", defaultLength},
	"iso8859-9":           {charmap.Windows1254, "windows-1254", defaultLength},
	"iso88599":            {charmap.Windows1254, "windows-1254", defaultLength},
	"iso_8859-9":          {charmap.Windows1254, "windows-1254", defaultLength},
	"iso_8859-9:1989":     {charmap.Windows1254, "windows-1254", defaultLength},
	"l5":                  {charmap.Windows1254, "windows-1254", defaultLength},
	"latin5":              {charmap.Windows1254, "windows-1254", defaultLength},
	"windows-1254":        {charmap.Windows1254, "windows-1254", defaultLength},
	"x-cp1254":            {charmap.Windows1254, "windows-1254", defaultLength},
	"cp1255":              {charmap.Windows1255, "windows-1255", defaultLength},
	"windows-1255":        {charmap.Windows1255, "windows-1255", defaultLength},
	"x-cp1255":            {charmap.Windows1255, "windows-1255", defaultLength},
	"cp1256":              {charmap.Windows1256, "windows-1256", defaultLength},
	"windows-1256":        {charmap.Windows1256, "windows-1256", defaultLength},
	"x-cp1256":            {charmap.Windows1256, "windows-1256", defaultLength},
	"cp1257":              {charmap.Windows1257, "windows-1257", defaultLength},
	"windows-1257":        {charmap.Windows1257, "windows-1257", defaultLength},
	"x-cp1257":            {charmap.Windows1257, "windows-1257", defaultLength},
	"cp1258":              {charmap.Windows1258, "windows-1258", defaultLength},
	"windows-1258":        {charmap.Windows1258, "windows-1258", defaultLength},
	"x-cp1258":            {charmap.Windows1258, "windows-1258", defaultLength},
	"x-mac-cyrillic":      {charmap.MacintoshCyrillic, "x-mac-cyrillic", defaultLength},
	"x-mac-ukrainian":     {charmap.MacintoshCyrillic, "x-mac-cyrillic", defaultLength},
	"chinese":             {simplifiedchinese.GBK, "gbk", gbkLength},
	"csgb2312":            {simplifiedchinese.GBK, "gbk", gbkLength},
	"csiso58gb231280":     {simplifiedchinese.GBK, "gbk", gbkLength},
	"gb2312":              {simplifiedchinese.GBK, "gbk", gbkLength},
	"gb_2312":             {simplifiedchinese.GBK, "gbk", gbkLength},
	"gb_2312-80":          {simplifiedchinese.GBK, "gbk", gbkLength},
	"gbk":                 {simplifiedchinese.GBK, "gbk", gbkLength},
	"iso-ir-58":           {simplifiedchinese.GBK, "gbk", gbkLength},
	"x-gbk":               {simplifiedchinese.GBK, "gbk", gbkLength},
	"gb18030":             {simplifiedchinese.GB18030, "gb18030", defaultLength},
	"hz-gb-2312":          {simplifiedchinese.HZGB2312, "hz-gb-2312", defaultLength},
	"big5":                {traditionalchinese.Big5, "big5", defaultLength},
	"big5-hkscs":          {traditionalchinese.Big5, "big5", defaultLength},
	"cn-big5":             {traditionalchinese.Big5, "big5", defaultLength},
	"csbig5":              {traditionalchinese.Big5, "big5", defaultLength},
	"x-x-big5":            {traditionalchinese.Big5, "big5", defaultLength},
	"cseucpkdfmtjapanese": {japanese.EUCJP, "euc-jp", defaultLength},
	"euc-jp":              {japanese.EUCJP, "euc-jp", defaultLength},
	"x-euc-jp":            {japanese.EUCJP, "euc-jp", defaultLength},
	"csiso2022jp":         {japanese.ISO2022JP, "iso-2022-jp", defaultLength},
	"iso-2022-jp":         {japanese.ISO2022JP, "iso-2022-jp", defaultLength},
	"csshiftjis":          {japanese.ShiftJIS, "shift_jis", defaultLength},
	"ms_kanji":            {japanese.ShiftJIS, "shift_jis", defaultLength},
	"shift-jis":           {japanese.ShiftJIS, "shift_jis", defaultLength},
	"shift_jis":           {japanese.ShiftJIS, "shift_jis", defaultLength},
	"sjis":                {japanese.ShiftJIS, "shift_jis", defaultLength},
	"windows-31j":         {japanese.ShiftJIS, "shift_jis", defaultLength},
	"x-sjis":              {japanese.ShiftJIS, "shift_jis", defaultLength},
	"cseuckr":             {korean.EUCKR, "euc-kr", defaultLength},
	"csksc56011987":       {korean.EUCKR, "euc-kr", defaultLength},
	"euc-kr":              {korean.EUCKR, "euc-kr", defaultLength},
	"iso-ir-149":          {korean.EUCKR, "euc-kr", defaultLength},
	"korean":              {korean.EUCKR, "euc-kr", defaultLength},
	"ks_c_5601-1987":      {korean.EUCKR, "euc-kr", defaultLength},
	"ks_c_5601-1989":      {korean.EUCKR, "euc-kr", defaultLength},
	"ksc5601":             {korean.EUCKR, "euc-kr", defaultLength},
	"ksc_5601":            {korean.EUCKR, "euc-kr", defaultLength},
	"windows-949":         {korean.EUCKR, "euc-kr", defaultLength},
	"csiso2022kr":         {encoding.Replacement, "replacement", defaultLength},
	"iso-2022-kr":         {encoding.Replacement, "replacement", defaultLength},
	"iso-2022-cn":         {encoding.Replacement, "replacement", defaultLength},
	"iso-2022-cn-ext":     {encoding.Replacement, "replacement", defaultLength},
	"utf-16be":            {unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM), "utf-16be", defaultLength},
	"utf-16":              {unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), "utf-16le", defaultLength},
	"utf-16le":            {unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), "utf-16le", defaultLength},
	"x-user-defined":      {charmap.XUserDefined, "x-user-defined", defaultLength},
}

func defaultLength(_ []byte) int {
	return 4
}

func gbkLength(bytes []byte) (length int) {
	if len(bytes) == 0 {
		return 1
	}
	if bytes[0] < 0x80 {
		return 1
	}
	return 2
}
