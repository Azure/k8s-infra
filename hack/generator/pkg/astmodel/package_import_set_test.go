/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	. "github.com/onsi/gomega"
	"testing"
)

var (
	simpleTestRef PackageReference = MakeExternalPackageReference("simple")
	pathTestRef   PackageReference = MakeExternalPackageReference("package/path")

	simpleTestImport = NewPackageImport(simpleTestRef)
	pathTestImport   = NewPackageImport(pathTestRef)
)

/*
 * EmptyPackageImportSet() tests
 */

func TestEmptyPackageImportSet_ReturnsEmptySet(t *testing.T) {
	g := NewGomegaWithT(t)
	set := EmptyPackageImportSet()
	g.Expect(set.imports).To(HaveLen(0))
}

/*
 * AddImport() tests
 */

func TestAddImport_WhenImportMissing_IncreasesSizeOfSet(t *testing.T) {
	g := NewGomegaWithT(t)
	set := EmptyPackageImportSet()
	set.AddImport(simpleTestImport)
	g.Expect(set.imports).To(HaveLen(1))
}

func TestAddImport_WhenImportPresent_LeavesSetSameSize(t *testing.T) {
	g := NewGomegaWithT(t)
	set := EmptyPackageImportSet()
	set.AddImport(simpleTestImport)
	set.AddImport(simpleTestImport)
	g.Expect(set.imports).To(HaveLen(1))
}

/*
 * AddReference() tests
 */

func TestAddReference_WhenReferenceMissing_IncreasesSizeOfSet(t *testing.T) {
	g := NewGomegaWithT(t)
	set := EmptyPackageImportSet()
	set.AddReference(simpleTestRef)
	g.Expect(set.imports).To(HaveLen(1))
}

func TestAddImport_WhenReferencePresent_LeavesSetSameSize(t *testing.T) {
	g := NewGomegaWithT(t)
	set := EmptyPackageImportSet()
	set.AddReference(simpleTestRef)
	set.AddReference(simpleTestRef)
	g.Expect(set.imports).To(HaveLen(1))
}

/*
 * Merge() tests
 */

func TestMerge_GivenEmptySet_LeavesSetUnchanged(t *testing.T) {
	g := NewGomegaWithT(t)
	setA := EmptyPackageImportSet()
	setA.AddReference(simpleTestRef)
	setA.AddReference(pathTestRef)
	setB := EmptyPackageImportSet()
	setA.Merge(setB)
	g.Expect(setA.imports).To(HaveLen(2))
}

func TestMerge_GivenIdenticalSet_LeavesSetUnchanged(t *testing.T) {
	g := NewGomegaWithT(t)
	setA := EmptyPackageImportSet()
	setA.AddReference(simpleTestRef)
	setA.AddReference(pathTestRef)
	setB := EmptyPackageImportSet()
	setB.AddReference(simpleTestRef)
	setB.AddReference(pathTestRef)
	setA.Merge(setB)
	g.Expect(setA.imports).To(HaveLen(2))
}

func TestMerge_GivenDisjointSets_MergesSets(t *testing.T) {
	g := NewGomegaWithT(t)
	setA := EmptyPackageImportSet()
	setA.AddReference(simpleTestRef)
	setB := EmptyPackageImportSet()
	setB.AddReference(pathTestRef)
	setA.Merge(setB)
	g.Expect(setA.imports).To(HaveLen(2))
}

/*
 * Contains() tests
 */

func TestContains_GivenMemberOfSet_ReturnsTrue(t *testing.T) {
	g := NewGomegaWithT(t)
	set := EmptyPackageImportSet()
	set.AddImport(simpleTestImport)
	g.Expect(set.ContainsImport(simpleTestImport)).To(BeTrue())
}

func TestContains_GivenNonMemberOfSet_ReturnsFalse(t *testing.T) {
	g := NewGomegaWithT(t)
	set := EmptyPackageImportSet()
	set.AddImport(simpleTestImport)
	g.Expect(set.ContainsImport(pathTestImport)).To(BeFalse())
}

/*
 * Remove() tests
 */

func TestRemove_WhenItemInSet_RemovesIt(t *testing.T) {
	g := NewGomegaWithT(t)
	set := EmptyPackageImportSet()
	set.AddImport(simpleTestImport)
	set.Remove(simpleTestImport)
	g.Expect(set.ContainsImport(simpleTestImport)).To(BeFalse())
}

func TestRemove_WhenItemNotInSet_LeavesSetWithoutIt(t *testing.T) {
	g := NewGomegaWithT(t)
	set := EmptyPackageImportSet()
	set.AddImport(simpleTestImport)
	set.Remove(pathTestImport)
	g.Expect(set.ContainsImport(pathTestImport)).To(BeFalse())
}
