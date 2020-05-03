/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"go/ast"
	"sort"
)

// StructType represents an (unnamed) struct type
type StructType struct {
	fields []*FieldDefinition
}

// NewStructType is a factory method for creating a new StructTypeDefinition
func NewStructType(fields []*FieldDefinition) *StructType {
	return &StructType{fields}
}

// Fields returns all our field definitions
// A copy of the slice is returned to preserve immutability
func (structType *StructType) Fields() []*FieldDefinition {
	return append(structType.fields[:0:0], structType.fields...)
}

// AsType implements Type for StructType
var _ Type = (*StructType)(nil)

func (structType *StructType) AsType() ast.Expr {

	// Copy the slice of fields and sort it
	fields := structType.Fields()
	sort.Slice(fields, func(i int, j int) bool {
		return fields[i].fieldName < fields[j].fieldName
	})

	fieldDefinitions := make([]*ast.Field, len(fields))
	for i, f := range fields {
		fieldDefinitions[i] = f.AsField()
	}

	return &ast.StructType{
		Fields: &ast.FieldList{
			List: fieldDefinitions,
		},
	}
}

func (structType *StructType) RequiredImports() []PackageReference {
	var result []PackageReference
	for _, field := range structType.fields {
		result = append(result, field.FieldType().RequiredImports()...)
	}

	return result
}
