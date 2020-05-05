/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

// EnumType represents a set of mutually exclusive predefined options
type EnumType struct {
	DefinitionName
	// BaseType is the underlying type used to define the values
	BaseType *PrimitiveType
	// Options is the set of all unique values
	Options []string
}

// EnumType must implement the Type interface correctly
var _ Type = (*EnumType)(nil)

// EnumType must implement the DefinitionFactory interface correctly
var _ DefinitionFactory = (*EnumType)(nil)

// NewEnumDefinition defines a new enumeration including the legal values
func NewEnumDefinition(baseType *PrimitiveType, options []string) *EnumType {
	return &EnumType{BaseType: baseType, Options: options}
}

// CreateDefinitions implements the DefinitionFactory interface for EnumType
func (enum *EnumType) CreateDefinitions(ref PackageReference, namehint string, idFactory IdentifierFactory) []Definition {
	var result []Definition

	identifier := idFactory.CreateEnumIdentifier(namehint)

	enum.DefinitionName = DefinitionName{PackageReference: ref, name: identifier}

	return result
}
