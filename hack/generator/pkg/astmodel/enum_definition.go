/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"go/ast"
	"go/token"
	"sort"
)

// EnumDefinition generates the full definition of an enumeration
type EnumDefinition struct {
	DefinitionName
	baseType *EnumType
}

var _ Definition = (*EnumDefinition)(nil)

// NewEnumDefinition is a factory method for creating new Enum Definitions
func NewEnumDefinition(name DefinitionName, t *EnumType) *EnumDefinition {
	return &EnumDefinition{DefinitionName: name, baseType: t}
}

// Name returns the unique name to use for specifying this enumeration
func (enum *EnumDefinition) Name() *DefinitionName {
	return &enum.DefinitionName
}

// Type returns the underlying EnumerationType for this enum
func (enum *EnumDefinition) Type() Type {
	return enum.baseType
}

// AsDeclarations generates the Go code representing this definition
func (enum *EnumDefinition) AsDeclarations() []ast.Decl {
	var specs []ast.Spec
	for _, v := range enum.baseType.Options {
		s := enum.createValueDeclaration(v)
		specs = append(specs, s)
	}

	declaration := &ast.GenDecl{
		Tok:   token.CONST,
		Doc:   &ast.CommentGroup{},
		Specs: specs,
	}

	result := []ast.Decl{
		enum.createBaseDeclaration(),
		declaration}

	return result
}

// Tidy does cleanup to ensure deterministic code generation
func (enum *EnumDefinition) Tidy() {
	sort.Slice(enum.baseType.Options, func(left int, right int) bool {
		return enum.baseType.Options[left].Identifier < enum.baseType.Options[right].Identifier
	})
}

func (enum *EnumDefinition) createBaseDeclaration() ast.Decl {
	var identifier *ast.Ident
	identifier = ast.NewIdent(enum.name)

	typeSpecification := &ast.TypeSpec{
		Name: identifier,
		Type: enum.baseType.BaseType.AsType(),
	}

	declaration := &ast.GenDecl{
		Tok: token.TYPE,
		Doc: &ast.CommentGroup{},
		Specs: []ast.Spec{
			typeSpecification,
		},
	}

	return declaration
}

func (enum *EnumDefinition) createValueDeclaration(value EnumValue) ast.Spec {

	var enumIdentifier *ast.Ident
	enumIdentifier = ast.NewIdent(enum.name)

	valueIdentifier := ast.NewIdent(enum.Name().name + value.Identifier)
	valueLiteral := ast.BasicLit{
		Kind:  token.STRING,
		Value: value.Value,
	}

	valueSpec := &ast.ValueSpec{
		Names: []*ast.Ident{valueIdentifier},
		Values: []ast.Expr{
			&ast.CallExpr{
				Fun:  enumIdentifier,
				Args: []ast.Expr{&valueLiteral},
			},
		},
	}

	return valueSpec
}

// CreateRelatedDefinitions returns a set of definitions related to this one
func (enum *EnumDefinition) CreateRelatedDefinitions(ref PackageReference, namehint string, idFactory IdentifierFactory) []Definition {
	return enum.baseType.CreateRelatedDefinitions(ref, namehint, idFactory)
}
