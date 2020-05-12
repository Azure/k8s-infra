/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import "go/ast"

type DefinitionName struct {
	PackageReference
	name string
}

func NewDefinitionName(pr PackageReference, name string) DefinitionName {
	return DefinitionName{pr, name}
}

func (name *DefinitionName) Name() string {
	return name.name
}

// A DefinitionName can be used as a Type,
// it is simply a reference to the name.
var _ Type = (*DefinitionName)(nil)

// AsType implements Type for DefinitionName
func (dn *DefinitionName) AsType() ast.Expr {
	return ast.NewIdent(dn.name)
}

func (dn *DefinitionName) References(t Type) bool {
	return dn.Equals(t)
}

func (dn *DefinitionName) RequiredImports() []PackageReference {
	return []PackageReference{dn.PackageReference}
}

// Equals returns true if the passed type is references the same definition, false otherwise
func (dn *DefinitionName) Equals(t Type) bool {
	if d, ok := t.(*DefinitionName); ok {
		return dn.name == d.name && dn.PackageReference.Equals(&d.PackageReference)
	}

	return false
}
