/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import "go/ast"

// PlaceholderTypeDefinition is used during JSON tree-walking
// to prevent infinite recursion. It will always be replaced before code emit.
type PlaceholderTypeDefinition struct {
	*TypeName
}

func NewPlaceholderTypeDefiner(name *TypeName) TypeDefiner {
	return &PlaceholderTypeDefinition{name}
}

// PlaceholderTypeDefinition is a TypeDefiner
var _ TypeDefiner = (*PlaceholderTypeDefinition)(nil)

func (placeholder *PlaceholderTypeDefinition) AsDeclarations() []ast.Decl {
	panic("Placeholder was not replaced! Bad code! Bad code!")
}

func (placeholder *PlaceholderTypeDefinition) Type() Type {
	panic("Placeholder was not replaced! Bad code! Bad code!")
}

func (placeholder *PlaceholderTypeDefinition) Name() *TypeName {
	return placeholder.TypeName
}
