/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astbuilder

import (
	"go/ast"
	"go/token"
)

func LiteralString(content string) ast.Expr {
	return &ast.BasicLit{
		Value: "\"" + content + "\"",
		Kind:  token.STRING,
	}
}
