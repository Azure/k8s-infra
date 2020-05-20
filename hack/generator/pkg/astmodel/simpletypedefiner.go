/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"go/ast"
	"go/token"
)

// SimpleTypeDefiner is a TypeDefiner for simple cases (not structs or enums)
type SimpleTypeDefiner struct {
	name    *DefinitionName
	theType Type
}

func (gtd *SimpleTypeDefiner) Name() *DefinitionName {
	return gtd.name
}

func (gtd *SimpleTypeDefiner) Type() Type {
	return gtd.theType
}

func (gtd *SimpleTypeDefiner) AsDeclarations() []ast.Decl {
	return []ast.Decl{
		&ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: ast.NewIdent(gtd.name.name),
					Type: gtd.theType.AsType(),
				},
			},
		},
	}
}
