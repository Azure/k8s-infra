package astmodel

import (
	"go/ast"
)

type MapType struct {
	key   Type
	value Type
}

func NewMap(key Type, value Type) *MapType {
	return &MapType{key, value}
}

func NewStringMap(value Type) *MapType {
	return NewMap(StringType, value)
}

// AsType implements Type for MapType
func (m *MapType) AsType() ast.Expr {
	return &ast.MapType{
		Key:   m.key.AsType(),
		Value: m.value.AsType(),
	}
}
