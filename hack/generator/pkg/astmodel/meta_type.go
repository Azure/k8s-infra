/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

// A MetaType is a type wrapper that provide additional information/context about another type
type MetaType interface {
	// Unwrap returns the type contained within the wrapper type
	Unwrap() Type
}

// AsPrimitiveType unwraps any wrappers around the provided type and returns either the underlying
// PrimitiveType or nil
func AsPrimitiveType(aType Type) *PrimitiveType {
	if primitive, ok := aType.(*PrimitiveType); ok {
		return primitive
	}

	if wrapper, ok := aType.(MetaType); ok {
		return AsPrimitiveType(wrapper.Unwrap())
	}

	return nil
}

// AsObjectType unwraps any wrappers around the provided type and returns either the underlying
// ObjectType or nil
func AsObjectType(aType Type) *ObjectType {
	if obj, ok := aType.(*ObjectType); ok {
		return obj
	}

	if wrapper, ok := aType.(MetaType); ok {
		return AsObjectType(wrapper.Unwrap())
	}

	return nil
}
