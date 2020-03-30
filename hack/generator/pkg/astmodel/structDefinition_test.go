package astmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewStructDefinition_GivenValues_InitializesField(t *testing.T) {
	const name = "demo"

	fullNameField := NewFieldDefinition("fullName", "string").WithDescription("Full legal name")
	familyNameField := NewFieldDefinition("familiyName", "string").WithDescription("Shared family name")
	knownAsField := NewFieldDefinition("knownAs", "string").WithDescription("Commonly known as")

	definition := NewStructDefinition(name, &fullNameField, &familyNameField, &knownAsField)

	assert.Equal(t, name, definition.name)
	assert.Equal(t, 3, len(definition.fields))
}

func Test_StructDefinition_Implements_AstGeneratorInterface(t *testing.T) {
	field := NewStructDefinition("name")
	assert.Implements(t, (*AstGenerator)(nil), field)
}

func Test_StructDefinitionAsAst_GivenValidField_ReturnsNonNilResult(t *testing.T) {
	field := NewStructDefinition("name")
	node, _ := field.AsAst()
	assert.NotNil(t, node)
}
