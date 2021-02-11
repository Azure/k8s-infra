/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astbuilder

import (
	"go/token"

	"github.com/dave/dst"
)

// SimpleAssignment performs a simple assignment like:
//     <lhs> := <rhs>       // tok = token.DEFINE
// or  <lhs> = <rhs>        // tok = token.ASSIGN
func SimpleAssignment(lhs dst.Expr, tok token.Token, rhs dst.Expr) *dst.AssignStmt {
	return &dst.AssignStmt{
		Lhs: []dst.Expr{
			dst.Clone(lhs).(dst.Expr),
		},
		Tok: tok,
		Rhs: []dst.Expr{
			dst.Clone(rhs).(dst.Expr),
		},
	}
}

// SimpleAssignmentWithErr performs a simple assignment like:
// 	    <lhs>, err := <rhs>       // tok = token.DEFINE
// 	or  <lhs>, err = <rhs>        // tok = token.ASSIGN
func SimpleAssignmentWithErr(lhs dst.Expr, tok token.Token, rhs dst.Expr) *dst.AssignStmt {
	errId := dst.NewIdent("err")
	return &dst.AssignStmt{
		Lhs: []dst.Expr{
			dst.Clone(lhs).(dst.Expr),
			errId,
		},
		Tok: tok,
		Rhs: []dst.Expr{
			dst.Clone(rhs).(dst.Expr),
		},
	}
}

// SimpleIf creates a simple if statement with a single statement in each branch
//
// if <condition> {
//     <trueBranch>
// } else {
//     <falseBranch>
// }
//
func SimpleIf(condition dst.Expr, trueBranch dst.Stmt, falseBranch dst.Stmt) *dst.IfStmt {
	return &dst.IfStmt{
		Cond: condition,
		Body: StatementBlock(trueBranch),
		Else: StatementBlock(falseBranch),
	}
}

// IfNotNil executes a series of statements if the supplied expression is not nil
//
// if <source> != nil {
//     <statements>
// }
//
func IfNotNil(toCheck dst.Expr, statements ...dst.Stmt) *dst.IfStmt {
	return &dst.IfStmt{
		Cond: &dst.BinaryExpr{
			X:  dst.Clone(toCheck).(dst.Expr),
			Op: token.NEQ,
			Y:  dst.NewIdent("nil"),
		},
		Body: StatementBlock(statements...),
	}
}
