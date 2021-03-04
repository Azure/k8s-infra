/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"reflect"
)

// TypeMerger is like a visitor for 2 types.
// The `ctx` argument can be used to “smuggle” additional data down the call-chain.
type TypeMerger struct {
	mergers  map[TypePair]TypeMergerFunc
	fallback TypeMergerFunc
}

type TypePair struct{ left, right reflect.Type }
type TypeMergerFunc func(left, right Type) Type

func NewTypeMerger(fallback TypeMergerFunc) TypeMerger {
	return TypeMerger{
		mergers:  make(map[TypePair]TypeMergerFunc),
		fallback: fallback,
	}
}

var typeInterface reflect.Type = reflect.TypeOf((*Type)(nil)).Elem() // yuck

func (m *TypeMerger) Add(merger interface{}) {

	it := reflect.ValueOf(merger)
	if it.Kind() != reflect.Func {
		panic("merger must be a function")
	}

	mergerType := it.Type()

	if mergerType.NumIn() != 2 {
		panic("merger must take two arguments")
	}

	if mergerType.NumOut() != 1 {
		panic("merger must return one value")
	}

	if mergerType.Out(0) != typeInterface {
		panic("merger must return a Type")
	}

	leftArg := mergerType.In(0)
	rightArg := mergerType.In(1)
	if !leftArg.AssignableTo(typeInterface) || !rightArg.AssignableTo(typeInterface) {
		panic("merger must take in types assignable to Type")
	}

	key := TypePair{leftArg, rightArg}
	m.mergers[key] = func(left, right Type) Type {
		return it.Call([]reflect.Value{reflect.ValueOf(left), reflect.ValueOf(right)})[0].Interface().(Type)
	}
}

// Merge merges the two types according to the provided mergers and fallback
func (m *TypeMerger) Merge(left, right Type) Type {
	if left == nil {
		return right
	}

	if right == nil {
		return left
	}

	leftType := reflect.ValueOf(left).Type()
	rightType := reflect.ValueOf(right).Type()

	if merger, ok := m.mergers[TypePair{leftType, rightType}]; ok {
		return merger(left, right)
	}

	return m.fallback(left, right)
}
