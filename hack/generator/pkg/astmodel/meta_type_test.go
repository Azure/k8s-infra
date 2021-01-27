package astmodel

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestAsPrimitiveType(t *testing.T) {

	objectType := NewObjectType()
	arrayType := NewArrayType(StringType)
	mapType := NewMapType(StringType, StringType)
	optionalType := NewOptionalType(arrayType)

	cases := []struct {
		name     string
		subject  Type
		expected Type
	}{
		{"PrimitivesArePrimitives", StringType, StringType},
		{"ObjectAreNotPrimitives", objectType, nil},
		{"ArraysAreNotPrimitives", arrayType, nil},
		{"MapsAreNotPrimitives", mapType, nil},
		{"OptionalAreNotPrimitives", optionalType, nil},
		{"OptionalContainingPrimitive", NewOptionalType(StringType), StringType},
		{"OptionalNotContainingPrimitive", NewOptionalType(objectType), nil},
		{"FlaggedContainingPrimitive", OneOfFlag.ApplyTo(StringType), StringType},
		{"FlaggedNotContainingPrimitive", OneOfFlag.ApplyTo(objectType), nil},
		{"ValidatedContainingPrimitive", MakeValidatedType(StringType, nil), StringType},
		{"ValidatedNotContainingPrimitive", MakeValidatedType(objectType, nil), nil},
		{"ErroredContainingPrimitive", MakeErroredType(StringType, nil, nil), StringType},
		{"ErroredNotContainingPrimitive", MakeErroredType(objectType, nil, nil), nil},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)

			actual := AsPrimitiveType(c.subject)

			if c.expected == nil {
				g.Expect(actual).To(BeNil())
			} else {
				g.Expect(actual).To(Equal(c.expected))
			}

		})
	}
}

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

func TestAsArrayType(t *testing.T) {

	objectType := NewObjectType()
	arrayType := NewArrayType(StringType)
	mapType := NewMapType(StringType, StringType)
	optionalType := NewOptionalType(objectType)

	cases := []struct {
		name     string
		subject  Type
		expected Type
	}{
		{"PrimitivesAreNotArrays", StringType, nil},
		{"ObjectsAreNotArrays", objectType, nil},
		{"ArraysAreArrays", arrayType, arrayType},
		{"MapsAreNotArrays", mapType, nil},
		{"OptionalAreNotArrays", optionalType, nil},
		{"OptionalContainingArray", NewOptionalType(arrayType), arrayType},
		{"OptionalNotContainingArray", NewOptionalType(StringType), nil},
		{"FlaggedContainingArray", OneOfFlag.ApplyTo(arrayType), arrayType},
		{"FlaggedNotContainingArray", OneOfFlag.ApplyTo(StringType), nil},
		{"ValidatedContainingArray", MakeValidatedType(arrayType, nil), arrayType},
		{"ValidatedNotContainingArray", MakeValidatedType(StringType, nil), nil},
		{"ErroredContainingArray", MakeErroredType(arrayType, nil, nil), arrayType},
		{"ErroredNotContainingArray", MakeErroredType(StringType, nil, nil), nil},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)

			actual := AsArrayType(c.subject)

			if c.expected == nil {
				g.Expect(actual).To(BeNil())
			} else {
				g.Expect(actual).To(Equal(c.expected))
			}

		})
	}
}
