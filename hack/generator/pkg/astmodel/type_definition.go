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

	// flags are used to identify resources during processing;
	// they have no direct representation in the generated code.
	flags map[Flag]struct{}
}

func MakeTypeDefinition(name TypeName, theType Type) TypeDefinition {
	return TypeDefinition{
		name:    name,
		theType: theType,
		flags:   make(map[Flag]struct{}),
	}
}

// Name returns the name being associated with the type
func (def TypeDefinition) Name() TypeName {
	return def.name
}

// Type returns the type being associated with the name
func (def TypeDefinition) Type() Type {
	return def.theType
}

// Description returns the description to be attached to this type definition (as a comment)
// We return a new slice to preserve immutability
func (def TypeDefinition) Description() []string {
	var result []string
	result = append(result, def.description...)
	return result
}

func (def TypeDefinition) References() TypeNameSet {
	return def.theType.References()
}

// WithDescription replaces the description of the definition with a new one (if any)
func (def TypeDefinition) WithDescription(desc []string) TypeDefinition {
	result := def.copy()
	result.description = append(result.description, desc...)
	return result
}

// WithType returns an updated TypeDefinition with the specified type
func (def TypeDefinition) WithType(t Type) TypeDefinition {
	result := def.copy()
	result.theType = t
	return result
}

// WithName returns an updated TypeDefinition with the specified name
func (def TypeDefinition) WithName(typeName TypeName) TypeDefinition {
	result := def.copy()
	result.name = typeName
	return result
}

func (def TypeDefinition) AsDeclarations(codeGenerationContext *CodeGenerationContext) []ast.Decl {
	return def.theType.AsDeclarations(codeGenerationContext, def.name, def.description)
}

// AddFlag includes the specified flag on the resource
func (def TypeDefinition) AddFlag(flag Flag) TypeDefinition {
	result := def.copy()
	result.flags[flag] = struct{}{}
	return result
}

// HasFlag returns true if the specified flag is present on this resource
func (def TypeDefinition) HasFlag(flag Flag) bool {
	_, ok := def.flags[flag]
	return ok
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
func (def TypeDefinition) RequiredImports() []PackageReference {
	return def.theType.RequiredImports()
}

// FileNameHint returns what a file that contains this name (if any) should be called
// this is not always used as we often combine multiple definitions into one file
func FileNameHint(name TypeName) string {
	return transformToSnakeCase(name.name)
}

// copy makes an independent deep clone of this type definition
func (def TypeDefinition) copy() TypeDefinition {
	result := def

	result.description = make([]string, len(def.description))
	copy(result.description, def.description)

	return result
}
