package astmodel

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestStructType_Equals_WhenGivenType_ReturnsExpectedResult(t *testing.T) {

	// TODO: Equal to self
	// TODO: Equal to same
	// TODO: Equal if fields in different sequence
	// TODO: Not equal if fields have different names
	// TODO: Not equal to different type
	// TODO: Equal if reference to this struct

	fullNameField := NewFieldDefinition("FullName", "full-name", StringType)
	familyNameField := NewFieldDefinition("FamilyName", "family-name", StringType)
	knownAsField := NewFieldDefinition("KnownAs", "known-as", StringType)
	genderField := NewFieldDefinition("Gender", "gender", StringType)

	personType := NewStructType(fullNameField, familyNameField, knownAsField)
	otherPersonType := NewStructType(fullNameField, familyNameField, knownAsField)
	reorderedType := NewStructType(knownAsField, familyNameField, fullNameField)
	shorterType := NewStructType(knownAsField, fullNameField)
	longerType := NewStructType(fullNameField, familyNameField, knownAsField, genderField)
	mapType := NewMap(StringType, personType)

	cases := []struct {
		name      string
		thisType  Type
		otherType Type
		expected  bool
	}{
		// Expect equal to self
		{"Equal to self", personType, personType, true},
		{"Equal to self", otherPersonType, otherPersonType, true},
		// Expect equal to same
		{"Equal to same", personType, otherPersonType, true},
		{"Equal to same", otherPersonType, personType, true},
		// Expect equal when fields are reordered
		{"Equal when fields reordered", personType, reorderedType, true},
		{"Equal when fields reordered", reorderedType, personType, true},
		// Expect not-equal when fields missing
		{"Not-equal when fields missing", personType, shorterType, false},
		{"Not-equal when fields missing", longerType, personType, false},
		// Expect not-equal when fields added
		{"Not-equal when fields added", personType, longerType, false},
		{"Not-equal when fields added", shorterType, personType, false},
		// Expect not-equal for different type
		{"Not-equal when different type", personType, mapType, false},
		{"Not-equal when different type", mapType, personType, false},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)

			areEqual := c.thisType.Equals(c.otherType)

			g.Expect(areEqual).To(Equal(c.expected))
		})
	}
}
