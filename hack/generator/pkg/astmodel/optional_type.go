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
func (optional *OptionalType) AsType(codeGenerationContext *CodeGenerationContext) ast.Expr {
	// Special case interface{} as it shouldn't be a pointer
	if optional.element == AnyType {
		return optional.element.AsType(codeGenerationContext)
	}

	return &ast.StarExpr{
		X: optional.element.AsType(codeGenerationContext),
	}
}

// RequiredImports returns the imports required by the 'element' type
func (optional *OptionalType) RequiredImports() []*PackageReference {
	return optional.element.RequiredImports()
}

// References is true if it is this type or the 'element' type references it
func (optional *OptionalType) References(d *TypeName) bool {
	return optional.element.References(d)
}

// CreateInternalDefinitions invokes CreateInternalDefinitions on the inner type
func (optional *OptionalType) CreateInternalDefinitions(name *TypeName, idFactory IdentifierFactory) (Type, []TypeDefiner) {
	newElementType, otherTypes := optional.element.CreateInternalDefinitions(name, idFactory)
	return NewOptionalType(newElementType), otherTypes
}

// CreateDefinitions defines a named type for this OptionalType
func (optional *OptionalType) CreateDefinitions(name *TypeName, _ IdentifierFactory, _ bool) (TypeDefiner, []TypeDefiner) {
	return NewSimpleTypeDefiner(name, optional), nil
}
