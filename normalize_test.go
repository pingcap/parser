package parser_test

import (
	"github.com/pingcap/parser"
	"testing"
)

func TestDigestEqForSimpleSQL(t *testing.T) {
	sqlGroups := [][]string{
		{"select * from b where id = 1", "select * from b where id = '1'", "select * from b where id =2"},
		{"select 2 from b, c where b.id =          c.id where c.id > 1", "select 4 from b, c where " +
			"b.id = c.id where c.id > 23"},
	}
	for _, sqlGroup := range sqlGroups {
		var d string
		for _, sql := range sqlGroup {
			dig := parser.Digest(sql)
			if d == "" {
				d = dig
				continue
			}
			if d != dig {
				t.Errorf("digest for %s's digest result %s not eq to previous %s", sql, d, dig)
			}
		}
	}
}

func TestDigestNotEqForSimpleSQL(t *testing.T) {
	sqlGroups := [][]string{
		{"select * from b where id = 1", "select a from b where id = 1", "select * from d where bid =1"},
	}
	for _, sqlGroup := range sqlGroups {
		var d string
		for _, sql := range sqlGroup {
			dig := parser.Digest(sql)
			if d == "" {
				d = dig
				continue
			}
			if d == dig {
				t.Errorf("digest for %s's digest result %s not eq to previous %s", sql, d, dig)
			}
		}
	}
}
