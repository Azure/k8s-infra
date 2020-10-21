package astbuilder

import (
	"go/ast"
	"go/token"
)

// SimpleAssignment performs a simple assignment like:
// 	<lhs> <tok> <rhs>
func SimpleAssignment(lhs ast.Expr, tok token.Token, rhs ast.Expr) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{
			lhs,
		},
		Tok: tok,
		Rhs: []ast.Expr{
			rhs,
		},
	}
}

// SimpleAssignment performs a simple assignment like:
// 	<lhs>, err <tok> <rhs>
func SimpleAssignmentWithErr(lhs ast.Expr, tok token.Token, rhs ast.Expr) *ast.AssignStmt {
	errId := ast.NewIdent("err")
	return &ast.AssignStmt{
		Lhs: []ast.Expr{
			lhs,
			errId,
		},
		Tok: tok,
		Rhs: []ast.Expr{
			rhs,
		},
	}
}
