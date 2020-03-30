package astmodel

import (
	"go/ast"
	"go/token"
)

// StructDefinition encapsulates the definition of a struct
type StructDefinition struct {
	name   string
	fields []*FieldDefinition
}

// NewStructDefinition is a factory method for creating a new StructDefinition
func NewStructDefinition(name string, fields ...*FieldDefinition) *StructDefinition {
	return &StructDefinition{
		name:   name,
		fields: fields,
	}
}

// AsAst generates an AST node representing this field definition
func (definition *StructDefinition) AsAst() (ast.Node, error) {
	declaration, err := definition.AsDeclaration()
	return &declaration, err
}

// AsDeclaration generates an AST node representing this struct definition
func (definition *StructDefinition) AsDeclaration() (ast.GenDecl, error) {

	identifier := ast.NewIdent(definition.name)

	fieldDefinitions := make([]*ast.Field, len(definition.fields))
	for i, f := range definition.fields {
		definition, _ := f.AsField()
		fieldDefinitions[i] = &definition
	}

	structDefinition := &ast.StructType{
		Fields: &ast.FieldList{
			List: fieldDefinitions,
		},
	}

	typeSpecification := &ast.TypeSpec{
		Name: identifier,
		Type: structDefinition,
	}

	declaration := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			typeSpecification,
		},
	}

	return *declaration, nil
}

//TODO: Perhaps use this method in AsDeclaration(), above

// ToFieldList generates an AST fieldlist for a sequence of field definitions
func ToFieldList(fields []*FieldDefinition) (*ast.FieldList, error) {
	astFields := make([]*ast.Field, len(fields))
	for i, f := range fields {
		astField, err := f.AsField()
		if err != nil {
			return nil, err
		}
		astFields[i] = &astField
	}

	fieldList := &ast.FieldList{
		List: astFields,
	}

	return fieldList, nil
}
