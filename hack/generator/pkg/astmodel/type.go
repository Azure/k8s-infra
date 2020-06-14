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
	RequiredImports() []*PackageReference

	// References determines if this type has a direct reference to the given definition name
	// For example an Array of Persons references a Person
	References(d *TypeName) bool

	// AsTypeAst renders a Go abstract syntax tree for referencing the type.
	// For simple types this is just the type name. For named types, it's just the name of the type.
	// For more complex types (e.g. maps) this is the appropriate declaration.
	AsTypeAst(codeGenerationContext *CodeGenerationContext) ast.Expr

	// AsTypeDeclarations renders as Go abstract syntax trees for each required declaration
	// Returns the type declaration for this specific type, plus a slice of other supporting declarations
	// (yes this says ast.Expr but that is what the Go 'ast' package uses for types)
	AsTypeDeclarations(codeGenerationContext *CodeGenerationContext) (ast.Expr, []ast.Expr)

	// Equals returns true if the passed type is the same as this one, false otherwise
	Equals(t Type) bool

	// CreateNamedTypes gives a name to the type and might generate some associated definitions as well (the second result)
	// that must also be included in the output.
	//
	// isResource is only relevant to struct types and identifies if they are a root resource for Kubebuilder
	CreateNamedTypes(name *TypeName, idFactory IdentifierFactory, isResource bool) (*NamedType, []*NamedType)

	// CreateInternalDefinitions creates definitions for nested types where needed (e.g. nested anonymous enums, structs),
	// and returns the new, updated type to use in this type’s place.
	CreateInternalDefinitions(nameHint *TypeName, idFactory IdentifierFactory) (Type, []*NamedType)

	// Visit allows a function to walk the tree of all the types nested within a type
	Visit(visitor func(t Type))
}
