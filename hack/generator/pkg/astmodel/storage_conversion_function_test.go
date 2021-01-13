package astmodel

import (
	"bytes"
	"github.com/sebdah/goldie/v2"
	"strings"
	"testing"
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

	return []StorageConversionPropertyTestCase{
		{"SetStringFromString", requiredStringProperty, requiredStringProperty},
		{"SetStringFromOptionalString", requiredStringProperty, optionalStringProperty},
		{"SetOptionalStringFromString", optionalStringProperty, requiredStringProperty},
		{"SetOptionalStringFromOptionalString", optionalStringProperty, optionalStringProperty},

		{"SetIntFromInt", requiredIntProperty, requiredIntProperty},
		{"SetIntFromIntString", requiredIntProperty, optionalIntProperty},
	}
}

func TestStorageConversionFunction_AsFuncForDirectConvertFrom(t *testing.T) {
	for _, c := range CreateStorageConversionFunctionTestCases() {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			RunTestStorageConversionFunction_AsFunc(c, ConvertFrom, true, t)
		})
	}
}

func TestStorageConversionFunction_AsFuncForDirectConvertTo(t *testing.T) {
	for _, c := range CreateStorageConversionFunctionTestCases() {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			RunTestStorageConversionFunction_AsFunc(c, ConvertTo, true, t)
		})
	}
}

func TestStorageConversionFunction_AsFuncForIndirectConvertFrom(t *testing.T) {
	for _, c := range CreateStorageConversionFunctionTestCases() {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			RunTestStorageConversionFunction_AsFunc(c, ConvertFrom, false, t)
		})
	}
}

func TestStorageConversionFunction_AsFuncForIndirectConvertTo(t *testing.T) {
	for _, c := range CreateStorageConversionFunctionTestCases() {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			RunTestStorageConversionFunction_AsFunc(c, ConvertTo, false, t)
		})
	}
}

func RunTestStorageConversionFunction_AsFunc(c StorageConversionPropertyTestCase, direction StorageConversionDirection, direct bool, t *testing.T) {

	idFactory := NewIdentifierFactory()
	ref := MakeLocalPackageReference("Verification", "vVersion")

	subjectType := NewObjectType().
		WithProperty(c.receiverProperty)

	subjectDefinition := MakeTypeDefinition(
		MakeTypeName(ref, "Subject"),
		subjectType)

	stagingTypeName := MakeTypeName(ref, "Staging")
	stagingType := NewObjectType().WithProperty(c.otherProperty)
	stagingDefinition := MakeTypeDefinition(stagingTypeName, stagingType)

	hubTypeName := stagingTypeName
	if !direct {
		hubTypeName = MakeTypeName(ref, "Hub")
	}

	var fn *StorageConversionFunction
	if direction == ConvertFrom {
		fn = NewStorageConversionFromFunction(subjectDefinition, hubTypeName, stagingDefinition, idFactory)
	} else {
		fn = NewStorageConversionToFunction(subjectDefinition, hubTypeName, stagingDefinition, idFactory)
	}

	subjectDefinition = subjectDefinition.WithType(subjectType.WithFunction(fn))

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
	testName.WriteString(fn.name)
	if direct {
		testName.WriteString("-Direct-")
	} else {
		testName.WriteString("-ViaStaging-")
	}
	testName.WriteString(c.name)

	g.Assert(t, testName.String(), buf.Bytes())
}
