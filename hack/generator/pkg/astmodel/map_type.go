/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"go/ast"
)

// MapType is used to define fields that contain additional property values
type MapType struct {
	key   Type
	value Type
}

// NewMapType creates a new map with the specified key and value types
func NewMapType(key Type, value Type) *MapType {
	return &MapType{key, value}
}

// NewStringMapType creates a new map with string keys and the specified value type
func NewStringMapType(value Type) *MapType {
	return NewMapType(StringType, value)
}

// assert that we implemented Type correctly
var _ Type = (*MapType)(nil)

// AsTypeAst renders a Go abstract syntax tree for referencing the map as an anonymous type
func (m *MapType) AsTypeAst(codeGenerationContext *CodeGenerationContext) ast.Expr {
	return &ast.MapType{
		Key:   m.key.AsTypeAst(codeGenerationContext),
		Value: m.value.AsTypeAst(codeGenerationContext),
	}
}

// AsTypeDeclarations implements Type for MapType to create the abstract syntax tree for a map
func (m *MapType) AsTypeDeclarations(codeGenerationContext *CodeGenerationContext) (ast.Expr, []ast.Expr) {
	keyType, otherKeyTypes := m.key.AsTypeDeclarations(codeGenerationContext)
	valueType, otherValueTypes := m.value.AsTypeDeclarations(codeGenerationContext)

	mapType := &ast.MapType{
		Key:   keyType,
		Value: valueType,
	}

	return mapType, append(otherKeyTypes, otherValueTypes...)
}

// RequiredImports returns a list of packages required by this
func (m *MapType) RequiredImports() []*PackageReference {
	var result []*PackageReference
	result = append(result, m.key.RequiredImports()...)
	result = append(result, m.value.RequiredImports()...)
	return result
}

// References this type has to the given type
func (m *MapType) References(d *TypeName) bool {
	return m.key.References(d) || m.value.References(d)
}

// Equals returns true if the passed type is a map type with the same kinds of keys and elements, false otherwise
func (m *MapType) Equals(t Type) bool {
	if m == t {
		return true
	}

	if mt, ok := t.(*MapType); ok {
		return (m.key.Equals(mt.key) && m.value.Equals(mt.value))
	}

	return false
}

// CreateInternalDefinitions invokes CreateInCreateInternalDefinitions on both key and map types
func (m *MapType) CreateInternalDefinitions(name *TypeName, idFactory IdentifierFactory) (Type, []*NamedType) {
	newKeyType, keyOtherTypes := m.key.CreateInternalDefinitions(name, idFactory)
	newValueType, valueOtherTypes := m.value.CreateInternalDefinitions(name, idFactory)
	return NewMapType(newKeyType, newValueType), append(keyOtherTypes, valueOtherTypes...)
}

// CreateNamedTypes defines a named type for this MapType
func (m *MapType) CreateNamedTypes(name *TypeName, _ IdentifierFactory, _ bool) (*NamedType, []*NamedType) {
	return NewNamedType(name, m), nil
}

// Visit the map type, our key type, and our value type
func (m *MapType) Visit(visitor func(t Type)) {
	visitor(m)
	m.key.Visit(visitor)
	m.value.Visit(visitor)
}
