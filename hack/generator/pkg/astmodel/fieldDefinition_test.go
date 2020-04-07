package astmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewFieldDefinition_GivenValues_InitializesFields(t *testing.T) {
	name := "fullName"
	fieldtype := "string"

	field := NewFieldDefinition(name, fieldtype)

	assert.Equal(t, name, field.name)
	assert.Equal(t, fieldtype, field.fieldType)
	assert.Equal(t, "", field.description)
}

func Test_FieldDefinitionWithDescription_GivenDescription_SetsField(t *testing.T) {
	name := "fullName"
	fieldtype := "string"
	description := "description"

	field := NewFieldDefinition(name, fieldtype).WithDescription(description)

	assert.Equal(t, description, field.description)
}

func Test_FieldDefinitionWithDescription_GivenDescription_DoesNotModifyOriginal(t *testing.T) {
	name := "fullName"
	fieldtype := "string"
	description := "description"

	original := NewFieldDefinition(name, fieldtype)
	field := original.WithDescription(description)

	assert.NotEqual(t, original.description, field.description)
}

func Test_FieldDefinition_Implements_DefinitionInterface(t *testing.T) {
	name := "fullName"
	fieldtype := "string"

	field := NewFieldDefinition(name, fieldtype)

	assert.Implements(t, (*Definition)(nil), field)
}

func Test_FieldDefinitionAsAst_GivenValidField_ReturnsNonNilResult(t *testing.T) {
	name := "fullName"
	fieldtype := "string"

	field := NewFieldDefinition(name, fieldtype)

	node, _ := field.AsAst()
	assert.NotNil(t, node)
}
