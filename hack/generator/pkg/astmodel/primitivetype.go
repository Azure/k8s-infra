package astmodel

import (
	"go/ast"
)

// PrimitiveType represents a Go primitive
type PrimitiveType struct {
	name string
}

var Enum = &PrimitiveType{"ENUM"} // TODO

var IntType = &PrimitiveType{"int"}
var StringType = &PrimitiveType{"string"}
var FloatType = &PrimitiveType{"float64"}
var BoolType = &PrimitiveType{"bool"}
var AnyType = &PrimitiveType{"interface{}"}

// AsType implements Type for PrimitiveType
func (prim *PrimitiveType) AsType() ast.Expr {
	return ast.NewIdent(prim.name)
}
