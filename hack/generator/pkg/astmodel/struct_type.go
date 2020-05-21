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

// Ensure StructType implements the Type interface correctly
var _ Type = (*StructType)(nil)

// NewStructType is a factory method for creating a new StructTypeDefinition
func NewStructType(fields ...*FieldDefinition) *StructType {
	return &StructType{fields}
}

// Fields returns all our field definitions
// A copy of the slice is returned to preserve immutability
func (structType *StructType) Fields() []*FieldDefinition {
	return append(structType.fields[:0:0], structType.fields...)
}

// AsType implements Type for StructType
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

// RequiredImports returns a list of packages required by this
func (structType *StructType) RequiredImports() []PackageReference {
	var result []PackageReference
	for _, field := range structType.fields {
		result = append(result, field.FieldType().RequiredImports()...)
	}

	return result
}

// References this type has to the given type
func (structType *StructType) References(d *TypeName) bool {
	for _, field := range structType.fields {
		if field.FieldType().References(d) {
			return true
		}
	}

	return false
}

// Equals returns true if the passed type is a struct type with the same fields, false otherwise
// The order of the fields is not relevant
func (structType *StructType) Equals(t Type) bool {
	if structType == t {
		return true
	}

	if st, ok := t.(*StructType); ok {
		if len(structType.fields) != len(st.fields) {
			// Different number of fields, not equal
			return false
		}

		ourFields := make(map[FieldName]*FieldDefinition)
		for _, f := range structType.fields {
			ourFields[f.fieldName] = f
		}

		for _, f := range st.fields {
			ourfield, ok := ourFields[f.fieldName]
			if !ok {
				// Didn't find the field, not equal
				return false
			}

			if !ourfield.Equals(f) {
				// Different field, even though same name; not-equal
				return false
			}
		}

		// All fields match, equal
		return true
	}

	return false
}

func (st *StructType) CreateDefinitions(name *TypeName, idFactory IdentifierFactory, isResource bool) (TypeDefiner, []TypeDefiner) {

	ref := NewStructReference(name.name, name.groupName, name.packageName, isResource)

	var otherTypes []TypeDefiner
	var newFields []*FieldDefinition

	for _, field := range st.fields {
		newField := field

		// TODO: figure out a generic way to do this:
		if et, ok := newField.FieldType().(*EnumType); ok {
			// enums that are not explicitly refs get named here:
			enumName := name.name + string(field.fieldName)
			defName := NewTypeName(name.PackageReference, enumName)
			ed, edOther := et.CreateDefinitions(&defName, idFactory, false)

			// append all definitions into output
			otherTypes = append(append(otherTypes, ed), edOther...)

			newField = NewFieldDefinition(newField.fieldName, newField.jsonName, ed.Name())
		} else if st, ok := newField.FieldType().(*StructType); ok {
			// inline structs get named here:

			structName := name.name + string(field.fieldName)
			defName := NewTypeName(name.PackageReference, structName)
			sd, sdOther := st.CreateDefinitions(&defName, idFactory, false) // nested types are never resources

			// append all definitions into output
			otherTypes = append(append(otherTypes, sd), sdOther...)

			newField = NewFieldDefinition(newField.fieldName, newField.jsonName, sd.Name())
		}

		newFields = append(newFields, newField)
	}

	return NewStructDefinition(ref, newFields...), otherTypes
}
