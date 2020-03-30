package astmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewFieldDefinition_GivenValues_InitializesFields(t *testing.T) {
	const name = "fullName"
	const fieldtype = "string"
	const description = "Full legal name of a person"

	field := NewFieldDefinition(name, fieldtype, description)

	assert.Equal(t, name, field.name)
	assert.Equal(t, fieldtype, field.fieldType)
	assert.Equal(t, description, field.description)
}

func Test_FieldDefinition_Implements_AstGeneratorInterface(t *testing.T) {
	field := NewFieldDefinition("name", "fieldtype", "description")
	assert.Implements(t, (*AstGenerator)(nil), field)
}

func Test_FieldDefinitionAsAst_GivenValidField_ReturnsNonNilResult(t *testing.T) {
	field := NewFieldDefinition("name", "fieldtype", "description")
	node, _ := field.AsAst()
	assert.NotNil(t, node)
}
