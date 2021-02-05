/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"bytes"
	"testing"

	"github.com/sebdah/goldie/v2"

	. "github.com/onsi/gomega"
)

type StorageConversionPropertyTestCase struct {
	name          string
	currentObject TypeDefinition
	nextObject    TypeDefinition
	hubObject     *TypeDefinition
	types         Types
}

func CreateStorageConversionFunctionTestCases() []*StorageConversionPropertyTestCase {

	vCurrent := makeTestLocalPackageReference("Verification", "vCurrent")
	vNext := makeTestLocalPackageReference("Verification", "vNext")
	vHub := makeTestLocalPackageReference("Verification", "vHub")

	alpha := EnumValue{Identifier: "Alpha", Value: "alpha"}
	beta := EnumValue{Identifier: "Beta", Value: "beta"}

	enumType := NewEnumType(StringType, alpha, beta)
	currentEnum := MakeTypeDefinition(MakeTypeName(vCurrent, "Release"), enumType)
	nextEnum := MakeTypeDefinition(MakeTypeName(vNext, "Release"), enumType)

	requiredStringProperty := NewPropertyDefinition("name", "name", StringType)
	optionalStringProperty := NewPropertyDefinition("name", "name", NewOptionalType(StringType))
	requiredIntProperty := NewPropertyDefinition("age", "age", IntType)
	optionalIntProperty := NewPropertyDefinition("age", "age", NewOptionalType(IntType))

	arrayOfRequiredIntProperty := NewPropertyDefinition("scores", "scores", NewArrayType(IntType))
	arrayOfOptionalIntProperty := NewPropertyDefinition("scores", "scores", NewArrayType(NewOptionalType(IntType)))

	mapOfRequiredIntsProperty := NewPropertyDefinition("ratings", "ratings", NewMapType(StringType, IntType))
	mapOfOptionalIntsProperty := NewPropertyDefinition("ratings", "ratings", NewMapType(StringType, NewOptionalType(IntType)))

	requiredCurrentEnumProperty := NewPropertyDefinition("release", "release", currentEnum.name)
	requiredNextEnumProperty := NewPropertyDefinition("release", "release", nextEnum.name)
	optionalCurrentEnumProperty := NewPropertyDefinition("release", "release", NewOptionalType(currentEnum.name))
	optionalNextEnumProperty := NewPropertyDefinition("release", "release", NewOptionalType(nextEnum.name))

	nastyProperty := NewPropertyDefinition(
		"nasty",
		"nasty",
		NewMapType(
			StringType,
			NewArrayType(
				NewMapType(StringType, BoolType))))

	testDirect := func(
		name string,
		currentProperty *PropertyDefinition,
		nextProperty *PropertyDefinition,
		otherDefinitions ...TypeDefinition) *StorageConversionPropertyTestCase {

		currentType := NewObjectType().WithProperty(currentProperty)
		currentDefinition := MakeTypeDefinition(
			MakeTypeName(vCurrent, "Person"),
			currentType)

		nextType := NewObjectType().WithProperty(nextProperty)
		nextDefinition := MakeTypeDefinition(
			MakeTypeName(vNext, "Person"),
			nextType)

		types := make(Types)
		types.Add(currentDefinition)
		types.Add(nextDefinition)
		types.AddAll(otherDefinitions)

		return &StorageConversionPropertyTestCase{
			name:          name + "-Direct",
			currentObject: currentDefinition,
			nextObject:    nextDefinition,
			types:         types,
		}
	}

	testIndirect := func(name string,
		currentProperty *PropertyDefinition,
		nextProperty *PropertyDefinition,
		otherDefinitions ...TypeDefinition) *StorageConversionPropertyTestCase {

		result := testDirect(name, currentProperty, nextProperty, otherDefinitions...)

		hubType := NewObjectType().WithProperty(nextProperty)
		hubDefinition := MakeTypeDefinition(
			MakeTypeName(vHub, "Person"),
			hubType)

		result.name = name + "-ViaStaging"
		result.hubObject = &hubDefinition
		result.types.Add(hubDefinition)

		return result
	}

	return []*StorageConversionPropertyTestCase{
		testDirect("SetStringFromString", requiredStringProperty, requiredStringProperty),
		testDirect("SetStringFromOptionalString", requiredStringProperty, optionalStringProperty),
		testDirect("SetOptionalStringFromString", optionalStringProperty, requiredStringProperty),
		testDirect("SetOptionalStringFromOptionalString", optionalStringProperty, optionalStringProperty),

		testIndirect("SetStringFromString", requiredStringProperty, requiredStringProperty),
		testIndirect("SetStringFromOptionalString", requiredStringProperty, optionalStringProperty),
		testIndirect("SetOptionalStringFromString", optionalStringProperty, requiredStringProperty),
		testIndirect("SetOptionalStringFromOptionalString", optionalStringProperty, optionalStringProperty),

		testDirect("SetIntFromInt", requiredIntProperty, requiredIntProperty),
		testDirect("SetIntFromOptionalInt", requiredIntProperty, optionalIntProperty),

		testIndirect("SetIntFromInt", requiredIntProperty, requiredIntProperty),
		testIndirect("SetIntFromOptionalInt", requiredIntProperty, optionalIntProperty),

		testDirect("SetArrayOfRequiredFromArrayOfRequired", arrayOfRequiredIntProperty, arrayOfRequiredIntProperty),
		testDirect("SetArrayOfRequiredFromArrayOfOptional", arrayOfRequiredIntProperty, arrayOfOptionalIntProperty),
		testDirect("SetArrayOfOptionalFromArrayOfRequired", arrayOfOptionalIntProperty, arrayOfRequiredIntProperty),

		testIndirect("SetArrayOfRequiredFromArrayOfRequired", arrayOfRequiredIntProperty, arrayOfRequiredIntProperty),
		testIndirect("SetArrayOfRequiredFromArrayOfOptional", arrayOfRequiredIntProperty, arrayOfOptionalIntProperty),
		testIndirect("SetArrayOfOptionalFromArrayOfRequired", arrayOfOptionalIntProperty, arrayOfRequiredIntProperty),

		testDirect("SetMapOfRequiredFromMapOfRequired", mapOfRequiredIntsProperty, mapOfRequiredIntsProperty),
		testDirect("SetMapOfRequiredFromMapOfOptional", mapOfRequiredIntsProperty, mapOfOptionalIntsProperty),
		testDirect("SetMapOfOptionalFromMapOfRequired", mapOfOptionalIntsProperty, mapOfRequiredIntsProperty),

		testIndirect("SetMapOfRequiredFromMapOfRequired", mapOfRequiredIntsProperty, mapOfRequiredIntsProperty),
		testIndirect("SetMapOfRequiredFromMapOfOptional", mapOfRequiredIntsProperty, mapOfOptionalIntsProperty),
		testIndirect("SetMapOfOptionalFromMapOfRequired", mapOfOptionalIntsProperty, mapOfRequiredIntsProperty),

		testDirect("NastyTest", nastyProperty, nastyProperty),
		testIndirect("NastyTest", nastyProperty, nastyProperty),

		testDirect("SetRequiredEnumFromRequiredEnum", requiredCurrentEnumProperty, requiredNextEnumProperty, currentEnum, nextEnum),
		testDirect("SetRequiredEnumFromOptionalEnum", requiredCurrentEnumProperty, optionalNextEnumProperty, currentEnum, nextEnum),
		testDirect("SetOptionalEnumFromRequiredEnum", optionalCurrentEnumProperty, requiredNextEnumProperty, currentEnum, nextEnum),
		testDirect("SetOptionalEnumFromOptionalEnum", optionalCurrentEnumProperty, optionalNextEnumProperty, currentEnum, nextEnum),
	}
}

func TestStorageConversionFunction_AsFunc(t *testing.T) {
	for _, c := range CreateStorageConversionFunctionTestCases() {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			RunTestStorageConversionFunction_AsFunc(c, t)
		})
	}
}

func RunTestStorageConversionFunction_AsFunc(c *StorageConversionPropertyTestCase, t *testing.T) {
	g := NewGomegaWithT(t)

	idFactory := NewIdentifierFactory()

	receiverType := NewObjectType().WithProperty(c.receiverProperty)
	receiverDefinition := MakeTypeDefinition(
		MakeTypeName(vCurrent, "Person"),
		receiverType)
	conversionContext := NewStorageConversionContext(c.types)

	hubTypeDefinition := MakeTypeDefinition(
		MakeTypeName(vHub, "Person"),
		NewObjectType().WithProperty(c.otherProperty))

	var intermediateTypeDefinition *TypeDefinition = nil
	if !direct {
		def := MakeTypeDefinition(
			MakeTypeName(vNext, "Person"),
			NewObjectType().WithProperty(c.otherProperty))
		intermediateTypeDefinition = &def
	}

	convertFrom, errs := NewStorageConversionFromFunction(receiverDefinition, hubTypeDefinition, intermediateTypeDefinition, conversionContext)
	gm.Expect(errs).To(BeNil())

	convertTo, errs := NewStorageConversionToFunction(receiverDefinition, hubTypeDefinition, intermediateTypeDefinition, conversionContext)
	gm.Expect(errs).To(BeNil())

	receiverDefinition = receiverDefinition.WithType(receiverType.WithFunction(convertFrom).WithFunction(convertTo))

	defs := []TypeDefinition{receiverDefinition}
	packages := make(map[PackageReference]*PackageDefinition)

	packageDefinition := NewPackageDefinition(vCurrent.Group(), vCurrent.PackageName(), "1")
	packageDefinition.AddDefinition(receiverDefinition)

	packages[currentPackage] = packageDefinition

	// put all definitions in one file, regardless.
	// the package reference isn't really used here.
	fileDef := NewFileDefinition(currentPackage, defs, packages)

	assertFileGeneratesExpectedCode(t, fileDef, c.name)
}

func assertFileGeneratesExpectedCode(t *testing.T, fileDef *FileDefinition, testName string) {
	g := goldie.New(t)

	buf := &bytes.Buffer{}
	fileWriter := NewGoSourceFileWriter(fileDef)
	err := fileWriter.SaveToWriter(buf)
	if err != nil {
		t.Fatalf("could not generate file: %v", err)
	}

	g.Assert(t, testName, buf.Bytes())
}
