package astmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewStructDefinition_GivenValues_InitializesFields(t *testing.T) {
	const name = "demo"
	const version = "2020-01-01"

	fullNameField := NewFieldDefinition("fullName", "string").WithDescription("Full legal name")
	familyNameField := NewFieldDefinition("familiyName", "string").WithDescription("Shared family name")
	knownAsField := NewFieldDefinition("knownAs", "string").WithDescription("Commonly known as")

	definition := NewStructDefinition(name, version, &fullNameField, &familyNameField, &knownAsField)

	assert.Equal(t, name, definition.name)
	assert.Equal(t, version, definition.version)
	assert.Equal(t, 3, len(definition.fields))
}

func Test_StructDefinition_Implements_DefinitionInterface(t *testing.T) {
	field := NewStructDefinition("name", "2020-01-01")
	assert.Implements(t, (*Definition)(nil), field)
}

func Test_StructDefinitionAsAst_GivenValidField_ReturnsNonNilResult(t *testing.T) {
	field := NewStructDefinition("name", "2020-01-01")
	node, _ := field.AsAst()
	assert.NotNil(t, node)
}
