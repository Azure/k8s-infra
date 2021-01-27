package astmodel

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestAsObjectType(t *testing.T) {

	objectType := NewObjectType()
	arrayType := NewArrayType(StringType)
	mapType := NewMapType(StringType, StringType)
	optionalType := NewOptionalType(StringType)

	cases := []struct {
		name     string
		subject  Type
		expected Type
	}{
		{"PrimitivesAreNotObjects", StringType, nil},
		{"ObjectsAreObjects", objectType, objectType},
		{"ArraysAreNotObjects", arrayType, nil},
		{"MapsAreNotObjects", mapType, nil},
		{"OptionalAreNotObjects", optionalType, nil},
		{"OptionalContainingObject", NewOptionalType(objectType), objectType},
		{"OptionalNotContainingObject", NewOptionalType(StringType), nil},
		{"FlaggedContainingObject", OneOfFlag.ApplyTo(objectType), objectType},
		{"FlaggedNotContainingObject", OneOfFlag.ApplyTo(StringType), nil},
		{"ValidatedContainingObject", MakeValidatedType(objectType, nil), objectType},
		{"ValidatedNotContainingObject", MakeValidatedType(StringType, nil), nil},
		{"ErroredContainingObject", MakeErroredType(objectType, nil, nil), objectType},
		{"ErroredNotContainingObject", MakeErroredType(StringType, nil, nil), nil},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)

			actual := AsObjectType(c.subject)

			if c.expected == nil {
				g.Expect(actual).To(BeNil())
			} else {
				g.Expect(actual).To(Equal(c.expected))
			}

		})
	}
}
