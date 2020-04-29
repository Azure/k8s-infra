/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"go/ast"
)

type ArrayType struct {
	element Type
}

func NewArrayType(element Type) *ArrayType {
	return &ArrayType{element}
}

func (array *ArrayType) AsType() ast.Expr {
	return &ast.ArrayType{
		Elt: array.element.AsType(),
	}
}
