/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import "github.com/google/go-cmp/cmp"

var allowUnexported = cmp.AllowUnexported(
	PackageReference{},
	StructType{},
	PrimitiveType{},
	ArrayType{},
	TypeName{},
	MapType{},
	OptionalType{},
	EnumType{},
)

// NodesEqual compares two astmodel objects and returns whether they're
// equivalent. It uses github.com/google/go-cmp/cmp.Equal with a list
// of local types for which to allow comparing unexported fields.
func NodesEqual(a, b interface{}) bool {
	return cmp.Equal(a, b, allowUnexported)
}
