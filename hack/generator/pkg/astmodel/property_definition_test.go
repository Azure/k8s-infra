/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"testing"

	. "github.com/onsi/gomega"
)

// Common values for testing
var (
	fieldName        = PropertyName("FullName")
	fieldType        = StringType
	fieldJsonName    = "family-name"
	fieldDescription = "description"
)

/*
 * NewPropertyDefinition() tests
 */

func Test_NewFieldDefinition_GivenValues_ReturnsInstanceWithExpectedFields(t *testing.T) {
	g := NewGomegaWithT(t)

	field := NewPropertyDefinition(fieldName, fieldJsonName, fieldType)

	g.Expect(field.propertyName).To(Equal(fieldName))
	g.Expect(field.propertyType).To(Equal(fieldType))
	g.Expect(field.jsonName).To(Equal(fieldJsonName))
	g.Expect(field.description).To(BeEmpty())
}

func Test_NewFieldDefinition_GivenValues_ReturnsInstanceWithExpectedGetters(t *testing.T) {
	g := NewGomegaWithT(t)

	field := NewPropertyDefinition(fieldName, fieldJsonName, fieldType)

	g.Expect(field.PropertyName()).To(Equal(fieldName))
	g.Expect(field.PropertyType()).To(Equal(fieldType))
}

/*
 * NewEmbeddedStructDefinition() tests
 */

func Test_NewEmbeddedStructDefinition_ReturnsInstanceWithExpectedFields(t *testing.T) {
	g := NewGomegaWithT(t)

	st := NewStructType()
	field := NewEmbeddedStructDefinition(st)

	g.Expect(field.propertyName).To(Equal(PropertyName("")))
	g.Expect(field.propertyType).To(Equal(st))
	g.Expect(field.jsonName).To(Equal(""))
	g.Expect(field.description).To(Equal(""))

}

/*
 * WithDescription() tests
 */

func Test_FieldDefinitionWithDescription_GivenDescription_SetsField(t *testing.T) {
	g := NewGomegaWithT(t)

	field := NewPropertyDefinition(fieldName, fieldJsonName, fieldType).WithDescription(&fieldDescription)

	g.Expect(field.description).To(Equal(fieldDescription))
}

func Test_FieldDefinitionWithDescription_GivenDescription_ReturnsDifferentReference(t *testing.T) {
	g := NewGomegaWithT(t)

	original := NewPropertyDefinition(fieldName, fieldJsonName, fieldType)
	field := original.WithDescription(&fieldDescription)

	g.Expect(field).NotTo(Equal(original))
}

func Test_FieldDefinitionWithDescription_GivenDescription_DoesNotModifyOriginal(t *testing.T) {
	g := NewGomegaWithT(t)

	original := NewPropertyDefinition(fieldName, fieldJsonName, fieldType)
	field := original.WithDescription(&fieldDescription)

	g.Expect(field.description).NotTo(Equal(original.description))
}

func Test_FieldDefinitionWithDescription_GivenNilDescription_SetsDescriptionToEmptyString(t *testing.T) {
	g := NewGomegaWithT(t)

	original := NewPropertyDefinition(fieldName, fieldJsonName, fieldType).WithDescription(&fieldDescription)
	field := original.WithDescription(nil)

	g.Expect(field.description).To(Equal(""))
}

func Test_FieldDefinitionWithDescription_GivenNilDescription_ReturnsDifferentReference(t *testing.T) {
	g := NewGomegaWithT(t)

	original := NewPropertyDefinition(fieldName, fieldJsonName, fieldType).WithDescription(&fieldDescription)
	field := original.WithDescription(nil)

	g.Expect(field).NotTo(Equal(original))
}

func Test_FieldDefinitionWithNoDescription_GivenNilDescription_ReturnsSameReference(t *testing.T) {
	g := NewGomegaWithT(t)

	original := NewPropertyDefinition(fieldName, fieldJsonName, fieldType)
	field := original.WithDescription(nil)

	g.Expect(field).To(Equal(original))
}

func Test_FieldDefinitionWithDescription_GivenSameDescription_ReturnsSameReference(t *testing.T) {
	g := NewGomegaWithT(t)

	original := NewPropertyDefinition(fieldName, fieldJsonName, fieldType).WithDescription(&fieldDescription)
	field := original.WithDescription(&fieldDescription)

	g.Expect(field).To(Equal(original))
}

/*
 * WithType() tests
 */

func Test_FieldDefinitionWithType_GivenNewType_SetsFieldOnResult(t *testing.T) {
	g := NewGomegaWithT(t)

	original := NewPropertyDefinition(fieldName, fieldJsonName, fieldType)
	field := original.WithType(IntType)

	g.Expect(field.propertyType).To(Equal(IntType))
}

func Test_FieldDefinitionWithType_GivenNewType_DoesNotModifyOriginal(t *testing.T) {
	g := NewGomegaWithT(t)

	original := NewPropertyDefinition(fieldName, fieldJsonName, fieldType)
	_ = original.WithType(IntType)

	g.Expect(original.propertyType).To(Equal(fieldType))
}

func Test_FieldDefinitionWithType_GivenNewType_ReturnsDifferentReference(t *testing.T) {
	g := NewGomegaWithT(t)

	original := NewPropertyDefinition(fieldName, fieldJsonName, fieldType)
	field := original.WithType(IntType)

	g.Expect(field).NotTo(Equal(original))
}

func Test_FieldDefinitionWithType_GivenSameType_ReturnsExistingReference(t *testing.T) {
	g := NewGomegaWithT(t)

	original := NewPropertyDefinition(fieldName, fieldJsonName, fieldType)
	field := original.WithType(fieldType)

	g.Expect(field).To(BeIdenticalTo(original))
}

/*
 * MakeRequired() Tests
 */

func Test_FieldDefinition_MakeRequired_ReturnsDifferentReference(t *testing.T) {
	g := NewGomegaWithT(t)

	original := NewPropertyDefinition(fieldName, fieldJsonName, fieldType)
	field := original.MakeRequired()

	g.Expect(field).NotTo(BeIdenticalTo(original))
}

/*
 * MakeTypeOptional() Tests
 */

func Test_FieldDefintionWithRequiredType_MakeTypeOptional_ReturnsDifferentReference(t *testing.T) {
	g := NewGomegaWithT(t)

	original := NewPropertyDefinition(fieldName, fieldJsonName, fieldType)
	field := original.MakeTypeOptional()

	g.Expect(field).NotTo(BeIdenticalTo(original))
}

func Test_FieldDefintionWithOptionalType_MakeTypeOptional_ReturnsExistingReference(t *testing.T) {
	g := NewGomegaWithT(t)

	original := NewPropertyDefinition(fieldName, fieldJsonName, fieldType).MakeTypeOptional()
	field := original.MakeTypeOptional()

	g.Expect(field).To(BeIdenticalTo(original))
}

/*
 * AsAst() Tests
 */

func Test_FieldDefinitionAsAst_GivenValidField_ReturnsNonNilResult(t *testing.T) {
	g := NewGomegaWithT(t)

	field := NewPropertyDefinition(fieldName, fieldJsonName, fieldType).
		MakeRequired().
		WithDescription(&fieldDescription)

	node := field.AsField(nil)

	g.Expect(node).NotTo(BeNil())
}

/*
 * Equals Tests
 */

func TestFieldDefinition_Equals_WhenGivenFieldDefinition_ReturnsExpectedResult(t *testing.T) {

	strField := createStringProperty("FullName", "Full Legal Name")
	otherStrField := createStringProperty("FullName", "Full Legal Name")

	intField := CreateIntProperty("Age", "Age at last birthday")

	differentName := createStringProperty("Name", "Full Legal Name")
	differentType := CreateIntProperty("FullName", "Full Legal Name")
	differentDescription := CreateIntProperty("FullName", "The whole thing")

	cases := []struct {
		name       string
		thisField  *PropertyDefinition
		otherField *PropertyDefinition
		expected   bool
	}{
		// Expect equal to self
		{"Equal to self", strField, strField, true},
		{"Equal to self", intField, intField, true},
		// Expect equal to same
		{"Equal to same", strField, otherStrField, true},
		{"Equal to same", otherStrField, strField, true},
		// Expect not-equal when properties are different
		{"Not-equal if names are different", strField, differentName, false},
		{"Not-equal if types are different", strField, differentType, false},
		{"Not-equal if descriptions are different", strField, differentDescription, false},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)

			areEqual := c.thisField.Equals(c.otherField)

			g.Expect(areEqual).To(Equal(c.expected))
		})
	}
}
