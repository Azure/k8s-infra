/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astbuilder

import "go/ast"

// CallFunc() creates an expression to call a function with specified arguments
func CallFunc(funcName *ast.Ident, arguments ...ast.Expr) ast.Expr {
	return &ast.CallExpr{
		Fun:  funcName,
		Args: arguments,
	}
}

// CallFuncByName() generates an expression to call a function of the specified name with the
// given arguments
func CallFuncByName(funcName string, arguments ...ast.Expr) ast.Expr {
	return CallFunc(ast.NewIdent(funcName), arguments...)
}

// CallQualifiedFunc() generates an expression to call a qualified function with the specified
// arguments
func CallQualifiedFunc(qualifier *ast.Ident, method *ast.Ident, arguments ...ast.Expr) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   qualifier,
			Sel: method,
		},
		Args: arguments,
	}
}

// CallQualifiedFuncByName() generates an expression to call a qualified function of the specified
// name with the given arguments
func CallQualifiedFuncByName(receiver string, method string, arguments ...ast.Expr) ast.Expr {
	return CallQualifiedFunc(ast.NewIdent(receiver), ast.NewIdent(method), arguments...)
}

// InvokeFunc() creates a statement to invoke a function with specified arguments
// If you want to use the result of the function call as a value, use CallFunc() instead
func InvokeFunc(funcName *ast.Ident, arguments ...ast.Expr) ast.Stmt {
	return &ast.ExprStmt{
		X: CallFunc(funcName, arguments...),
	}
}

// InvokeQualifiedFunc() creates a statement to invoke a qualified function with specified arguments
// If you want to use the result of the function call as a value, use CallQualifiedFunc() instead
func InvokeQualifiedFunc(qualifier *ast.Ident, method *ast.Ident, arguments ...ast.Expr) ast.Stmt {
	return &ast.ExprStmt{
		X: CallQualifiedFunc(qualifier, method, arguments...),
	}
}

// InvokeQualifiedFuncByName() creates a statement to invoke a qualified function of the specified
// name with the given arguments
// If you want to use the result of the function call as a value, use CallQualifiedFuncByName() instead
func InvokeQualifiedFuncByName(qualifier string, method string, arguments ...ast.Expr) ast.Stmt {
	return InvokeQualifiedFunc(ast.NewIdent(qualifier), ast.NewIdent(method), arguments...)
}
