/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astbuilder

import (
	"fmt"
	"go/ast"
	"go/token"
)

// LiteralString() creates the AST node for a literal string value
func LiteralString(content string) *ast.BasicLit {
	return &ast.BasicLit{
		Value: "\"" + content + "\"",
		Kind:  token.STRING,
	}
}

// LiteralStringf() creates the AST node for a literal string value based on a format string
func LiteralStringf(format string, a ...interface{}) *ast.BasicLit {
	return LiteralString(fmt.Sprintf(format, a))
}
