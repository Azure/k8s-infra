/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"testing"

	. "github.com/onsi/gomega"
)

var (
	pkg             = MakePackageReference("foo")
	alphaDefinition = createTestDefinition("alpha")
	betaDefinition  = createTestDefinition("beta")
	gammaDefinition = createTestDefinition("gamma")
	deltaDefinition = createTestDefinition("delta")
)

/*
 * Add() Tests
 */

func Test_TypesAdd_GivenType_ModifiesSet(t *testing.T) {
	g := NewGomegaWithT(t)

	types := make(Types)
	types.Add(alphaDefinition)

	g.Expect(types).To(ContainElement(alphaDefinition))
}

func Test_TypesAdd_GivenTypeAlreadyPresent_Panics(t *testing.T) {
	g := NewGomegaWithT(t)

	types := make(Types)
	types.Add(alphaDefinition)

	g.Expect(func() { types.Add(alphaDefinition) }).To(Panic())
}

/*
 * AddAll() Tests
 */

func Test_TypesAddAll_GivenTypes_ModifiesSet(t *testing.T) {
	g := NewGomegaWithT(t)

	types := createTestTypes(alphaDefinition, betaDefinition)
	otherTypes := createTestTypes(gammaDefinition, deltaDefinition)

	types.AddAll(otherTypes)

	g.Expect(types).To(ContainElements(gammaDefinition, deltaDefinition))
}

func Test_TypesAddAll_GivenOverlappingTypes_Panics(t *testing.T) {
	g := NewGomegaWithT(t)

	types := createTestTypes(alphaDefinition, betaDefinition)
	otherTypes := createTestTypes(betaDefinition, deltaDefinition)

	g.Expect(func() { types.AddAll(otherTypes) }).To(Panic())
}

/*
 * Where() Tests
 */

func Test_TypesWhere_GivenPredicate_ReturnsExpectedSet(t *testing.T) {
	g := NewGomegaWithT(t)

	types := createTestTypes(alphaDefinition, betaDefinition, gammaDefinition, deltaDefinition).
		Where(func(def TypeDefinition) bool {
			return len(def.name.name) == 5
		})

	g.Expect(types).To(ContainElements(alphaDefinition, gammaDefinition, deltaDefinition))
	g.Expect(types).NotTo(ContainElement(betaDefinition))
}

/*
 * Utility functions
 */

func createTestDefinition(name string) TypeDefinition {
	n := MakeTypeName(pkg, name)
	return MakeTypeDefinition(n, StringType)
}

func createTestTypes(defs ...TypeDefinition) Types {
	result := make(Types)
	for _, d := range defs {
		result.Add(d)
	}

	return result
}
