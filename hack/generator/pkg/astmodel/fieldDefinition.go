package astmodel

import (
	"fmt"
	"go/ast"
)

// FieldDefinition encapsulates the definition of a field
type FieldDefinition struct {
	name        string
	fieldType   string
	description string
}

// NewFieldDefinition is a factory method for creating a new FieldDefinition
// name is the name for the new field (mandatory)
// fieldType is the type for the new field (mandatory)
func NewFieldDefinition(name string, fieldType string) *FieldDefinition {
	return &FieldDefinition{
		name:        name,
		fieldType:   fieldType,
		description: "",
	}
}

// Name returns the name of the field
func (field FieldDefinition) Name() string {
	return field.name
}

// FieldType returns the data type of the field
func (field FieldDefinition) FieldType() string {
	return field.fieldType
}

// WithDescription returns a new FieldDefinition with the specified description
func (field FieldDefinition) WithDescription(description string) FieldDefinition {
	field.description = description
	return field
}

// AsAst generates an AST node representing this field definition
func (field FieldDefinition) AsAst() (ast.Node, error) {
	ast, err := field.AsField()
	return &ast, err
}

// AsField generates an AST field node representing this field definition
func (field FieldDefinition) AsField() (ast.Field, error) {

	typeNode := ast.NewIdent(field.fieldType)

	// TODO: add field tags for api hints / json binding
	result := ast.Field{
		Names: []*ast.Ident{ast.NewIdent(field.name)},
		Type:  typeNode,
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

	return result, nil
}
