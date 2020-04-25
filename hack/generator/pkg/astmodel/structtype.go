package astmodel

import "go/ast"

// StructType represents an (unnamed) struct type
type StructType struct {
	fields []*FieldDefinition
}

// NewStructType is a factory method for creating a new StructTypeDefinition
func NewStructType(fields []*FieldDefinition) *StructType {
	return &StructType{fields}
}

//TODO: Don't want to return the original slice, don't want it to be modified
func (structType *StructType) Fields() []*FieldDefinition {
	return structType.fields
}

// AsType implements Type for StructType
func (structType *StructType) AsType() ast.Expr {

	fieldDefinitions := make([]*ast.Field, len(structType.fields))
	for i, f := range structType.fields {
		fieldDefinitions[i] = f.AsField()
	}

	return &ast.StructType{
		Fields: &ast.FieldList{
			List: fieldDefinitions,
		},
	}
}
