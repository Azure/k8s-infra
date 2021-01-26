/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astbuilder

import (
	"github.com/dave/dst"
	"go/token"
)

// MakeList returns the call expression for making a slice, like:
// 	make([]<value>)
func MakeList(listType dst.Expr, len dst.Expr) *dst.CallExpr {
	return &dst.CallExpr{
		Fun: dst.NewIdent("make"),
		Args: []dst.Expr{
			listType,
			len,
		},
	}
}

// AppendList returns a statement for a list append, like:
//     <lhs> = append(<lhs>, <rhs>)
func AppendList(lhs dst.Expr, rhs dst.Expr) dst.Stmt {
	return SimpleAssignment(
		dst.Clone(lhs).(dst.Expr),
		token.ASSIGN,
		CallFunc("append", dst.Clone(lhs).(dst.Expr), dst.Clone(rhs).(dst.Expr)))
}

func IterateOverList(item string, list dst.Expr, statements ...dst.Stmt) *dst.RangeStmt {
	return &dst.RangeStmt{
		Key:   dst.NewIdent("_"),
		Value: dst.NewIdent(item),
		Tok:   token.DEFINE,
		X:     list,
		Body: &dst.BlockStmt{
			List: statements,
		},
	}
}

func IterateOverListWithIndex(index string, item string, list dst.Expr, statements ...dst.Stmt) *dst.RangeStmt {
	return &dst.RangeStmt{
		Key:   dst.NewIdent(index),
		Value: dst.NewIdent(item),
		Tok:   token.DEFINE,
		X:     list,
		Body: &dst.BlockStmt{
			List: cloneStmtSlice(statements),
		},
	}
}
