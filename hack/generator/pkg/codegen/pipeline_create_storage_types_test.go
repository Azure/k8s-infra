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
)

/*
 * mapTypeNameIntoStoragePackage() tests
 */

func TestMapTypeNameIntoStoragePackage_GivenObjectReference_ReturnsNameInStoragePackage(t *testing.T) {
	g := NewGomegaWithT(t)

	objectName := createNameForTest(testObjectName)
	objectDefinition := createObjectDefinitionForTest(objectName)
	types := createTypesForTest(objectDefinition)
	factory := StorageTypeFactory{
		types: types,
	}

	storageName, err := factory.mapTypeNameIntoStoragePackage(objectName)
	g.Expect(err).To(BeNil())

	grp, err := storageName.PackageReference.Group()
	g.Expect(err).To(BeNil())

	pkg := storageName.PackageReference.Package()
	g.Expect(storageName.Name()).To(Equal(testObjectName))
	g.Expect(grp).To(Equal(testGroup))
	g.Expect(pkg).To(Equal("v20200801s"))
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
