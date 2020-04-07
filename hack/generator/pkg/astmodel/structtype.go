package astmodel

import "go/ast"

// StructType represents an (unnamed) struct type
type StructType struct {
	Fields []*FieldDefinition
}

// NewStructType is a factory method for creating a new StructTypeDefinition
func NewStructType(fields []*FieldDefinition) *StructType {
	return &StructType{fields}
}

// AsType implements Type for StructType
func (definition *StructType) AsType() ast.Expr {

	fieldDefinitions := make([]*ast.Field, len(definition.Fields))
	for i, f := range definition.Fields {
		fieldDefinitions[i] = f.AsField()
	}

	return &ast.StructType{
		Fields: &ast.FieldList{
			List: fieldDefinitions,
		},
	}
}
