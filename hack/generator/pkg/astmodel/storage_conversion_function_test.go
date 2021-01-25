package astmodel

import (
	"bytes"
	"github.com/sebdah/goldie/v2"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

type StorageConversionPropertyTestCase struct {
	name             string
	receiverProperty *PropertyDefinition
	otherProperty    *PropertyDefinition
}

func CreateStorageConversionFunctionTestCases() []StorageConversionPropertyTestCase {
	requiredStringProperty := NewPropertyDefinition("name", "required-string", StringType)
	optionalStringProperty := NewPropertyDefinition("name", "optional-string", NewOptionalType(StringType))
	requiredIntProperty := NewPropertyDefinition("age", "required-int", IntType)
	optionalIntProperty := NewPropertyDefinition("age", "optional-int", NewOptionalType(IntType))

	arrayOfRequiredIntProperty := NewPropertyDefinition("scores", "array-required-int", NewArrayType(IntType))
	arrayOfOptionalIntProperty := NewPropertyDefinition("scores", "array-optional-int", NewArrayType(NewOptionalType(IntType)))

	mapOfRequiredIntsProperty := NewPropertyDefinition("ratings", "map-string-required-int", NewMapType(StringType, IntType))
	mapOfOptionalIntsProperty := NewPropertyDefinition("ratings", "map-string-required-int", NewMapType(StringType, NewOptionalType(IntType)))

	nastyProperty := NewPropertyDefinition(
		"nasty",
		"my-oh-my",
		NewMapType(
			StringType,
			NewArrayType(
				NewMapType(StringType, BoolType))))

	return []StorageConversionPropertyTestCase{
		{"SetStringFromString", requiredStringProperty, requiredStringProperty},
		{"SetStringFromOptionalString", requiredStringProperty, optionalStringProperty},
		{"SetOptionalStringFromString", optionalStringProperty, requiredStringProperty},
		{"SetOptionalStringFromOptionalString", optionalStringProperty, optionalStringProperty},

		{"SetIntFromInt", requiredIntProperty, requiredIntProperty},
		{"SetIntFromOptionalInt", requiredIntProperty, optionalIntProperty},

		{"SetArrayOfRequiredFromArrayOfRequired", arrayOfRequiredIntProperty, arrayOfRequiredIntProperty},
		{"SetArrayOfRequiredFromArrayOfOptional", arrayOfRequiredIntProperty, arrayOfOptionalIntProperty},
		{"SetArrayOfOptionalFromArrayOfRequired", arrayOfOptionalIntProperty, arrayOfRequiredIntProperty},

		{"SetMapOfRequiredFromMapOfRequired", mapOfRequiredIntsProperty, mapOfRequiredIntsProperty},
		{"SetMapOfRequiredFromMapOfOptional", mapOfRequiredIntsProperty, mapOfOptionalIntsProperty},
		{"SetMapOfOptionalFromMapOfRequired", mapOfOptionalIntsProperty, mapOfRequiredIntsProperty},

		{"NastyTest", nastyProperty, nastyProperty},
	}
}

func TestStorageConversionFunction_AsFuncForDirectConversions(t *testing.T) {
	for _, c := range CreateStorageConversionFunctionTestCases() {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			RunTestStorageConversionFunction_AsFunc(c, true, t)
		})
	}
}

func TestStorageConversionFunction_AsFuncForIndirectConversions(t *testing.T) {
	for _, c := range CreateStorageConversionFunctionTestCases() {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			RunTestStorageConversionFunction_AsFunc(c, false, t)
		})
	}
}

func RunTestStorageConversionFunction_AsFunc(c StorageConversionPropertyTestCase, direct bool, t *testing.T) {
	gm := NewGomegaWithT(t)

	idFactory := NewIdentifierFactory()
	ref := MakeLocalPackageReference("Verification", "vVersion")

	subjectType := NewObjectType().
		WithProperty(c.receiverProperty)

	subjectDefinition := MakeTypeDefinition(
		MakeTypeName(ref, "Person"),
		subjectType)

	stagingTypeName := MakeTypeName(ref, "Party")
	stagingType := NewObjectType().WithProperty(c.otherProperty)
	stagingDefinition := MakeTypeDefinition(stagingTypeName, stagingType)

	hubTypeName := stagingTypeName
	if !direct {
		hubTypeName = MakeTypeName(ref, "Hub")
	}

	convertFrom, errs := NewStorageConversionFromFunction(subjectDefinition, hubTypeName, stagingDefinition, idFactory)
	gm.Expect(errs).To(BeEmpty())

	convertTo, errs := NewStorageConversionToFunction(subjectDefinition, hubTypeName, stagingDefinition, idFactory)
	gm.Expect(errs).To(BeEmpty())

	subjectDefinition = subjectDefinition.WithType(subjectType.WithFunction(convertFrom).WithFunction(convertTo))

	defs := []TypeDefinition{subjectDefinition, stagingDefinition}
	packages := make(map[PackageReference]*PackageDefinition)

	g := goldie.New(t)

	packageDefinition := NewPackageDefinition(ref.Group(), ref.PackageName(), "1")
	for _, def := range defs {
		packageDefinition.AddDefinition(def)
	}
	packages[ref] = packageDefinition

	// put all definitions in one file, regardless.
	// the package reference isn't really used here.
	fileDef := NewFileDefinition(ref, defs, packages)

	buf := &bytes.Buffer{}
	err := fileDef.SaveToWriter(buf)
	if err != nil {
		t.Fatalf("could not generate file: %v", err)
	}

	var testName strings.Builder
	testName.WriteString(c.name)

	if direct {
		testName.WriteString("-Direct")
	} else {
		testName.WriteString("-ViaStaging")
	}

	g.Assert(t, testName.String(), buf.Bytes())
}
