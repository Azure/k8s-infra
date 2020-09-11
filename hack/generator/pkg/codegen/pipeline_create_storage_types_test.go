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

const (
	testGroup      = "demo"
	testVersion    = "v20200801"
	testObjectName = "anObject"
	testEnumName   = "anEnum"
)

/*
 * mapTypeName() tests
 */

func TestMapTypeName_GivenObjectReference_ReturnsNameInStoragePackage(t *testing.T) {
	g := NewGomegaWithT(t)

	objectName := createNameForTest(testObjectName)
	objectDefinition := createObjectDefinitionForTest(objectName)
	types := createTypesForTest(objectDefinition)
	factory := StorageTypeFactory{
		types: types,
	}

	storageName, err := factory.mapTypeName(objectName)
	g.Expect(err).To(BeNil())

	grp, err := storageName.PackageReference.Group()
	g.Expect(err).To(BeNil())

	pkg := storageName.PackageReference.Package()
	g.Expect(storageName.Name()).To(Equal(testObjectName))
	g.Expect(grp).To(Equal(testGroup))
	g.Expect(pkg).To(Equal("v20200801s"))
}

func TestMapTypeName_GivenEnumReference_ReturnsOriginalName(t *testing.T) {
	g := NewGomegaWithT(t)

	enumName := createNameForTest(testEnumName)
	enumDefinition := createEnumDefinitionForTest(enumName)
	types := createTypesForTest(enumDefinition)
	factory := StorageTypeFactory{
		types: types,
	}

	storageName, err := factory.mapTypeName(enumName)
	g.Expect(err).To(BeNil())

	grp, pkg, _ := storageName.PackageReference.GroupAndPackage()
	g.Expect(storageName.Name()).To(Equal(testEnumName))
	g.Expect(grp).To(Equal(testGroup))
	g.Expect(pkg).To(Equal(testVersion))
}

/*
 * Support methods
 */

func createNameForTest(n string) astmodel.TypeName {
	return astmodel.MakeTypeName(astmodel.MakeLocalPackageReference(testGroup, testVersion), n)
}

func createTypesForTest(definitions ...astmodel.TypeDefinition) astmodel.Types {
	types := make(astmodel.Types)
	for _, d := range definitions {
		types.Add(d)
	}

	return types
}

func createObjectDefinitionForTest(objectName astmodel.TypeName) astmodel.TypeDefinition {
	idProperty := astmodel.NewPropertyDefinition("Id", "id", astmodel.StringType)
	objectType := astmodel.NewObjectType().WithProperties(idProperty)
	objectDefinition := astmodel.MakeTypeDefinition(objectName, objectType)
	return objectDefinition
}

func createEnumDefinitionForTest(enumName astmodel.TypeName, options ...string) astmodel.TypeDefinition {
	var opts []astmodel.EnumValue
	for _, o := range options {
		opts = append(opts, astmodel.EnumValue{Identifier: o, Value: o})
	}
	enumType := astmodel.NewEnumType(astmodel.StringType, opts)
	enumDefinition := astmodel.MakeTypeDefinition(enumName, enumType)
	return enumDefinition
}
