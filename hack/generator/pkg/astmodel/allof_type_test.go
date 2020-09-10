/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestAllOfOneTypeReturnsThatType(t *testing.T) {
	g := NewGomegaWithT(t)

	oneType := StringType
	result := MakeAllOfType([]Type{oneType})

	g.Expect(result).To(BeIdenticalTo(oneType))
}

func TestAllOfIdenticalTypesReturnsThatType(t *testing.T) {
	g := NewGomegaWithT(t)

	oneType := StringType
	result := MakeAllOfType([]Type{oneType, oneType})

	g.Expect(result).To(BeIdenticalTo(oneType))
}

func TestAllOfFlattensNestedAllOfs(t *testing.T) {
	g := NewGomegaWithT(t)

	result := MakeAllOfType([]Type{BoolType, MakeAllOfType([]Type{StringType, IntType})})

	expected := MakeAllOfType([]Type{BoolType, StringType, IntType})

	g.Expect(result).To(Equal(expected))
}

func TestAllOfOneOfBecomesOneOfAllOf(t *testing.T) {
	g := NewGomegaWithT(t)

	oneOf := MakeOneOfType([]Type{StringType, FloatType})
	result := MakeAllOfType([]Type{BoolType, IntType, oneOf})

	expected := MakeOneOfType([]Type{
		MakeAllOfType([]Type{BoolType, IntType, StringType}),
		MakeAllOfType([]Type{BoolType, IntType, FloatType}),
	})

	g.Expect(result).To(Equal(expected))
}

func TestAllOfEqualityDoesNotCareAboutOrder(t *testing.T) {
	g := NewGomegaWithT(t)

	x := MakeAllOfType([]Type{StringType, BoolType})
	y := MakeAllOfType([]Type{BoolType, StringType})

	g.Expect(x.Equals(y)).To(BeTrue())
	g.Expect(y.Equals(x)).To(BeTrue())
}

func TestAllOfMustHaveAllTypesToBeEqual(t *testing.T) {
	g := NewGomegaWithT(t)

	x := MakeAllOfType([]Type{StringType, BoolType, FloatType})
	y := MakeAllOfType([]Type{BoolType, StringType})

	g.Expect(x.Equals(y)).To(BeFalse())
	g.Expect(y.Equals(x)).To(BeFalse())
}

func TestAllOfsWithDifferentTypesAreNotEqual(t *testing.T) {
	g := NewGomegaWithT(t)

	x := MakeAllOfType([]Type{StringType, FloatType})
	y := MakeAllOfType([]Type{BoolType, StringType})

	g.Expect(x.Equals(y)).To(BeFalse())
	g.Expect(y.Equals(x)).To(BeFalse())
}

func TestAllOfAsTypePanics(t *testing.T) {
	g := NewGomegaWithT(t)

	x := AllOfType{}
	g.Expect(func() {
		x.AsType(&CodeGenerationContext{})
	}).To(PanicWith("should have been replaced by generation time by 'convertAllOfAndOneOf' phase"))
}

func TestAllOfAsDeclarationsPanics(t *testing.T) {
	g := NewGomegaWithT(t)

	x := AllOfType{}
	g.Expect(func() {
		x.AsDeclarations(&CodeGenerationContext{}, TypeName{}, []string{})
	}).To(PanicWith("should have been replaced by generation time by 'convertAllOfAndOneOf' phase"))
}

func TestAllOfRequiredImportsPanics(t *testing.T) {
	g := NewGomegaWithT(t)

	x := AllOfType{}
	g.Expect(func() {
		x.RequiredImports()
	}).To(PanicWith("should have been replaced by generation time by 'convertAllOfAndOneOf' phase"))
}
