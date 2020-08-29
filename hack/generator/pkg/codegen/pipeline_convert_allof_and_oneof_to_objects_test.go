/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"testing"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"

	. "github.com/onsi/gomega"
)

var synth synthesizer = synthesizer{
	defs:      make(astmodel.Types),
	idFactory: astmodel.NewIdentifierFactory(),
}

var mapStringInterface = astmodel.NewStringMapType(astmodel.AnyType)
var mapStringString = astmodel.NewStringMapType(astmodel.StringType)
var mapInterfaceString = astmodel.NewMapType(astmodel.AnyType, astmodel.StringType)

var emptyObject = astmodel.NewObjectType()

func defineEnum(strings ...string) astmodel.Type {
	var values []astmodel.EnumValue
	for _, value := range strings {
		values = append(values, astmodel.EnumValue{
			Identifier: value,
			Value:      value,
		})
	}

	return astmodel.NewEnumType(
		astmodel.StringType,
		values)
}

// any type merged with AnyType is just the type
func TestMergeWithAny(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(synth.intersectTypes(astmodel.StringType, astmodel.AnyType)).To(Equal(astmodel.StringType))
	g.Expect(synth.intersectTypes(astmodel.AnyType, astmodel.StringType)).To(Equal(astmodel.StringType))
}

// merging maps is merging their keys and values
func TestMergeMaps(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(synth.intersectTypes(mapStringInterface, mapStringString)).To(Equal(mapStringString))
	g.Expect(synth.intersectTypes(mapStringString, mapStringInterface)).To(Equal(mapStringString))
	g.Expect(synth.intersectTypes(mapInterfaceString, mapStringInterface)).To(Equal(mapStringString))
	g.Expect(synth.intersectTypes(mapStringInterface, mapInterfaceString)).To(Equal(mapStringString))
}

// merging a map with string keys with an empty object results in the map
func TestMergeMapEmptyObject(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(synth.intersectTypes(mapStringString, emptyObject)).To(Equal(mapStringString))
	g.Expect(synth.intersectTypes(emptyObject, mapStringString)).To(Equal(mapStringString))
}

// merging a map with an object puts the map into 'additionalProperties'
func TestMergeMapObject(t *testing.T) {
	g := NewGomegaWithT(t)

	newMap := astmodel.NewMapType(astmodel.StringType, astmodel.FloatType)
	oneProp := astmodel.NewObjectType().WithProperties(
		astmodel.NewPropertyDefinition("x", "x", astmodel.IntType),
	)

	expected := oneProp.WithProperties(
		astmodel.NewPropertyDefinition("additionalProperties", "additionalProperties", newMap),
	)

	g.Expect(synth.intersectTypes(newMap, oneProp)).To(Equal(expected))
	g.Expect(synth.intersectTypes(oneProp, newMap)).To(Equal(expected))
}

// merging two objects results in the union of their properties
func TestMergeObjectObject(t *testing.T) {
	g := NewGomegaWithT(t)

	propX := astmodel.NewPropertyDefinition("x", "x", astmodel.IntType)
	obj1 := astmodel.NewObjectType().WithProperties(propX)

	propY := astmodel.NewPropertyDefinition("y", "y", astmodel.FloatType)
	obj2 := astmodel.NewObjectType().WithProperties(propY)

	expected := astmodel.NewObjectType().WithProperties(propX, propY)

	g.Expect(synth.intersectTypes(obj1, obj2)).To(Equal(expected))
	g.Expect(synth.intersectTypes(obj2, obj1)).To(Equal(expected))
}

// merging two enums results in the intersection of their values
func TestMergeEnumEnum(t *testing.T) {
	g := NewGomegaWithT(t)

	enum1 := defineEnum("x", "y", "z", "g")
	enum2 := defineEnum("g", "a", "b", "c")

	expected := defineEnum("g")

	g.Expect(synth.intersectTypes(enum1, enum2)).To(Equal(expected))
	g.Expect(synth.intersectTypes(enum2, enum1)).To(Equal(expected))
}

// cannot merge enums with different base types
func TestMergeBadEnums(t *testing.T) {
	g := NewGomegaWithT(t)

	enum := defineEnum("a", "b")
	enumInt := astmodel.NewEnumType(astmodel.IntType, []astmodel.EnumValue{})

	var err error

	_, err = synth.intersectTypes(enum, enumInt)
	g.Expect(err.Error()).To(ContainSubstring("differing base types"))

	_, err = synth.intersectTypes(enumInt, enum)
	g.Expect(err.Error()).To(ContainSubstring("differing base types"))
}

// merging an enum with its base type results in the enum
func TestMergeEnumBaseType(t *testing.T) {
	g := NewGomegaWithT(t)

	enum := defineEnum("a", "b")

	g.Expect(synth.intersectTypes(enum, astmodel.StringType)).To(Equal(enum))
	g.Expect(synth.intersectTypes(astmodel.StringType, enum)).To(Equal(enum))
}

// cannot merge an enum with another non-base type
func TestMergeEnumWrongBaseType(t *testing.T) {
	g := NewGomegaWithT(t)

	enum := defineEnum("a", "b")

	var err error

	_, err = synth.intersectTypes(enum, astmodel.IntType)
	g.Expect(err.Error()).To(ContainSubstring("don't know how to merge enum type"))

	_, err = synth.intersectTypes(astmodel.IntType, enum)
	g.Expect(err.Error()).To(ContainSubstring("don't know how to merge enum type"))
}

// merging two optionals merges their contents
func TestMergeOptionalOptional(t *testing.T) {
	g := NewGomegaWithT(t)

	enum1 := astmodel.NewOptionalType(defineEnum("a", "b"))
	enum2 := astmodel.NewOptionalType(defineEnum("b", "c"))

	expected := astmodel.NewOptionalType(defineEnum("b"))

	g.Expect(synth.intersectTypes(enum1, enum2)).To(Equal(expected))
	g.Expect(synth.intersectTypes(enum2, enum1)).To(Equal(expected))
}

// merging an optional with something else that it can be merged with results in that result
// TODO: dubious?
func TestMergeOptionalEnum(t *testing.T) {
	// this feels a bit wrong but it seems to be expected in real life specs

	g := NewGomegaWithT(t)

	enum1 := defineEnum("a", "b")
	enum2 := astmodel.NewOptionalType(defineEnum("b", "c"))

	expected := defineEnum("b")

	g.Expect(synth.intersectTypes(enum1, enum2)).To(Equal(expected))
	g.Expect(synth.intersectTypes(enum2, enum1)).To(Equal(expected))
}

// merging objects with common properties merges the types of those properties
func TestMergeObjectObjectCommonProperties(t *testing.T) {
	g := NewGomegaWithT(t)

	obj1 := astmodel.NewObjectType().WithProperties(
		astmodel.NewPropertyDefinition("x", "x", defineEnum("a", "b")))

	obj2 := astmodel.NewObjectType().WithProperties(
		astmodel.NewPropertyDefinition("x", "x", defineEnum("b", "c")))

	expected := astmodel.NewObjectType().WithProperties(
		astmodel.NewPropertyDefinition("x", "x", defineEnum("b")))

	g.Expect(synth.intersectTypes(obj1, obj2)).To(Equal(expected))
	g.Expect(synth.intersectTypes(obj2, obj1)).To(Equal(expected))
}

// merging a oneOf with a type that is in the oneOf results in that type
func TestMergeOneOf(t *testing.T) {
	g := NewGomegaWithT(t)

	oneOf := astmodel.MakeOneOfType([]astmodel.Type{astmodel.IntType, astmodel.StringType, astmodel.BoolType})

	g.Expect(synth.intersectTypes(oneOf, astmodel.BoolType)).To(Equal(astmodel.BoolType))
	g.Expect(synth.intersectTypes(astmodel.BoolType, oneOf)).To(Equal(astmodel.BoolType))
}

// wooh!
func TestMergeOneOfEnum(t *testing.T) {
	// this is based on a real-world example
	g := NewGomegaWithT(t)

	oneOf := astmodel.MakeOneOfType([]astmodel.Type{
		astmodel.StringType,
		defineEnum("a", "b", "c"),
	})

	enum := defineEnum("b")

	expected := defineEnum("b")

	g.Expect(synth.intersectTypes(oneOf, enum)).To(Equal(expected))
	g.Expect(synth.intersectTypes(enum, oneOf)).To(Equal(expected))
}

func TestOneOfResourceSpec(t *testing.T) {
	g := NewGomegaWithT(t)

	r := astmodel.NewResourceType(astmodel.StringType, astmodel.IntType)
	oneOf := astmodel.MakeOneOfType([]astmodel.Type{astmodel.BoolType, r}).(astmodel.OneOfType)

	synth.specOrStatus = chooseSpec

	expected := astmodel.NewObjectType().WithProperties(
		astmodel.NewPropertyDefinition(astmodel.PropertyName("Bool0"), "bool0", astmodel.BoolType).
			MakeOptional().WithDescription("mutually exclusive with all other properties"),
		astmodel.NewPropertyDefinition(astmodel.PropertyName("String1"), "string1", astmodel.StringType).MakeOptional().
			MakeOptional().WithDescription("mutually exclusive with all other properties"),
	)

	result, err := synth.oneOfObject(oneOf)
	g.Expect(err).To(BeNil())

	result = astmodel.NewObjectType().WithProperties(result.(*astmodel.ObjectType).Properties()...)
	g.Expect(result).To(Equal(expected))
}