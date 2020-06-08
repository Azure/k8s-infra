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

// AsDeclarations generates the actual Go declarations
func (namedType *NamedType) AsDeclarations(codeGenerationContext *CodeGenerationContext) []ast.Decl {
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

	return []ast.Decl{
		&ast.GenDecl{
			Doc: docComments,
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: ast.NewIdent(namedType.name.name),
					Type: namedType.underlyingType.AsType(codeGenerationContext),
				},
			},
		},
	}
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

// AsTypeReference renders a Go abstract syntax tree for referencing the type.
func (namedType *NamedType) AsTypeReference(codeGenerationContext *CodeGenerationContext) ast.Expr {
	return ast.NewIdent(namedType.name.name)
}

// AsType generates a reference to our type as an identifier
func (namedType *NamedType) AsType(_ *CodeGenerationContext) ast.Expr {
	return ast.NewIdent(namedType.name.name)
}

func (namedType *NamedType) Equals(t Type) bool {
	// TODO: This feels dodgy
	return namedType.name.Equals(t)
}

func (namedType *NamedType) CreateInternalDefinitions(nameHint *TypeName, idFactory IdentifierFactory) (Type, []*NamedType) {
	return namedType.underlyingType.CreateInternalDefinitions(nameHint, idFactory)
}

func (namedType *NamedType) CreateDefinitions(name *TypeName, idFactory IdentifierFactory, isResource bool) (*NamedType, []*NamedType) {
	return namedType.underlyingType.CreateDefinitions(name, idFactory, isResource)
}
