/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import "go/ast"

// TypeName is a name associated with another Type (it also is usable as a Type)
type TypeName struct {
	PackageReference
	name string
}

// NewTypeName is a factory method for creating a TypeName
func NewTypeName(pr PackageReference, name string) TypeName {
	return TypeName{pr, name}
}

// Name returns the package-local name of the type
func (dn *TypeName) Name() string {
	return dn.name
}

// A TypeName can be used as a Type,
// it is simply a reference to the name.
var _ Type = (*TypeName)(nil)

// AsType implements Type for TypeName
func (dn *TypeName) AsType() ast.Expr {
	return ast.NewIdent(dn.name)
}

// References indicates whether this Type includes any direct references to the given Type
func (dn *TypeName) References(d *TypeName) bool {
	return dn.Equals(d)
}

// RequiredImports returns all the imports required for this definition
func (dn *TypeName) RequiredImports() []PackageReference {
	return []PackageReference{dn.PackageReference}
}

// Equals returns true if the passed type is the same TypeName, false otherwise
func (dn *TypeName) Equals(t Type) bool {
	if d, ok := t.(*TypeName); ok {
		return dn.name == d.name && dn.PackageReference.Equals(&d.PackageReference)
	}

	return false
}

func (dn *TypeName) CreateDefinitions(name *TypeName, idFactory IdentifierFactory) (TypeDefiner, []TypeDefiner) {
	return &SimpleTypeDefiner{name, dn}, nil
}
