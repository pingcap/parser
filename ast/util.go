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

import (
	"strings"
)

type NodeTextCleaner struct {
}

// Enter implements Visitor interface.
func (checker *NodeTextCleaner) Enter(in Node) (out Node, skipChildren bool) {
	in.SetText("")
	return in, false
}

// Leave implements Visitor interface.
func (checker *NodeTextCleaner) Leave(in Node) (out Node, ok bool) {
	return in, true
}

func WriteName(sb *strings.Builder, name string) {
	sb.WriteString("`")
	sb.WriteString(EscapeName(name))
	sb.WriteString("`")
}

func EscapeName(name string) string {
	return strings.Replace(name, "`", "``", -1)
}
