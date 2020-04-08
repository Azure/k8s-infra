package astmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewFieldDefinition_GivenValues_InitializesFields(t *testing.T) {
	g := NewGomegaWithT(t)

	fieldName := "FullName"
	fieldtype := StringType
	jsonName := "family-name"

	field := NewFieldDefinition(fieldName, jsonName, fieldtype)

	assert.Equal(t, name, field.name)
	assert.Equal(t, fieldtype, field.fieldType)
	assert.Equal(t, "", field.description)
}

func Test_FieldDefinitionWithDescription_GivenDescription_SetsField(t *testing.T) {
	description := "description"
	field := NewFieldDefinition("FullName", "fullname", StringType).WithDescription(&description)

	assert.Equal(t, description, field.description)
}

func Test_FieldDefinitionWithDescription_GivenDescription_DoesNotModifyOriginal(t *testing.T) {
	description := "description"
	original := NewFieldDefinition("FullName", "fullName", StringType)

	field := original.WithDescription(&description)

	assert.NotEqual(t, original.description, field.description)
}

func Test_FieldDefinition_Implements_DefinitionInterface(t *testing.T) {
	g := NewGomegaWithT(t)

	var fieldDefinition interface{} = NewFieldDefinition("FullName", "fullName", StringType)

	definition, ok := fieldDefinition.(Definition)

	g.Expect(ok).To(BeTrue())
	g.Expect(definition).NotTo(BeNil())
}

func Test_FieldDefinitionAsAst_GivenValidField_ReturnsNonNilResult(t *testing.T) {
	g := NewGomegaWithT(t)

	field := NewFieldDefinition("FullName", "fullName", StringType)
	node := field.AsAst()

	assert.NotNil(t, node)
}
