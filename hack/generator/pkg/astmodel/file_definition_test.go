/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"testing"

	. "github.com/onsi/gomega"
)

/*
 * NewFileDefinition tests
 */

func Test_NewFileDefinition_GivenValues_InitializesFields(t *testing.T) {
	g := NewGomegaWithT(t)

	person := NewTestStruct(
		"Person",
		NewStringFieldDefinition("fullName"),
		NewStringFieldDefinition("knownAs"),
		NewStringFieldDefinition("familyName"),
	)
	file := NewFileDefinition(&person.Name().PackageReference, []TypeDefiner{&person}, nil)

	g.Expect(*file.packageReference).To(Equal(person.Name().PackageReference))
	g.Expect(file.definitions).To(HaveLen(1))
}

/*
 * calcRanks() tests
 */

func Test_CalcRanks_GivenMultipleRoots_AssignsRankZeroToAll(t *testing.T) {
	g := NewGomegaWithT(t)

	root1 := NewTestStruct("r1")
	root2 := NewTestStruct("b")
	root3 := NewTestStruct("c")
	root4 := NewTestStruct("d")

	ranks := calcRanks([]TypeDefiner{&root1, &root2, &root3, &root4})

	g.Expect(ranks[*root1.Name()]).To(Equal(0))
	g.Expect(ranks[*root2.Name()]).To(Equal(0))
	g.Expect(ranks[*root3.Name()]).To(Equal(0))
	g.Expect(ranks[*root4.Name()]).To(Equal(0))
}

func Test_CalcRanks_GivenLinearDependencies_AssignsRanksInOrder(t *testing.T) {
	g := NewGomegaWithT(t)

	rank3 := NewTestStruct("d")
	referenceToRank3 := NewFieldDefinition("f3", "f3", rank3.Name())

	rank2 := NewTestStruct("c", referenceToRank3)
	referenceToRank2 := NewFieldDefinition("f2", "f2", rank2.Name())

	rank1 := NewTestStruct("b", referenceToRank2)
	referenceToRank1 := NewFieldDefinition("f1", "f1", rank1.Name())

	rank0 := NewTestStruct("a", referenceToRank1)

	ranks := calcRanks([]TypeDefiner{&rank0, &rank1, &rank2, &rank3})

	g.Expect(ranks[*rank0.Name()]).To(Equal(0))
	g.Expect(ranks[*rank1.Name()]).To(Equal(1))
	g.Expect(ranks[*rank2.Name()]).To(Equal(2))
	g.Expect(ranks[*rank3.Name()]).To(Equal(3))
}

func Test_CalcRanks_GivenDiamondDependencies_AssignRanksInOrder(t *testing.T) {
	g := NewGomegaWithT(t)

	bottom := NewTestStruct("bottom")
	referenceToBottom := NewFieldDefinition("b", "b", bottom.Name())

	left := NewTestStruct("l", referenceToBottom)
	referenceToLeft := NewFieldDefinition("l", "l", left.Name())

	right := NewTestStruct("r", referenceToBottom)
	referenceToRight := NewFieldDefinition("r", "r", right.Name())

	top := NewTestStruct("a", referenceToLeft, referenceToRight)

	ranks := calcRanks([]TypeDefiner{&top, &left, &right, &bottom})

	g.Expect(ranks[*top.Name()]).To(Equal(0))
	g.Expect(ranks[*right.Name()]).To(Equal(1))
	g.Expect(ranks[*left.Name()]).To(Equal(1))
	g.Expect(ranks[*bottom.Name()]).To(Equal(2))
}

func Test_CalcRanks_GivenDiamondWithBar_AssignRanksInOrder(t *testing.T) {
	g := NewGomegaWithT(t)

	bottom := NewTestStruct("bottom")
	referenceToBottom := NewFieldDefinition("b", "b", bottom.Name())

	right := NewTestStruct("r", referenceToBottom)
	referenceToRight := NewFieldDefinition("r", "r", right.Name())

	left := NewTestStruct("l", referenceToBottom, referenceToRight)
	referenceToLeft := NewFieldDefinition("l", "l", left.Name())

	top := NewTestStruct("a", referenceToLeft, referenceToRight)

	ranks := calcRanks([]TypeDefiner{&top, &left, &right, &bottom})

	g.Expect(ranks[*top.Name()]).To(Equal(0))
	g.Expect(ranks[*right.Name()]).To(Equal(1))
	g.Expect(ranks[*left.Name()]).To(Equal(1))
	g.Expect(ranks[*bottom.Name()]).To(Equal(2))
}

func Test_CalcRanks_GivenDiamondWithReverseBar_AssignRanksInOrder(t *testing.T) {
	g := NewGomegaWithT(t)

	bottom := NewTestStruct("bottom")

	referenceToBottom := NewFieldDefinition("b", "b", bottom.Name())
	left := NewTestStruct("l", referenceToBottom)

	referenceToLeft := NewFieldDefinition("l", "l", left.Name())
	right := NewTestStruct("r", referenceToBottom, referenceToLeft)

	referenceToRight := NewFieldDefinition("r", "r", right.Name())
	top := NewTestStruct("a", referenceToLeft, referenceToRight)

	ranks := calcRanks([]TypeDefiner{&top, &left, &right, &bottom})

	g.Expect(ranks[*top.Name()]).To(Equal(0))
	g.Expect(ranks[*right.Name()]).To(Equal(1))
	g.Expect(ranks[*left.Name()]).To(Equal(1))
	g.Expect(ranks[*bottom.Name()]).To(Equal(2))
}

/*
 * Supporting methods
 */

func NewTestStruct(name string, fields ...*FieldDefinition) StructDefinition {
	ref := NewTypeName(*NewLocalPackageReference("group", "2020-01-01"), name)
	definition := NewStructDefinition(ref, NewStructType().WithFields(fields...))

	return *definition
}

func NewStringFieldDefinition(name string) *FieldDefinition {
	return NewFieldDefinition(FieldName(name), name, StringType)
}

