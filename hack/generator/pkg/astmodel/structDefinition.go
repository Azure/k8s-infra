package astmodel

import (
	"go/ast"
	"go/token"
)

// StructDefinition encapsulates the definition of a struct
type StructDefinition struct {
	name    string
	version string
	fields  []*FieldDefinition
}

// NewStructDefinition is a factory method for creating a new StructDefinition
func NewStructDefinition(name string, version string, fields ...*FieldDefinition) *StructDefinition {
	return &StructDefinition{
		name:    name,
		version: version,
		fields:  fields,
	}
}

// Name returns the name of the struct
func (definition *StructDefinition) Name() string {
	return definition.name
}

// Version returns the version of this struct
func (definition *StructDefinition) Version() string {
	return definition.version
}

// Field provides indexed access to our fields
func (definition *StructDefinition) Field(index int) FieldDefinition {
	return *definition.fields[index]
}

// FieldCount indicates how many fields are contained
func (definition *StructDefinition) FieldCount() int {
	return len(definition.fields)
}

// AsAst generates an AST node representing this field definition
func (definition *StructDefinition) AsAst() ast.Node {
	return definition.AsDeclaration()
}

// AsDeclaration generates an AST node representing this struct definition
func (definition *StructDefinition) AsDeclaration() *ast.GenDecl {

	identifier := ast.NewIdent(definition.name)

	fieldDefinitions := make([]*ast.Field, len(definition.fields))
	for i, f := range definition.fields {
		fieldDefinitions[i] = f.AsField()
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

	return declaration
}

//TODO: Perhaps use this method in AsDeclaration(), above
//TODO make this private as it might be unused elsewhere

// ToFieldList generates an AST fieldlist for a sequence of field definitions
func ToFieldList(fields []*FieldDefinition) *ast.FieldList {
	astFields := make([]*ast.Field, len(fields))
	for i, f := range fields {
		astFields[i] = f.AsField()
	}

	fieldList := &ast.FieldList{
		List: astFields,
	}

	return fieldList
}
