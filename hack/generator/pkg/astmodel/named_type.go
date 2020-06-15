/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"go/ast"
	"go/token"
)

// NamedType represents a named type in the output files, and knows how to generate the Go AST
type NamedType struct {
	name           *TypeName
	underlyingType Type
	description    *string
}

// Ensure NamedType is also a Type
var _ Type = (*NamedType)(nil)

func NewNamedType(name *TypeName, underlyingType Type) *NamedType {
	return &NamedType{
		name:           name,
		underlyingType: underlyingType,
	}
}

// Name is the name that will be bound to the type
func (namedType *NamedType) Name() *TypeName {
	return namedType.name
}

// Type is the type that the name will be bound to
func (namedType *NamedType) Type() Type {
	return namedType.underlyingType
}

// WithDescription adds (or removes!) a description for the defined type
func (namedType *NamedType) WithDescription(description *string) *NamedType {
	if namedType.description == description || *namedType.description == *description {
		return namedType
	}

	return &NamedType{
		name:           namedType.name,
		underlyingType: namedType.underlyingType,
		description:    description,
	}
}

// AsTypeAst renders a Go abstract syntax tree for referencing the type.
// For named types, that reference is just the identifier
// TODO: Do we need to take the CodeGenerationContext into account?
func (namedType *NamedType) AsTypeAst(_ *CodeGenerationContext) ast.Expr {
	return ast.NewIdent(namedType.name.name)
}

// AsDeclarationAsts renders as Go abstract syntax trees for each required declaration
// Returns one or more type declarations for this specific type
func (namedType *NamedType) AsDeclarations(codeGenerationContext *CodeGenerationContext) []ast.Decl {
	return namedType.AsDeclarationAsts(namedType.name.name, codeGenerationContext)
}

// AsDeclarationAsts generates a declaration for our type
func (namedType *NamedType) AsDeclarationAsts(_ string, codeGenerationContext *CodeGenerationContext) []ast.Decl {

	nameHint := namedType.name.name

	var docComments *ast.CommentGroup
	if namedType.description != nil {
		docComments = &ast.CommentGroup{
			List: []*ast.Comment{
				{
					Text: "\n/*" + *namedType.description + "*/",
				},
			},
		}
	}

	otherTypeDecl := namedType.underlyingType.AsDeclarationAsts(nameHint, codeGenerationContext)

	decl := &ast.GenDecl{
		Doc: docComments,
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(nameHint),
				Type: namedType.underlyingType.AsTypeAst(codeGenerationContext),
			},
		},
	}

	result := []ast.Decl{decl}
	result = append(result, otherTypeDecl...)
	return result
}

// FileNameHint returns what a file that contains this definition (if any) should be called
// this is not always used as we might combine multiple definitions into one file
func (namedType *NamedType) FileNameHint() string {
	return transformToSnakeCase(namedType.Name().name)
}

func (namedType *NamedType) RequiredImports() []*PackageReference {
	return namedType.underlyingType.RequiredImports()
}

func (namedType *NamedType) References(name *TypeName) bool {
	return namedType.name.References(name) || namedType.underlyingType.References(name)
}

func (namedType *NamedType) Equals(t Type) bool {
	if nt, ok := t.(*NamedType); ok {
		if !namedType.name.Equals(nt.name) {
			return false
		}

		return namedType.underlyingType.Equals(nt.underlyingType)
	}

	return false
}

func (namedType *NamedType) CreateInternalDefinitions(nameHint *TypeName, idFactory IdentifierFactory) (Type, []*NamedType) {
	return namedType.underlyingType.CreateInternalDefinitions(nameHint, idFactory)
}

func (namedType *NamedType) CreateNamedTypes(name *TypeName, idFactory IdentifierFactory, isResource bool) (*NamedType, []*NamedType) {
	return namedType.underlyingType.CreateNamedTypes(name, idFactory, isResource)
}

func (namedType *NamedType) Visit(visitor func(t Type)) {
	visitor(namedType)
	namedType.underlyingType.Visit(visitor)
}
