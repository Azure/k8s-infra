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

// AsTypeAst renders a Go abstract syntax tree for referencing the type.
func (optional *OptionalType) AsTypeAst(codeGenerationContext *CodeGenerationContext) ast.Expr {
	elementType := optional.element.AsTypeAst(codeGenerationContext)

	// Special case interface{} as it shouldn't be a pointer
	if optional.element == AnyType {
		return elementType
	}

	return &ast.StarExpr{
		X: elementType,
	}
}

// AsDeclarationAsts renders the Go abstract syntax tree for an optional type
// Optional types don't have any declarations of their own (but their underlying type might)
func (optional *OptionalType) AsDeclarationAsts(nameHint string, codeGenerationContext *CodeGenerationContext) []ast.Decl {
	return optional.element.AsDeclarationAsts(nameHint, codeGenerationContext)
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
