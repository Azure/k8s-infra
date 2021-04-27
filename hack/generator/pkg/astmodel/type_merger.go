/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"reflect"

	"github.com/pkg/errors"
)

// TypeMerger is like a visitor for 2 types.
// The `ctx` argument can be used to “smuggle” additional data down the call-chain.
type TypeMerger struct {
	mergers  []mergerRegistration
	fallback MergerFunc
}

type mergerRegistration struct {
	left, right reflect.Type
	merge       MergerFunc
}

type MergerFunc func(ctx interface{}, left, right Type) (Type, error)

func NewTypeMerger(fallback MergerFunc) TypeMerger {
	return TypeMerger{fallback: fallback}
}

var typeInterface reflect.Type = reflect.TypeOf((*Type)(nil)).Elem() // yuck
var errorInterface reflect.Type = reflect.TypeOf((*error)(nil)).Elem()
var mergerFuncType reflect.Type = reflect.TypeOf((*MergerFunc)(nil)).Elem()

func validateMerger(merger interface{}) (reflect.Value, reflect.Type, reflect.Type, bool) {
	it := reflect.ValueOf(merger)
	if it.Kind() != reflect.Func {
		panic("merger must be a function")
	}

	mergerType := it.Type()

	if mergerType.NumIn() != 2 && mergerType.NumIn() != 3 {
		panic("merger must take 2 arguments (Type, Type) or 3 arguments (X, Type, Type)")
	}

	inOffset := mergerType.NumIn() - 2
	leftArg := mergerType.In(inOffset + 0)
	rightArg := mergerType.In(inOffset + 1)

	if !leftArg.AssignableTo(typeInterface) || !rightArg.AssignableTo(typeInterface) {
		panic("merger must take in types assignable to Type")
	}

	if mergerType.NumOut() != 2 ||
		mergerType.Out(0) != typeInterface ||
		mergerType.Out(1) != errorInterface {
		panic("merger must return (Type, error)")
	}

	return it, leftArg, rightArg, inOffset != 0
}

func (m *TypeMerger) Add(mergeFunc interface{}) {
	merger, leftArg, rightArg, takesCtx := validateMerger(mergeFunc)

	m.mergers = append(m.mergers,
		mergerRegistration{
			left:  leftArg,
			right: rightArg,
			merge: reflect.MakeFunc(mergerFuncType, func(args []reflect.Value) []reflect.Value {
				// we dereference the Type here to the underlying value so that
				// the merger can take either Type or a specific type.
				// if it takes Type then the compiler/runtime will convert the value back to a Type
				ctxValue := args[0].Elem()
				leftValue := args[1].Elem()
				rightValue := args[2].Elem()
				if takesCtx {
					return merger.Call([]reflect.Value{ctxValue, leftValue, rightValue})
				} else {
					return merger.Call([]reflect.Value{leftValue, rightValue})
				}
			}).Interface().(MergerFunc),
		})
}

func (m *TypeMerger) AddUnordered(mergeFunc interface{}) {
	merger, leftArg, rightArg, takesCtx := validateMerger(mergeFunc)

	m.mergers = append(m.mergers,
		mergerRegistration{
			left:  leftArg,
			right: rightArg,
			merge: reflect.MakeFunc(mergerFuncType, func(args []reflect.Value) []reflect.Value {
				ctxValue := args[0].Elem()
				leftValue := args[1].Elem()
				rightValue := args[2].Elem()
				if takesCtx {
					return merger.Call([]reflect.Value{ctxValue, leftValue, rightValue})
				} else {
					return merger.Call([]reflect.Value{leftValue, rightValue})
				}
			}).Interface().(MergerFunc),
		})

	m.mergers = append(m.mergers,
		mergerRegistration{
			left:  rightArg, // flipped
			right: leftArg,  // flipped
			merge: reflect.MakeFunc(mergerFuncType, func(args []reflect.Value) []reflect.Value {
				ctxValue := args[0].Elem()
				leftValue := args[2].Elem()  // flipped
				rightValue := args[1].Elem() // flipped
				if takesCtx {
					return merger.Call([]reflect.Value{ctxValue, leftValue, rightValue})
				} else {
					return merger.Call([]reflect.Value{leftValue, rightValue})
				}
			}).Interface().(MergerFunc),
		})
}

// Merge merges the two types according to the provided mergers and fallback
func (m *TypeMerger) Merge(left, right Type, ctx ...interface{}) (Type, error) {

	if len(ctx) > 1 {
		// optional argument
		panic("can only pass one ctx value")
	}

	var ctxValue interface{}
	if len(ctx) == 1 {
		ctxValue = ctx[0]
	}

	if left == nil {
		return right, nil
	}

	if right == nil {
		return left, nil
	}

	leftType := reflect.ValueOf(left).Type()
	rightType := reflect.ValueOf(right).Type()

	for _, merger := range m.mergers {
		if (merger.left == leftType || merger.left == typeInterface) &&
			(merger.right == rightType || merger.right == typeInterface) {

			result, err := merger.merge(ctxValue, left, right)

			if (result == nil && err == nil) || errors.Is(err, ContinueMerge) {
				// these conditions indicate that the merger was not actually applicable,
				// despite having a type that matched
				continue
			}

			return result, err
		}
	}

	return m.fallback(ctxValue, left, right)
}

var ContinueMerge error = errors.New("special error that indicates that the merger was not applicable")
