/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astbuilder

import (
	"go/ast"
	"go/token"
)

// LiteralString() creates the AST node for a literal string value
func LiteralString(content string) ast.Expr {
	return &ast.BasicLit{
		Value: "\"" + content + "\"",
		Kind:  token.STRING,
	}
}
