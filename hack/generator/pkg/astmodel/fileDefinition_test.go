package astmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewFileDefinition_GivenValues_InitializesFields(t *testing.T) {
	packageName := "demo"
	person := NewTestStruct("Person", "fullName", "knownAs", "familyName")

	file := NewFileDefinition(packageName, &person)

	assert.Equal(t, packageName, file.packageName)
	assert.Equal(t, 1, len(file.structs))
}

func NewTestStruct(name string, fields ...string) StructDefinition {
	var fs []*FieldDefinition
	for _, n := range fields {
		fs = append(fs, NewFieldDefinition(n, "string"))
	}

	definition := NewStructDefinition(name, fs...)

	return *definition
}
