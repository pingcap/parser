package ast

import "strings"

type SQLSentence struct {
	strings.Builder
}

func WriteName(sb *strings.Builder, name string) {
	sb.WriteString("`")
	sb.WriteString(EscapeName(name))
	sb.WriteString("`")
}

func EscapeName(name string) string {
	return strings.Replace(name, "`", "``", -1)
}
