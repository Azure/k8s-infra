package astmodel

import (
	"fmt"
	"go/ast"
)

// FieldDefinition encapsulates the definition of a field
type FieldDefinition struct {
	name        string
	fieldType   Type
	description string
}

// NewFieldDefinition is a factory method for creating a new FieldDefinition
// name is the name for the new field (mandatory)
// fieldType is the type for the new field (mandatory)
func NewFieldDefinition(name string, fieldType Type) *FieldDefinition {
	return &FieldDefinition{
		name:        name,
		fieldType:   fieldType,
		description: "",
	}
}

// NewEmbeddedStructDefinition is a factory method for defining an embedding
// of another struct type.
func NewEmbeddedStructDefinition(structType Type) *FieldDefinition {
	// in Go, this is just a field without a name:
	return &FieldDefinition{
		name:        "",
		fieldType:   structType,
		description: "",
	}
}

// Name returns the name of the field
func (field *FieldDefinition) Name() string {
	return field.name
}

// FieldType returns the data type of the field
func (field *FieldDefinition) FieldType() Type {
	return field.fieldType
}

// WithDescription returns a new FieldDefinition with the specified description
func (field *FieldDefinition) WithDescription(description *string) *FieldDefinition {
	if description == nil {
		return field
	}

	result := *field
	result.description = *description
	return &result
}

// AsAst generates an AST node representing this field definition
func (field FieldDefinition) AsAst() ast.Node {
	return field.AsField()
}

// AsField generates an AST field node representing this field definition
func (field FieldDefinition) AsField() *ast.Field {

	// TODO: add field tags for api hints / json binding
	result := &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(field.name)},
		Type:  field.FieldType().AsType(),
	}

	if field.description != "" {
		result.Doc = &ast.CommentGroup{
			List: []*ast.Comment{
				{
					Text: fmt.Sprintf("\n/*\t%s: %s */", field.name, field.description),
				},
			},
		}
	}

	return result
}
