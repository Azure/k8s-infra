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

// AsTypeReference renders a Go abstract syntax tree for referencing the type.
// For simple types this is just the type name. For named types, it's just the name of the type.
// For more complex types (e.g. maps) this is the appropriate declaration.
func (optional *OptionalType) AsTypeReference(codeGenerationContext *CodeGenerationContext) ast.Expr {
	return &ast.StarExpr{
		X: optional.element.AsTypeReference(codeGenerationContext),
	}
}

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

// Equals returns true if this type is equal to the other type
func (optional *OptionalType) Equals(t Type) bool {
	if optional == t {
		return true // reference equality short-cut
	}

	if otherOptional, ok := t.(*OptionalType); ok {
		return optional.element.Equals(otherOptional.element)
	}

	return false
}

// CreateInternalDefinitions invokes CreateInternalDefinitions on the inner type
func (optional *OptionalType) CreateInternalDefinitions(name *TypeName, idFactory IdentifierFactory) (Type, []*NamedType) {
	newElementType, otherTypes := optional.element.CreateInternalDefinitions(name, idFactory)
	return NewOptionalType(newElementType), otherTypes
}

// CreateNamedTypes defines a named type for this OptionalType
func (optional *OptionalType) CreateNamedTypes(name *TypeName, _ IdentifierFactory, _ bool) (*NamedType, []*NamedType) {
	return NewNamedType(name, optional), nil
}

func (optional *OptionalType) Visit(visitor func(t Type)) {
	visitor(optional)
	optional.element.Visit(visitor)
}
