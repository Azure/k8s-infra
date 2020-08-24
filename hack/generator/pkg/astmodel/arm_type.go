/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"
	"go/ast"
)

// ArmType wraps an existing type to indicate that it is an ARM targetted variation
type ArmType struct {
	objectType ObjectType
}

// ArmType is a Type
var _ Type = &ArmType{}

// MakeArmType wraps an object type to indicate it's an ARM focussed variation
func MakeArmType(object ObjectType) ArmType {
	return ArmType{
		objectType: object,
	}
}

// RequiredImports returns a list of packages required by this type
func (at ArmType) RequiredImports() []PackageReference {
	return at.objectType.RequiredImports()
}

// References returns the names of all types that this type references
func (at ArmType) References() TypeNameSet {
	return at.objectType.References()
}

// AsType renders as a Go abstract syntax tree for a type
func (at ArmType) AsType(codeGenerationContext *CodeGenerationContext) ast.Expr {
	return at.objectType.AsType(codeGenerationContext)
}

// AsDeclarations renders as a Go abstract syntax tree for a declaration
func (at ArmType) AsDeclarations(codeGenerationContext *CodeGenerationContext, name TypeName, description []string) []ast.Decl {
	return at.objectType.AsDeclarations(codeGenerationContext, name, description)
}

// Equals decides if the types are the same
func (at ArmType) Equals(t Type) bool {
	if ost, ok := t.(*ArmType); ok {
		return TypeEquals(&at.objectType, ost)
	}

	return false
}

// String returns a string representation of our ARM type
func (at ArmType) String() string {
	return fmt.Sprintf("ARM(%v)", at.objectType)
}

// ObjectType returns the underlying object type of the arm type
func (at ArmType) ObjectType() ObjectType {
	return at.objectType
}

// IsArmType returns true if the passed type is a Arm type; false otherwise.
// We need to handle cases where it might be wrapped, so this isn't just a simple type check, we need to walk the top
// of the type tree to do the check.
func IsArmType(t Type) bool {
	_, ok := t.(ArmType)
	return ok
}

// IsArmDefinition returns true if the passed definition is for a Arm type; false otherwise.
func IsArmDefinition(definition TypeDefinition) bool {
	return IsArmType(definition.theType)
}
