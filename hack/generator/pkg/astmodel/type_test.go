package astmodel

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func TestWriteDebugDescription(t *testing.T) {

	here := MakeLocalPackageReference("local", "test", "v1")

	age := MakeTypeName(here, "Age")
	ageDefinition := MakeTypeDefinition(age, IntType)

	personId := MakeTypeName(here, "PersonId")
	personIdDefinition := MakeTypeDefinition(personId, StringType)

	suit := MakeTypeName(here, "Suit")
	diamonds := EnumValue{"diamonds", "Diamonds"}
	hearts := EnumValue{"hearts", "Hearts"}
	clubs := EnumValue{"clubs", "Clubs"}
	spades := EnumValue{"spades", "Spades"}
	suitEnum := NewEnumType(StringType, diamonds, hearts, clubs, spades)
	suitDefinition := MakeTypeDefinition(suit, suitEnum)

	types := make(Types)
	types.Add(ageDefinition)
	types.Add(personIdDefinition)
	types.Add(suitDefinition)

	cases := []struct {
		name     string
		subject  Type
		expected string
	}{
		{"Integer", IntType, "int"},
		{"String", StringType, "string"},
		{"OptionalInteger", NewOptionalType(IntType), "Optional[int]"},
		{"OptionalString", NewOptionalType(StringType), "Optional[string]"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)

			var builder strings.Builder
			c.subject.WriteDebugDescription(&builder, types)
			g.Expect(builder.String()).To(Equal(c.expected))
		})
	}
}
