/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"go/ast"
)

// OptionalType is used for items that may or may not be present
type OptionalType struct {
	element Type
}

// NewOptionalType creates a new optional type that may or may not have the specified 'element' type
func NewOptionalType(element Type) *OptionalType {
	return &OptionalType{element}
}

// assert we implemented Type correctly
var _ Type = (*OptionalType)(nil)

// AsType renders the Go abstract syntax tree for an optional type
func (optional *OptionalType) AsType() ast.Expr {
	return &ast.StarExpr{
		X: optional.element.AsType(),
	}
}

func (optional *OptionalType) RequiredImports() []PackageReference {
	return optional.element.RequiredImports()
}

func (optional *OptionalType) References(t Type) bool {
	return optional == t || optional.element.References(t)
}
