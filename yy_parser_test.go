package parser

import (
	"fmt"
	"github.com/pingcap/parser/ast"
	"log"
	"testing"
)

func TestName1(t *testing.T) {

	parser := New()

	sql := ` admin show test.t next_row_id; `
	stmt, _, err := parser.Parse(sql, ``, ``)
	log.Println(err)
	log.Println(stmt)

}

func TestName12(t *testing.T) {

	parser := New()

	sql := ` show table test.t next_row_id; `
	stmt, _, err := parser.Parse(sql, ``, ``)
	log.Println(err)
	log.Println(stmt)

}

type visitor struct{}

func (v *visitor) Enter(in ast.Node) (out ast.Node, skipChildren bool) {
	fmt.Printf("%T ", in)

	fmt.Printf("%s ", in.Text())

	fmt.Printf("\n")
	return in, false
}

func (v *visitor) Leave(in ast.Node) (out ast.Node, ok bool) {
	return in, true
}

func TestName3(t *testing.T) {

	sql := "SELECT /*+ TIDB_SMJ(employees) */ emp_no, first_name, last_name " +
		"FROM employees USE INDEX (last_name) " +
		"where last_name='Aamodt' and gender='F' and birth_date > '1960-01-01'"

	sql = `
 show test.t next_row_id; 
`

	sql = `
admin show test.t next_row_id; 
`

	/*
	*ast.AlterTableStmt
	*ast.TableName
	*ast.AlterTableSpec
	*ast.ColumnDef
	*ast.ColumnName
	*ast.ColumnOption
	*test_driver.ValueExpr
	*ast.ColumnOption
	*ast.ColumnOption
	*test_driver.ValueExpr
	*ast.ColumnPosition

	 */
	sql = `
alter table start_game_take_time_log_201909
    add server_serial_id tinyint default 0 not null comment '进程id';
`

	sqlParser := New()
	stmtNodes, _, err := sqlParser.Parse(sql, "", "")
	if err != nil {
		fmt.Printf("parse error:\n%v\n%s", err, sql)
		return
	}
	for _, stmtNode := range stmtNodes {
		v := visitor{}
		stmtNode.Accept(&v)

		log.Println()
	}

}
