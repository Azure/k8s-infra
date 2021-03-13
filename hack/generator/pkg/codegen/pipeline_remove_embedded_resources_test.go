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

var exampleTypeFlag = astmodel.TypeFlag("flag")
var resourceTypeName = newTestName("Resource")
var resourceTypeName2 = newTestName("Resource2")

func newTestName(name string) astmodel.TypeName {
	return astmodel.MakeTypeName(makeTestLocalPackageReference("group", "2020-01-01"), name)
}

// TODO: This is copied from astmodel tests
func newTestObject(name astmodel.TypeName, fields ...*astmodel.PropertyDefinition) astmodel.TypeDefinition {
	return astmodel.MakeTypeDefinition(name, astmodel.NewObjectType().WithProperties(fields...))
}

func typesWithSubresourceTypeNoOriginalNameUsage() astmodel.Types {
	result := make(astmodel.Types)

	suffix := "TestSuffix"

	originalTypeName := newTestName("T1")
	modifiedTypeName := makeEmbeddedResourceTypeName(embeddedResourceTypeName{original: originalTypeName, context: resourceTypeName.Name(), suffix: suffix, count: 0})
	modifiedObject := newTestObject(modifiedTypeName)
	result.Add(modifiedObject.WithType(exampleTypeFlag.ApplyTo(modifiedObject.Type())))

	prop := astmodel.NewPropertyDefinition(
		"prop1",
		"prop1",
		modifiedTypeName)
	resource := newTestObject(resourceTypeName, prop)
	result.Add(resource)

	return result
}

func typesWithSubresourceTypeOriginalNameUsage() astmodel.Types {
	result := make(astmodel.Types)

	suffix := "TestSuffix"

	originalTypeName := newTestName("T1")
	modifiedTypeName := makeEmbeddedResourceTypeName(embeddedResourceTypeName{original: originalTypeName, context: resourceTypeName.Name(), suffix: suffix, count: 0})
	modifiedObject := newTestObject(modifiedTypeName)
	result.Add(modifiedObject.WithType(exampleTypeFlag.ApplyTo(modifiedObject.Type())))
	result.Add(newTestObject(originalTypeName))

	prop1 := astmodel.NewPropertyDefinition(
		"prop1",
		"prop1",
		modifiedTypeName)
	prop2 := astmodel.NewPropertyDefinition(
		"prop2",
		"prop2",
		originalTypeName)

	resource := newTestObject(resourceTypeName, prop1, prop2)
	result.Add(resource)

	return result
}

func typesWithSubresourceTypeMultipleUsageContextsOneResource() astmodel.Types {
	result := make(astmodel.Types)

	suffix := "TestSuffix"

	originalTypeName := newTestName("T1")
	modifiedTypeName1 := makeEmbeddedResourceTypeName(embeddedResourceTypeName{original: originalTypeName, context: resourceTypeName.Name(), suffix: suffix, count: 0})
	modifiedObject1 := newTestObject(modifiedTypeName1)
	result.Add(modifiedObject1.WithType(exampleTypeFlag.ApplyTo(modifiedObject1.Type())))

	modifiedTypeName2 := makeEmbeddedResourceTypeName(embeddedResourceTypeName{original: originalTypeName, context: resourceTypeName.Name(), suffix: suffix, count: 1})
	modifiedObject2 := newTestObject(modifiedTypeName2)
	result.Add(modifiedObject2.WithType(exampleTypeFlag.ApplyTo(modifiedObject2.Type())))

	//result.Add(newTestObject(originalTypeName))

	prop1 := astmodel.NewPropertyDefinition(
		"prop1",
		"prop1",
		modifiedTypeName1)
	prop2 := astmodel.NewPropertyDefinition(
		"prop2",
		"prop2",
		modifiedTypeName2)

	resource := newTestObject(resourceTypeName, prop1, prop2)
	result.Add(resource)

	return result
}

func typesWithSubresourceTypeMultipleResourcesOneUsageContextEach() astmodel.Types {
	result := make(astmodel.Types)

	suffix := "TestSuffix"

	originalTypeName := newTestName("T1")
	modifiedTypeName1 := makeEmbeddedResourceTypeName(embeddedResourceTypeName{original: originalTypeName, context: resourceTypeName.Name(), suffix: suffix, count: 0})
	modifiedObject1 := newTestObject(modifiedTypeName1)
	result.Add(modifiedObject1.WithType(exampleTypeFlag.ApplyTo(modifiedObject1.Type())))

	modifiedTypeName2 := makeEmbeddedResourceTypeName(embeddedResourceTypeName{original: originalTypeName, context: resourceTypeName2.Name(), suffix: suffix, count: 0})
	modifiedObject2 := newTestObject(modifiedTypeName2)
	result.Add(modifiedObject2.WithType(exampleTypeFlag.ApplyTo(modifiedObject2.Type())))

	prop1 := astmodel.NewPropertyDefinition(
		"prop1",
		"prop1",
		modifiedTypeName1)
	prop2 := astmodel.NewPropertyDefinition(
		"prop2",
		"prop2",
		modifiedTypeName2)

	resource := newTestObject(resourceTypeName, prop1, prop2)
	result.Add(resource)

	return result
}

func TestCleanupTypeNames_TypeWithNoOriginalName_UpdatedNameCollapsed(t *testing.T) {
	g := NewGomegaWithT(t)

	expectedUpdatedTypeName := newTestName("T1")

	updatedTypes, err := cleanupTypeNames(typesWithSubresourceTypeNoOriginalNameUsage(), exampleTypeFlag)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(len(updatedTypes)).To(Equal(2))

	ot, ok := astmodel.AsObjectType(updatedTypes[resourceTypeName].Type())
	g.Expect(ok).To(BeTrue())

	property, ok := ot.Property("prop1")
	g.Expect(ok).To(BeTrue())

	propertyTypeName, ok := astmodel.AsTypeName(property.PropertyType())
	g.Expect(ok).To(BeTrue())

	g.Expect(propertyTypeName.Equals(expectedUpdatedTypeName)).To(BeTrue())
}

func TestCleanupTypeNames_TypeWithOriginalNameExists_UpdatedNamePartiallyCollapsed(t *testing.T) {
	g := NewGomegaWithT(t)

	expectedUpdatedTypeName := newTestName("T1_TestSuffix")
	expectedOriginalTypeName := newTestName("T1")

	updatedTypes, err := cleanupTypeNames(typesWithSubresourceTypeOriginalNameUsage(), exampleTypeFlag)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(len(updatedTypes)).To(Equal(3))

	ot, ok := astmodel.AsObjectType(updatedTypes[resourceTypeName].Type())
	g.Expect(ok).To(BeTrue())

	prop1, ok := ot.Property("prop1")
	g.Expect(ok).To(BeTrue())
	prop1TypeName, ok := astmodel.AsTypeName(prop1.PropertyType())
	g.Expect(ok).To(BeTrue())
	g.Expect(prop1TypeName.Equals(expectedUpdatedTypeName)).To(BeTrue())

	prop2, ok := ot.Property("prop2")
	g.Expect(ok).To(BeTrue())
	prop2TypeName, ok := astmodel.AsTypeName(prop2.PropertyType())
	g.Expect(ok).To(BeTrue())
	g.Expect(prop2TypeName.Equals(expectedOriginalTypeName)).To(BeTrue())
}

func TestCleanupTypeNames_UpdatedNamesAreAllForSameResource_UpdatedNamesStrippedOfResourceContext(t *testing.T) {
	g := NewGomegaWithT(t)

	expectedUpdatedTypeName1 := newTestName("T1_TestSuffix")
	expectedUpdatedTypeName2 := newTestName("T1_TestSuffix_1")

	updatedTypes, err := cleanupTypeNames(typesWithSubresourceTypeMultipleUsageContextsOneResource(), exampleTypeFlag)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(len(updatedTypes)).To(Equal(3))

	ot, ok := astmodel.AsObjectType(updatedTypes[resourceTypeName].Type())
	g.Expect(ok).To(BeTrue())

	prop1, ok := ot.Property("prop1")
	g.Expect(ok).To(BeTrue())
	prop1TypeName, ok := astmodel.AsTypeName(prop1.PropertyType())
	g.Expect(ok).To(BeTrue())
	g.Expect(prop1TypeName.Equals(expectedUpdatedTypeName1)).To(BeTrue())

	prop2, ok := ot.Property("prop2")
	g.Expect(ok).To(BeTrue())
	prop2TypeName, ok := astmodel.AsTypeName(prop2.PropertyType())
	g.Expect(ok).To(BeTrue())
	g.Expect(prop2TypeName.Equals(expectedUpdatedTypeName2)).To(BeTrue())
}

func TestCleanupTypeNames_UpdatedNamesAreEachForDifferentResource_UpdatedNamesStrippedOfCount(t *testing.T) {
	g := NewGomegaWithT(t)

	expectedUpdatedTypeName1 := newTestName("T1_Resource_TestSuffix")
	expectedUpdatedTypeName2 := newTestName("T1_Resource2_TestSuffix")

	updatedTypes, err := cleanupTypeNames(typesWithSubresourceTypeMultipleResourcesOneUsageContextEach(), exampleTypeFlag)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(len(updatedTypes)).To(Equal(3))

	ot, ok := astmodel.AsObjectType(updatedTypes[resourceTypeName].Type())
	g.Expect(ok).To(BeTrue())

	prop1, ok := ot.Property("prop1")
	g.Expect(ok).To(BeTrue())
	prop1TypeName, ok := astmodel.AsTypeName(prop1.PropertyType())
	g.Expect(ok).To(BeTrue())
	g.Expect(prop1TypeName.Equals(expectedUpdatedTypeName1)).To(BeTrue())

	prop2, ok := ot.Property("prop2")
	g.Expect(ok).To(BeTrue())
	prop2TypeName, ok := astmodel.AsTypeName(prop2.PropertyType())
	g.Expect(ok).To(BeTrue())
	g.Expect(prop2TypeName.Equals(expectedUpdatedTypeName2)).To(BeTrue())
}
