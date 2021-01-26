/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"reflect"

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
		genOptionalType,
		genEnumType,
		//genOneOfType,
		//genAllOfType,
		//genObjectType,
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

var genOptionalType gopter.Gen = gopter.DeriveGen(
	func(t Type) Type {
		return NewOptionalType(t)
	},
	func(t Type) Type {
		// NewOptionalType doesn't always wrap its argument
		opt, ok := t.(*OptionalType)
		if ok {
			return opt.Element()
		}

		return t
	},
	delayed(&genType))

var genEnumType gopter.Gen = gopter.DeriveGen(
	func(values []string) Type {
		var enumValues []EnumValue
		for _, val := range values {
			enumValues = append(enumValues, EnumValue{val, val})
		}

		return NewEnumType(StringType, enumValues)
	},
	func(t Type) []string {
		var result []string
		for _, val := range t.(*EnumType).options {
			result = append(result, val.Value)
		}

		return result
	},
	gen.SliceOf(gen.AlphaString()))

// *** nothing below here works ***

var genObjectType gopter.Gen = gopter.DeriveGen(
	func(ps map[string]Type) Type {
		var propDefs []*PropertyDefinition
		for pn, pt := range ps {
			prop := NewPropertyDefinition(PropertyName(pn), pn, pt)
			propDefs = append(propDefs, prop)
		}

		return NewObjectType().WithProperties(propDefs...)
	},
	func(t Type) map[string]Type {
		result := make(map[string]Type)
		for pn, pt := range t.(*ObjectType).properties {
			result[string(pn)] = pt.propertyType
		}

		return result
	},
	gen.MapOf(gen.AlphaString(), delayed(&genType)))

var typeInterface = reflect.TypeOf((*Type)(nil)).Elem()
var genTypes = gen.SliceOf(delayedGenType, typeInterface)

// genOneOfType is not invertible
var genOneOfType gopter.Gen = genTypes.Map(func(types []Type) Type {
	return MakeOneOfType(types...)
})

// genAllOfType is not invertible
var genAllOfType gopter.Gen = genTypes.Map(func(types []Type) Type {
	return MakeAllOfType(types...)
})

func delayedGenType(gp *gopter.GenParameters) *gopter.GenResult {
	if gp == gopter.MinGenParams {
		return &gopter.GenResult{ResultType: typeInterface}
	}

	return genType(gp)
}
