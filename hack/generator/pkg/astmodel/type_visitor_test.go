/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"github.com/pkg/errors"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_Visit_GivenCountingTypeVisitor_ReturnsExpectedCounts(t *testing.T) {

	arrType := NewArrayType(StringType)
	mapType := NewMapType(StringType, StringType)
	optType := NewOptionalType(IntType)

	intProp := NewPropertyDefinition("number", "number", IntType)
	arrProp := NewPropertyDefinition("arrType", "arrType", arrType)
	mapProp := NewPropertyDefinition("mapProp", "mapProp", mapType)
	optProp := NewPropertyDefinition("optType", "optType", optType)

	objType := EmptyObjectType.WithProperties(intProp, arrProp, mapProp, optProp)

	resType := NewResourceType(objType, StringType)

	cases := []struct {
		name     string
		theType  Type
		expected int
	}{
		// Simple types get visited once
		{"Strings are visited once", StringType, 1},
		{"Integers are visited once", IntType, 1},
		{"Booleans are visited once", BoolType, 1},
		// Composite types require more visits
		{"Arrays require two visits", arrType, 2},
		{"Map types require three visits", mapType, 3},
		{"Optional types require two visits", optType, 2},
		// Complex types require even more visits
		{"Object types require one visit plus visiting each property", objType, 9},
		{"Resource types require one visit plus visiting spec and status", resType, 11},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)

			v := MakeCountingTypeVisitor()
			_, err := v.Visit(c.theType, nil)

			g.Expect(v.VisitCount).To(Equal(c.expected))
			g.Expect(err).To(BeNil())
		})
	}
}

type CountingTypeVisitor struct {
	TypeVisitor
	VisitCount int
}

func MakeCountingTypeVisitor() *CountingTypeVisitor {
	v := MakeTypeVisitor()

	result := &CountingTypeVisitor{}

	result.VisitTypeName = func(this *TypeVisitor, it TypeName, ctx interface{}) (Type, error) {
		result.VisitCount++
		return v.VisitTypeName(this, it, ctx)
	}

	result.VisitArrayType = func(this *TypeVisitor, it *ArrayType, ctx interface{}) (Type, error) {
		result.VisitCount++
		return v.VisitArrayType(this, it, ctx)
	}

	result.VisitPrimitive = func(this *TypeVisitor, it *PrimitiveType, ctx interface{}) (Type, error) {
		result.VisitCount++
		return v.VisitPrimitive(this, it, ctx)
	}

	result.VisitObjectType = func(this *TypeVisitor, it *ObjectType, ctx interface{}) (Type, error) {
		result.VisitCount++
		return v.VisitObjectType(this, it, ctx)
	}

	result.VisitMapType = func(this *TypeVisitor, it *MapType, ctx interface{}) (Type, error) {
		result.VisitCount++
		return v.VisitMapType(this, it, ctx)
	}

	result.VisitEnumType = func(this *TypeVisitor, it *EnumType, ctx interface{}) (Type, error) {
		result.VisitCount++
		return v.VisitEnumType(this, it, ctx)
	}

	result.VisitOptionalType = func(this *TypeVisitor, it *OptionalType, ctx interface{}) (Type, error) {
		result.VisitCount++
		return v.VisitOptionalType(this, it, ctx)
	}

	result.VisitResourceType = func(this *TypeVisitor, it *ResourceType, ctx interface{}) (Type, error) {
		result.VisitCount++
		return v.VisitResourceType(this, it, ctx)
	}

	return result
}

func TestIdentityVisitorReturnsEqualResult(t *testing.T) {

	mapOfStringToString := NewMapType(StringType, StringType)
	mapOfStringToBool := NewMapType(StringType, BoolType)

	fullName := NewPropertyDefinition("FullName", "full-name", StringType)
	familyName := NewPropertyDefinition("FamilyName", "family-name", StringType)
	knownAs := NewPropertyDefinition("KnownAs", "known-as", StringType)

	transform := NewFakeFunction("Transform")
	transmogrify := NewFakeFunction("Transmogrify")
	skew := NewFakeFunction("Skew")

	person := NewObjectType().WithProperties(fullName, familyName, knownAs)
	individual := NewObjectType().WithProperties(fullName).
		WithFunction(transform).
		WithFunction(transmogrify).
		WithFunction(skew)

	resource := NewResourceType(person, individual)

	cases := []struct {
		name    string
		subject Type
	}{
		{"Strings", StringType},
		{"Integers", IntType},
		{"Booleans", BoolType},
		{"Map of String to String", mapOfStringToString},
		{"Map of String to Bool", mapOfStringToBool},
		{"Object type with properties", person},
		{"Object type with functions", individual},
		{"Resource type", resource},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)

			v := MakeTypeVisitor()
			result, err := v.Visit(c.subject, nil)

			g.Expect(err).To(BeNil())
			g.Expect(result.Equals(c.subject)).To(BeTrue())

			// Test both ways to be paranoid
			// If this test fails while the previous line passes, there's likely a bug in .Equals()
			g.Expect(c.subject.Equals(result)).To(BeTrue())
		})
	}
}

func TestMakeTypeVisitorWithInjectedFunctions(t *testing.T) {

	cases := []struct {
		name          string
		subject       Type
		visitation    interface{}
		expectedType  Type
		expectedError string
	}{
		{
			"PrimitiveTypeHandler",
			StringType,
			func(tv *TypeVisitor, p *PrimitiveType, _ interface{}) (Type, error) {
				return p, errors.New(p.name)
			},
			StringType,
			"string",
		},
		{
			"PrimitiveTypeSimplified",
			StringType,
			func(p *PrimitiveType) (Type, error) {
				return p, errors.New(p.name)
			},
			StringType,
			"string",
		},
		{
			"OptionalTypeHandler",
			NewOptionalType(IntType),
			func(tv *TypeVisitor, ot *OptionalType, _ interface{}) (Type, error) {
				return ot, errors.New(ot.String())
			},
			NewOptionalType(IntType),
			"optional",
		},
		{
			"OptionalTypeSimplified",
			NewOptionalType(IntType),
			func(ot *OptionalType) (Type, error) {
				return ot, errors.New(ot.String())
			},
			NewOptionalType(IntType),
			"optional",
		},
		{
			"MapTypeHandler",
			NewMapType(StringType, IntType),
			func(tv *TypeVisitor, mt *MapType, _ interface{}) (Type, error) {
				return mt, errors.New(mt.String())
			},
			NewMapType(StringType, IntType),
			"map",
		},
		{
			"MapTypeSimplified",
			NewMapType(StringType, IntType),
			func(mt *MapType) (Type, error) {
				return mt, errors.New(mt.String())
			},
			NewMapType(StringType, IntType),
			"map",
		},
		{
			"ArrayTypeHandler",
			NewArrayType(StringType),
			func(tv *TypeVisitor, at *ArrayType, _ interface{}) (Type, error) {
				return at, errors.New(at.String())
			},
			NewArrayType(StringType),
			"[]string",
		},
		{
			"ArrayTypeSimplified",
			NewArrayType(StringType),
			func(at *ArrayType) (Type, error) {
				return at, errors.New(at.String())
			},
			NewArrayType(StringType),
			"[]string",
		},
		{
			"FlaggedTypeHandler",
			StorageFlag.ApplyTo(StringType),
			func(tv *TypeVisitor, ft *FlaggedType, _ interface{}) (Type, error) {
				return ft, errors.New(ft.String())
			},
			StorageFlag.ApplyTo(StringType),
			"storage",
		},
		{
			"FlaggedTypeSimplified",
			StorageFlag.ApplyTo(StringType),
			func(ft *FlaggedType) (Type, error) {
				return ft, errors.New(ft.String())
			},
			StorageFlag.ApplyTo(StringType),
			"storage",
		},
	}

	// TODO: Pending tests for
	// VisitTypeName:
	// VisitObjectType:
	// VisitEnumType:
	// VisitResourceType:
	// VisitOneOfType:
	// VisitAllOfType:
	// VisitValidatedType:
	// VisitErroredType:

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)

			v := MakeTypeVisitor(c.visitation)
			typ, err := v.Visit(c.subject, nil)
			g.Expect(typ.Equals(c.expectedType)).To(BeTrue())
			g.Expect(err.Error()).To(ContainSubstring(c.expectedError))
		})
	}
}
