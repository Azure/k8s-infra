/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
)

// work around for golang initialization cycle
func delayed(p *gopter.Gen) gopter.Gen {
	return func(gp *gopter.GenParameters) *gopter.GenResult {
		return (*p)(gp)
	}
}

// a gopter Gen for Type
var genType gopter.Gen

func init() {
	genType = gen.OneGenOf(
		genTypeName,
		genPrimitiveType,
		genArrayType,
		genMapType,
	)
}

var genTypeName gopter.Gen = gopter.DeriveGen(
	func(group, version, name string) TypeName {
		return MakeTypeName(makeTestLocalPackageReference(group, version), name)
	},
	func(tn TypeName) (string, string, string) {
		local, ok := tn.PackageReference.AsLocalPackage()
		if !ok {
			panic("unable to convert to local package reference")
		}
		return local.group, local.version, tn.Name()
	},
	gen.AlphaString(),
	gen.AlphaString(),
	gen.AlphaString())

var genArrayType gopter.Gen = gopter.DeriveGen(
	func(t Type) Type {
		return NewArrayType(t)
	},
	func(t Type) Type {
		return t.(*ArrayType).Element()
	},
	delayed(&genType))

var genMapType gopter.Gen = gopter.DeriveGen(
	func(k, v Type) Type {
		return NewMapType(k, v)
	},
	func(t Type) (Type, Type) {
		mt := t.(*MapType)
		return mt.KeyType(), mt.ValueType()
	},
	delayed(&genType),
	delayed(&genType))

var genPrimitiveType gopter.Gen = gen.OneConstOf(
	IntType,
	StringType,
	UInt32Type,
	UInt64Type,
	FloatType,
	BoolType,
	AnyType,
)
