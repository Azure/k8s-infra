package astmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewStructDefinition_GivenValues_InitializesFields(t *testing.T) {
	const name = "demo"
	const version = "2020-01-01"
	fullNameField := createStringField("fullName", "Full legal name")
	familyNameField := createStringField("familiyName", "Shared family name")
	knownAsField := createStringField("knownAs", "Commonly known as")

	definition := NewStructDefinition(name, version, fullNameField, familyNameField, knownAsField)

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
	node := field.AsAst()

	g.Expect(node).NotTo(BeNil())
}

func createStringField(name string, description string) *FieldDefinition {
	return NewFieldDefinition(name, name, StringType).WithDescription(&description)
}
