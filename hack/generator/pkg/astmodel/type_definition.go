/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"go/ast"
	"go/token"
)

// TypeDefinition is a name paired with a type
type TypeDefinition struct {
	name        TypeName
	description []string
	theType     Type
}

func MakeTypeDefinition(name TypeName, theType Type) TypeDefinition {
	return TypeDefinition{name: name, theType: theType}
}

// Name returns the name being associated with the type
func (std TypeDefinition) Name() TypeName {
	return std.name
}

// Type returns the type being associated with the name
func (std TypeDefinition) Type() Type {
	return std.theType
}

// Description returns the description to be attached to this type definition (as a comment)
// We return a new slice to preserve immutability
func (std TypeDefinition) Description() []string {
	var result []string
	for _, s := range std.description {
		result = append(result, s)
	}

	return result
}

func (std TypeDefinition) References() TypeNameSet {
	return std.theType.References()
}

// WithDescription replaces the description of the definition with a new one (if any)
func (std TypeDefinition) WithDescription(desc ...string) TypeDefinition {
	std.description = desc
	return std
}

// WithType returns an updated TypeDefinition with the specified type
func (std TypeDefinition) WithType(t Type) TypeDefinition {
	result := std
	result.theType = t
	return result
}

// WithName returns an updated TypeDefinition with the specified name
func (std TypeDefinition) WithName(typeName TypeName) TypeDefinition {
	result := std
	result.name = typeName
	return result
}

func (std TypeDefinition) AsDeclarations(codeGenerationContext *CodeGenerationContext) []ast.Decl {
	return std.theType.AsDeclarations(codeGenerationContext, std.name, std.description)
}

// AsSimpleDeclarations is a helper for types that only require a simple name/alias to be defined
func AsSimpleDeclarations(codeGenerationContext *CodeGenerationContext, name TypeName, description []string, theType Type) []ast.Decl {
	var docComments *ast.CommentGroup
	if len(description) > 0 {
		docComments = &ast.CommentGroup{}
		addDocComments(&docComments.List, description, 120)
	}

	return []ast.Decl{
		&ast.GenDecl{
			Doc: docComments,
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: ast.NewIdent(name.Name()),
					Type: theType.AsType(codeGenerationContext),
				},
			},
		},
	}
}

// RequiredImports returns a list of packages required by this type
func (std TypeDefinition) RequiredImports() []PackageReference {
	return std.theType.RequiredImports()
}

// FileNameHint returns what a file that contains this name (if any) should be called
// this is not always used as we often combine multiple definitions into one file
func FileNameHint(name TypeName) string {
	return transformToSnakeCase(name.name)
}
