/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestOneOfOneTypeReturnsThatType(t *testing.T) {
	g := NewGomegaWithT(t)

	oneType := StringType
	result := MakeOneOfType(oneType)

	g.Expect(result).To(BeIdenticalTo(oneType))
}

func TestOneOfIdenticalTypesReturnsThatType(t *testing.T) {
	g := NewGomegaWithT(t)

	oneType := StringType
	result := MakeOneOfType(oneType, oneType)

	g.Expect(result).To(BeIdenticalTo(oneType))
}

func TestOneOfFlattensNestedOneOfs(t *testing.T) {
	g := NewGomegaWithT(t)

	result := MakeOneOfType(BoolType, MakeOneOfType(StringType, IntType))

	expected := MakeOneOfType(BoolType, StringType, IntType)

	g.Expect(result).To(Equal(expected))
}

func TestOneOfEqualityDoesNotCareAboutOrder(t *testing.T) {
	g := NewGomegaWithT(t)

	x := MakeOneOfType(StringType, BoolType)
	y := MakeOneOfType(BoolType, StringType)

	g.Expect(x.Equals(y)).To(BeTrue())
	g.Expect(y.Equals(x)).To(BeTrue())
}

func TestOneOfMustHaveAllTypesToBeEqual(t *testing.T) {
	g := NewGomegaWithT(t)

	x := MakeOneOfType(StringType, BoolType, FloatType)
	y := MakeOneOfType(BoolType, StringType)

	g.Expect(x.Equals(y)).To(BeFalse())
	g.Expect(y.Equals(x)).To(BeFalse())
}

func TestOneOfsWithDifferentTypesAreNotEqual(t *testing.T) {
	g := NewGomegaWithT(t)

	x := MakeOneOfType(StringType, FloatType)
	y := MakeOneOfType(BoolType, StringType)

	g.Expect(x.Equals(y)).To(BeFalse())
	g.Expect(y.Equals(x)).To(BeFalse())
}

var expectedOneOfPanic = CreateCodeGenerationPanic("OneOfType should have been replaced by generation time by 'convertAllOfAndOneOf' phase")

func TestOneOfAsTypePanics(t *testing.T) {
	g := NewGomegaWithT(t)

	x := OneOfType{}
	g.Expect(func() {
		x.AsType(&CodeGenerationContext{})
	}).To(PanicWith(expectedOneOfPanic))
}

func TestOneOfAsDeclarationsPanics(t *testing.T) {
	g := NewGomegaWithT(t)

	x := OneOfType{}
	g.Expect(func() {
		x.AsDeclarations(&CodeGenerationContext{}, DeclarationContext{})
	}).To(PanicWith(expectedOneOfPanic))
}

func TestOneOfRequiredImportsPanics(t *testing.T) {
	g := NewGomegaWithT(t)

	x := OneOfType{}
	g.Expect(func() {
		x.RequiredPackageReferences()
	}).To(PanicWith(expectedOneOfPanic))
}
