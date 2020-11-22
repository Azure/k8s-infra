/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astbuilder

import (
	ast "github.com/dave/dst"
)

// CallFunc() creates an expression to call a function with specified arguments, generating code
// like:
// <funcName>(<arguments>...)
func CallFunc(funcName ast.Expr, arguments ...ast.Expr) ast.Expr {
	return &ast.CallExpr{
		Fun:  funcName,
		Args: arguments,
	}
}

// CallFuncByName() creates an expression to call a function of the specified name with the
// given arguments, generating code like:
// <funcName>(<arguments>...)
func CallFuncByName(funcName string, arguments ...ast.Expr) ast.Expr {
	return CallFunc(ast.NewIdent(funcName), arguments...)
}

// CallQualifiedFunc() creates an expression to call a qualified function with the specified
// arguments, generating code like:
// <qualifier>.<funcName>(arguments...)
func CallQualifiedFunc(qualifier string, funcName string, arguments ...ast.Expr) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(qualifier),
			Sel: ast.NewIdent(funcName),
		},
		Args: arguments,
	}
}

// InvokeFunc() creates a statement to invoke a function with specified arguments, generating code
// like
// <funcName>(arguments...)
// If you want to use the result of the function call as a value, use CallFunc() instead
func InvokeFunc(funcName *ast.Ident, arguments ...ast.Expr) ast.Stmt {
	return &ast.ExprStmt{
		X: CallFunc(funcName, arguments...),
	}
}

// InvokeQualifiedFunc() creates a statement to invoke a qualified function with specified
// arguments, generating code like:
// <qualifier>.<funcName>(arguments...)
// If you want to use the result of the function call as a value, use CallQualifiedFunc() instead
func InvokeQualifiedFunc(qualifier string, funcName string, arguments ...ast.Expr) ast.Stmt {
	return &ast.ExprStmt{
		X: CallQualifiedFunc(qualifier, funcName, arguments...),
	}
}

// InvokeQualifiedFuncByName() creates a statement to invoke a qualified function of the specified
// name with the given arguments, generating code like:
// <qualifier>.<funcName>(arguments...)
// If you want to use the result of the function call as a value, use CallQualifiedFuncByName() instead
func InvokeQualifiedFuncByName(qualifier string, funcName string, arguments ...ast.Expr) ast.Stmt {
	return InvokeQualifiedFunc(qualifier, funcName, arguments...)
}
