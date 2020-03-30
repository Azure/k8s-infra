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
func NewFieldDefinition(name string, fieldType string, description string) *FieldDefinition {
	return &FieldDefinition{
		name:        name,
		fieldType:   fieldType,
		description: description,
	}
}

// AsAst generates an AST node representing this field definition
func (field *FieldDefinition) AsAst() (ast.Node, error) {
	ast, err := field.AsField()
	return &ast, err
}

// AsField generates an AST field node representing this field definition
func (field *FieldDefinition) AsField() (ast.Field, error) {

	typeNode := ast.NewIdent(field.fieldType)

	commentNode := &ast.CommentGroup{
		List: []*ast.Comment{
			{
				Text: fmt.Sprintf("/* %s */", field.description),
			},
		},
	}

	result := &ast.Field{
		Doc: commentNode,
		Names: []*ast.Ident{
			{
				Name: field.name,
			},
		},
		Type:    typeNode,
		Tag:     nil, // TODO: add field tags for api hints / json binding
		Comment: nil,
	}

	return *result, nil
}
