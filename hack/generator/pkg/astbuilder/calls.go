package astbuilder

import "go/ast"

func CallFunc(funcName *ast.Ident, arguments ...ast.Expr) ast.Expr {
	return &ast.CallExpr{
		Fun:  funcName,
		Args: arguments,
	}
}

func InvokeFunc(funcName *ast.Ident, arguments ...ast.Expr) ast.Stmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun:  funcName,
			Args: arguments,
		},
	}
}

func CallFuncByName(funcName string, arguments ...ast.Expr) ast.Expr {
	return CallFunc(ast.NewIdent(funcName), arguments...)
}

func CallMethod(receiver *ast.Ident, method *ast.Ident, arguments ...ast.Expr) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   receiver,
			Sel: method,
		},
		Args: arguments,
	}
}

func CallMethodByName(receiver string, method string, arguments ...ast.Expr) ast.Expr {
	return CallMethod(ast.NewIdent(receiver), ast.NewIdent(method), arguments...)
}
