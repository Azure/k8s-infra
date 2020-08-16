package astmodel

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_Visit_GivenCountingTypeVisitor_ReturnsExpectedCounts(t *testing.T) {

	arrType := NewArrayType(StringType)
	mapType := NewMapType(StringType, StringType)
	optType := NewOptionalType(IntType)

	intProp := NewPropertyDefinition("number", "number", IntType)
	arrProp := NewPropertyDefinition("arrType", "arrType", arrType)
	mapProp := NewPropertyDefinition("mapProp", "mapProp", mapType)
	optProp := NewPropertyDefinition("optType", "optType", optType)

	objType := EmptyObjectType.WithProperties(intProp, arrProp, mapProp, optProp)

	resType := NewResourceType(objType, StringType)

	cases := []struct {
		name      string
		theType  Type
		expected  int
	}{
		// Simple types get visited once
		{"Strings are visited once", StringType, 1},
		{"Integers are visited once", IntType, 1},
		{"Booleans are visited once", BoolType, 1},
		// Composite types require more visits
		{"Arrays require two visits", arrType, 2},
		{"Map types require three visits", mapType, 3},
		{"Optional types require two visits", optType, 2},
		// Complex types require even more visits
		{"Object types require one visit plus visiting each property", objType, 9},
		{"Resource types require one visit plus visiting spec and status", resType, 11},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)

			v := MakeCountingTypeVisitor()
			v.Visit(c.theType, nil)

			g.Expect(v.VisitCount).To(Equal(c.expected))
		})
	}
}

type CountingTypeVisitor struct {
	TypeVisitor
	VisitCount int
}

func MakeCountingTypeVisitor() *CountingTypeVisitor {
	v := MakeTypeVisitor()

	result := &CountingTypeVisitor{}

	result.VisitTypeName = func(this *TypeVisitor, it TypeName, ctx interface{}) Type {
		result.VisitCount++
		return v.VisitTypeName(this, it, ctx)
	}

	result.VisitArrayType = func(this *TypeVisitor, it *ArrayType, ctx interface{}) Type {
		result.VisitCount++
		return v.VisitArrayType(this, it, ctx)
	}

	result.VisitPrimitive = func(this *TypeVisitor, it *PrimitiveType, ctx interface{}) Type {
		result.VisitCount++
		return v.VisitPrimitive(this, it, ctx)
	}

	result.VisitObjectType = func(this *TypeVisitor, it *ObjectType, ctx interface{}) Type {
		result.VisitCount++
		return v.VisitObjectType(this, it, ctx)
	}

	result.VisitMapType = func(this *TypeVisitor, it *MapType, ctx interface{}) Type {
		result.VisitCount++
		return v.VisitMapType(this, it, ctx)
	}

	result.VisitEnumType = func(this *TypeVisitor, it *EnumType, ctx interface{}) Type {
		result.VisitCount++
		return v.VisitEnumType(this, it, ctx)
	}

	result.VisitOptionalType = func(this *TypeVisitor, it *OptionalType, ctx interface{}) Type {
		result.VisitCount++
		return v.VisitOptionalType(this, it, ctx)
	}

	result.VisitResourceType = func(this *TypeVisitor, it *ResourceType, ctx interface{}) Type {
		result.VisitCount++
		return v.VisitResourceType(this, it, ctx)
	}

	return result
}
