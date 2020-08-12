package codegen

import (
	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	. "github.com/onsi/gomega"
	"testing"
)

const (
	testGroup   = "demo"
	testVersion = "v20200801"
	testName    = "testName"
)

/*
 * mapTypeName() tests
 */

func TestMapTypeName_GivenObjectReference_ReturnsNameInStoragePackage(t *testing.T) {
	g := NewGomegaWithT(t)

	objectName := createNameForTest(testName)
	objectDefinition := createObjectDefinitionForTest(objectName)
	types := createTypesForTest(objectDefinition)

	storageName := mapTypeName(objectName, types)

	grp, pkg, _ := storageName.PackageReference.GroupAndPackage()

	g.Expect(storageName.Name()).To(Equal(testName))
	g.Expect(grp).To(Equal(testGroup))
	g.Expect(pkg).To(Equal("v20200801s"))
}

func TestMapTypeName_GivenEnumReference_ReturnsOriginalName(t *testing.T) {
	g := NewGomegaWithT(t)

	enumName := createNameForTest(testName)
	enumDefinition := createEnumDefinitionForTest(enumName)
	types := createTypesForTest(enumDefinition)

	storageName := mapTypeName(enumName, types)

	grp, pkg, _ := storageName.PackageReference.GroupAndPackage()

	g.Expect(storageName.Name()).To(Equal(testName))
	g.Expect(grp).To(Equal(testGroup))
	g.Expect(pkg).To(Equal(testVersion))
}

/*
 * mapTypeNameForProperty() tests
 */

func TestMapTypeNameForProperty_GivenNameOfObject_ReturnsMappedName(t *testing.T) {
	g := NewGomegaWithT(t)

	objectName := createNameForTest(testName)
	objectDefinition := createObjectDefinitionForTest(objectName)
	types := createTypesForTest(objectDefinition)

	storageType := mapTypeNameForProperty(objectName, types)

	g.Expect(storageType).To(BeAssignableToTypeOf(astmodel.TypeName{}))

	name := storageType.(astmodel.TypeName)
	grp, pkg, _ := name.PackageReference.GroupAndPackage()

	g.Expect(name.Name()).To(Equal(testName))
	g.Expect(grp).To(Equal(testGroup))
	g.Expect(pkg).To(Equal("v20200801s"))
}

func TestMapTypeNameForProperty_GivenNameOfEnum_UnderlyingType(t *testing.T) {
	g := NewGomegaWithT(t)

	enumName := createNameForTest(testName)
	enumDefinition := createEnumDefinitionForTest(enumName)
	types := createTypesForTest(enumDefinition)

	storageType := mapTypeNameForProperty(enumName, types)

	g.Expect(storageType).To(Equal(astmodel.StringType))
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
	var	opts []astmodel.EnumValue
	for _, o := range options {
		opts = append(opts, astmodel.EnumValue{Identifier: o, Value: o})
	}
	enumType := astmodel.NewEnumType(astmodel.StringType, opts)
	enumDefinition := astmodel.MakeTypeDefinition(enumName, enumType)
	return enumDefinition
}
