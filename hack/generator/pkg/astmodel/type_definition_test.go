/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"testing"

	. "github.com/onsi/gomega"
)

/*
 * MakeTypeDefinition() Tests
 */

func Test_MakeTypeDefinition_GivenValues_InitializesProperties(t *testing.T) {
	g := NewGomegaWithT(t)

	const name = "demo"
	const group = "group"
	const version = "2020-01-01"

	ref := MakeTypeName(MakeLocalPackageReference(group, version), name)
	objectType := NewObjectType().WithProperties(fullName, familyName, knownAs)
	objectDefinition := MakeTypeDefinition(ref, objectType)

	g.Expect(objectDefinition.Name().name).To(Equal(name))
	g.Expect(objectDefinition.Type()).To(Equal(objectType))

	definitionGroup, definitionPackage, err := objectDefinition.Name().PackageReference.GroupAndPackage()
	g.Expect(err).To(BeNil())

	g.Expect(definitionGroup).To(Equal(group))
	g.Expect(definitionPackage).To(Equal(version))
	g.Expect(objectDefinition.Type().(*ObjectType).properties).To(HaveLen(3))
}

/*
 * WithDescription() tests
 */

func Test_TypeDefinitionWithDescription_GivenDescription_ReturnsExpected(t *testing.T) {
	g := NewGomegaWithT(t)

	const name = "demo"
	const group = "group"
	const version = "2020-01-01"

	description := []string{"This is my test description"}

	ref := MakeTypeName(MakeLocalPackageReference(group, version), name)
	objectType := NewObjectType().WithProperties(fullName, familyName, knownAs)
	objectDefinition := MakeTypeDefinition(ref, objectType).WithDescription(description)

	g.Expect(objectDefinition.Description()).To(Equal(description))
}

/*
 * AsAst() Tests
 */

func Test_TypeDefinitionAsAst_GivenValidStruct_ReturnsNonNilResult(t *testing.T) {
	g := NewGomegaWithT(t)

	ref := MakeTypeName(MakeLocalPackageReference("group", "2020-01-01"), "name")
	definition := MakeTypeDefinition(ref, NewObjectType())
	node := definition.AsDeclarations(nil)

	g.Expect(node).NotTo(BeNil())
}

func createStringProperty(name string, description string) *PropertyDefinition {
	return NewPropertyDefinition(PropertyName(name), name, StringType).WithDescription(description)
}

func createIntProperty(name string, description string) *PropertyDefinition {
	return NewPropertyDefinition(PropertyName(name), name, IntType).WithDescription(description)
}

/*
 * AddFlag() tests
 */

func Test_TypeDefinitionAddFlag_GivenFlag_ReturnsDifferentInstance(t *testing.T) {
	g := NewGomegaWithT(t)

	ref := MakePackageReference("/foo/bar/baz")
	name := MakeTypeName(ref, "MyType")
	def := MakeTypeDefinition(name, StringType)
	flagged := def.AddFlag("foo")
	g.Expect(flagged).NotTo(BeIdenticalTo(def))
}

/*
 * HasFlag() tests
 */

func Test_TypeDefinitionHasFlag_GivenResourceWithFlag_ReturnsTrue(t *testing.T) {
	g := NewGomegaWithT(t)

	ref := MakePackageReference("/foo/bar/baz")
	name := MakeTypeName(ref, "MyType")
	def := MakeTypeDefinition(name, StringType).AddFlag("foo")

	g.Expect(def.HasFlag("foo")).To(BeTrue())
}

func Test_TypeDefinitionHasFlag_GivenResourceWithTwoFlags_ReturnsTrueForBoth(t *testing.T) {
	g := NewGomegaWithT(t)

	ref := MakePackageReference("/foo/bar/baz")
	name := MakeTypeName(ref, "MyType")
	def := MakeTypeDefinition(name, StringType).AddFlag("foo").AddFlag("bar")

	g.Expect(def.HasFlag("foo")).To(BeTrue())
	g.Expect(def.HasFlag("bar")).To(BeTrue())
}
