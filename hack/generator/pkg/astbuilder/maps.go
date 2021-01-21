package astbuilder

import (
	"github.com/dave/dst"
	"go/token"
)

// MakeMap returns the call expression for making a map, like:
// 	make(map[<key>]<value>)
func MakeMap(key dst.Expr, value dst.Expr) *dst.CallExpr {
	return &dst.CallExpr{
		Fun: dst.NewIdent("make"),
		Args: []dst.Expr{
			&dst.MapType{
				Key:   dst.Clone(key).(dst.Expr),
				Value: dst.Clone(value).(dst.Expr),
			},
		},
	}
}

// InsertMap returns an assignment statement for inserting an item into a map, like:
// 	<m>[<key>] = <rhs>
func InsertMap(m dst.Expr, key dst.Expr, rhs dst.Expr) *dst.AssignStmt {
	return SimpleAssignment(
		&dst.IndexExpr{
			X:     dst.Clone(m).(dst.Expr),
			Index: dst.Clone(key).(dst.Expr),
		},
		token.ASSIGN,
		dst.Clone(rhs).(dst.Expr))
}
