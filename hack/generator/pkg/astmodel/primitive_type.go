/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"go/ast"
)

// PrimitiveType represents a Go primitive type
type PrimitiveType struct {
	name string
}

// IntType represents a Go integer type
var IntType = &PrimitiveType{"int"}

// StringType represents the Go string type
var StringType = &PrimitiveType{"string"}

// FloatType represents the Go float64 type
var FloatType = &PrimitiveType{"float64"}

// BoolType represents the Go bool type
var BoolType = &PrimitiveType{"bool"}

// AnyType represents the root Go interface type, permitting any object
var AnyType = &PrimitiveType{"interface{}"}

// assert that we implemented Type correctly
var _ Type = (*PrimitiveType)(nil)

// AsType implements Type for PrimitiveType returning an abstract syntax tree
func (prim *PrimitiveType) AsType() ast.Expr {
	return ast.NewIdent(prim.name)
}

// RequiredImports returns a list of package required by this
func (prim *PrimitiveType) RequiredImports() []PackageReference {
	return nil
}

// References this type has to the given type
func (prim *PrimitiveType) References(d *DefinitionName) bool {
	// Primitive types dont have references
	return false
}

// Equals returns true if the passed type is another primitive type the same name, false otherwise
func (prim *PrimitiveType) Equals(t Type) bool {
	if p, ok := t.(*PrimitiveType); ok {
		return prim.name == p.name
	}

	return false
}

func (prim *PrimitiveType) MakeDefiner(name *DefinitionName, idFactory IdentifierFactory) TypeDefiner {
	return &SimpleTypeDefiner{name, prim}
}
