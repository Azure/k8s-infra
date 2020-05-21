/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"go/ast"
)

type TypeDefiner interface {
	Name() *TypeName
	Type() Type

	AsDeclarations() []ast.Decl
}

// FileNameHint returns what a file that contains this definition (if any) should be called
// this is not always used as we might combine multiple definitions into one file
func FileNameHint(def TypeDefiner) string {
	return def.Name().name
}

// Type represents something that is a Go type
type Type interface {
	RequiredImports() []PackageReference

	// ReferenceChecker is used to check for references to a specific definition
	// References determines if this type has a direct reference to the given definition name
	// For example, a struct references its field
	References(d *TypeName) bool

	// AsType renders as a Go abstract syntax tree for a type
	AsType() ast.Expr

	// Equals returns true if the passed type is the same as this one, false otherwise
	Equals(t Type) bool

	CreateDefinitions(name *TypeName, idFactory IdentifierFactory) (TypeDefiner, []TypeDefiner)
}
