/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"go/ast"
)

// Type represents something that is a Go type
type Type interface {
	// RequiredImports returns a list of packages required by this type
	RequiredImports() []PackageReference

	// References determines if this type has a direct reference to the given definition name
	// For example an Array of Persons references a Person
	References(d *TypeName) bool

	// AsType renders as a Go abstract syntax tree for a type
	// (yes this says ast.Expr but that is what the Go 'ast' package uses for types)
	AsType() ast.Expr

	// Equals returns true if the passed type is the same as this one, false otherwise
	Equals(t Type) bool

	// CreateDefinitions gives a name to the type and might generate some asssociated definitions as well (the second result)
	CreateDefinitions(name *TypeName, idFactory IdentifierFactory) (TypeDefiner, []TypeDefiner)
}
