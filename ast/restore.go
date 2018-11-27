package ast

import (
	"github.com/pingcap/parser/model"
	"strconv"
)

func (n *CreateTableStmt) Restore() string {
	rs := "CREATE TABLE"
	if n.IfNotExists {
		rs += " IF NOT EXISTS"
	}
	rs += " " + n.Table.Restore()
	if n.ReferTable != nil {
		rs += " LIKE " + n.ReferTable.Restore()
	}
	colsLen := len(n.Cols)
	constraintsLen := len(n.Constraints)
	if colsLen != 0 {
		rs += " ("
		for i, col := range n.Cols {
			rs += col.Restore()
			if i != colsLen-1 {
				rs += ", "
			}
		}
		if colsLen != 0 && constraintsLen != 0 {
			rs += ", "
		}
		for i, constraint := range n.Constraints {
			rs += constraint.Restore()
			if i != constraintsLen-1 {
				rs += ", "
			}
		}
		rs += ")"
	}

	return rs
}

func (n *TableName) Restore() string {
	rs := n.Name.String()
	if n.Schema.String() != "" {
		rs = n.Schema.String() + "." + rs
	}
	return rs
}

func (n *ColumnDef) Restore() string {
	rs := n.Name.Restore() + " " + n.Tp.String()
	for _, option := range n.Options {
		rs += " " + option.Restore()
	}
	return rs
}

func (n *ColumnName) Restore() string {
	rs := n.Name.String()
	if n.Table.String() != "" {
		rs = n.Table.String() + "." + rs
	} else {
		return rs
	}
	if n.Schema.String() != "" {
		rs = n.Schema.String() + "." + rs
	}
	return rs
}

func (n *ColumnOption) Restore() string {
	switch n.Tp {
	case ColumnOptionPrimaryKey:
		return "PRIMARY KEY"
	case ColumnOptionNotNull:
		return "NOT NULL"
	case ColumnOptionAutoIncrement:
		return "AUTO_INCREMENT"
	case ColumnOptionDefaultValue:
		return "DEFAULT 'expr'"
	case ColumnOptionUniqKey:
		return "UNIQUE KEY"
	case ColumnOptionNull:
		return "NULL"
	case ColumnOptionOnUpdate: // For Timestamp and Datetime only.
		return "ON UPDATE 'expr'"
	case ColumnOptionFulltext:
		return "FULLTEXT"
	case ColumnOptionComment:
		return "COMMENT 'expr'"
	case ColumnOptionGenerated:
		stored := "VIRTUAL"
		if n.Stored {
			stored = "STORED"
		}
		return "AS 'expr' " + stored
	case ColumnOptionReference:
		rs := "REFERENCES " + n.Refer.Table.Restore()
		colsLen := len(n.Refer.IndexColNames)
		if colsLen != 0 {
			rs += "("
			for i, col := range n.Refer.IndexColNames {
				rs += col.Column.Restore()
				if i != colsLen-1 {
					rs += ", "
				}
			}
			rs += ")"
		}
		if n.Refer.OnDelete != nil && n.Refer.OnDelete.ReferOpt != 0 {
			rs += " ON DELETE " + n.Refer.OnDelete.ReferOpt.String()
		}
		if n.Refer.OnUpdate != nil && n.Refer.OnUpdate.ReferOpt != 0 {
			rs += " ON UPDATE " + n.Refer.OnUpdate.ReferOpt.String()
		}
		return rs
	}
	return ""
}

func (n *Constraint) Restore() string {
	switch n.Tp {
	case ConstraintPrimaryKey:
		rs := "CONSTRAINT"
		if n.Name != "" {
			rs += " " + n.Name
		}
		rs += " PRIMARY KEY "
		keysLen := len(n.Keys)
		if keysLen != 0 {
			rs += "("
			for i, key := range n.Keys {
				rs += key.Column.Restore()
				if i != keysLen-1 {
					rs += ", "
				}
			}
			rs += ")"
		}
		if n.Option != nil {
			switch n.Option.Tp {
			case model.IndexTypeBtree, model.IndexTypeHash:
				rs += " USING " + n.Option.Tp.String()
			}
			if len(n.Option.Comment) != 0 {
				rs += " COMMENT '" + n.Option.Comment + "'"
			}
		}
		return rs
	case ConstraintKey:

	case ConstraintIndex:
		rs := "INDEX"
		if n.Name != "" {
			rs += " " + n.Name
		}
		keysLen := len(n.Keys)
		if keysLen != 0 {
			rs += " ("
			for i, key := range n.Keys {
				rs += key.Column.Restore()
				if i != keysLen-1 {
					rs += ", "
				}
			}
			rs += ")"
		}
		if n.Option != nil {
			switch n.Option.Tp {
			case model.IndexTypeBtree, model.IndexTypeHash:
				rs += " USING " + n.Option.Tp.String()
			}
			if len(n.Option.Comment) != 0 {
				rs += " COMMENT '" + n.Option.Comment + "'"
			}
		}
		return rs
	case ConstraintUniq:
		rs := "CONSTRAINT"
		if n.Name != "" {
			rs += " " + n.Name
		}
		rs += " UNIQUE KEY "
		keysLen := len(n.Keys)
		if keysLen != 0 {
			rs += "("
			for i, key := range n.Keys {
				rs += key.Column.Restore()
				if i != keysLen-1 {
					rs += ", "
				}
			}
			rs += ")"
		}
		if n.Option != nil {
			switch n.Option.Tp {
			case model.IndexTypeBtree, model.IndexTypeHash:
				rs += " USING " + n.Option.Tp.String()
			}
			if len(n.Option.Comment) != 0 {
				rs += " COMMENT '" + n.Option.Comment + "'"
			}
		}
		return rs
	case ConstraintUniqKey:
	case ConstraintUniqIndex:
	case ConstraintForeignKey:
		rs := "CONSTRAINT"
		if n.Name != "" {
			rs += " " + n.Name
		}
		rs += " FOREIGN KEY "
		keysLen := len(n.Keys)
		if keysLen != 0 {
			rs += "("
			for i, key := range n.Keys {
				rs += key.Column.Restore()
				if i != keysLen-1 {
					rs += ", "
				}
			}
			rs += ")"
		}
		rs += " REFERENCES " + n.Refer.Table.Restore()
		colsLen := len(n.Refer.IndexColNames)
		if colsLen != 0 {
			rs += "("
			for i, col := range n.Refer.IndexColNames {
				rs += col.Column.Restore()
				if i != colsLen-1 {
					rs += ", "
				}
			}
			rs += ")"
		}
		if n.Refer.OnDelete != nil && n.Refer.OnDelete.ReferOpt != 0 {
			rs += " ON DELETE " + n.Refer.OnDelete.ReferOpt.String()
		}
		if n.Refer.OnUpdate != nil && n.Refer.OnUpdate.ReferOpt != 0 {
			rs += " ON UPDATE " + n.Refer.OnUpdate.ReferOpt.String()
		}
		return rs
	case ConstraintFulltext:
		rs := "FULLTEXT INDEX"
		if n.Name != "" {
			rs += " " + n.Name
		}
		keysLen := len(n.Keys)
		if keysLen != 0 {
			rs += " ("
			for i, key := range n.Keys {
				rs += key.Column.Restore()
				if i != keysLen-1 {
					rs += ", "
				}
			}
			rs += ")"
		}
		if n.Option != nil {
			switch n.Option.Tp {
			case model.IndexTypeBtree, model.IndexTypeHash:
				rs += " USING " + n.Option.Tp.String()
			}
			if len(n.Option.Comment) != 0 {
				rs += " COMMENT '" + n.Option.Comment + "'"
			}
		}
		return rs
	}
	return strconv.Itoa(int(n.Tp))
}
