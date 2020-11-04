/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astbuilder

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// TextLiteral() creates the AST node for a literal text value
// No additional text is included
func TextLiteral(content string) *ast.BasicLit {
	return &ast.BasicLit{
		Value: content,
		Kind:  token.STRING,
	}
}

// TestLiteralf() creates the AST node for literal text based on a format string
func TestLiteralf(format string, a ...interface{}) *ast.BasicLit {
	return TextLiteral(fmt.Sprintf(format, a...))
}

// StringLiteral() creates the AST node for a literal string value
// Leading and trailing quotes are added as required and any existing quotes are escaped
func StringLiteral(content string) *ast.BasicLit {
	// Pay attention to the string escaping here!
	c := "\"" + strings.ReplaceAll(content, "\"", "\\\"") + "\""
	return TextLiteral(c)
}

// StringLiteralf() creates the AST node for a literal string value based on a format string
// Leading and trailing quotes are added as required and any existing quotes are escaped
func StringLiteralf(format string, a ...interface{}) *ast.BasicLit {
	return StringLiteral(fmt.Sprintf(format, a...))
}
