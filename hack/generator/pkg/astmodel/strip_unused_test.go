/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"testing"

	. "github.com/onsi/gomega"
)

const packagePath = "test.package/v1"

func TestConnectionChecker_Avoids_Negative_Memoisation(t *testing.T) {
	g := NewGomegaWithT(t)
	makeName := func(name string) TypeName {
		return *NewTypeName(*NewPackageReference(packagePath), name)
	}

	makeSet := func(names ...string) TypeNameSet {
		var typeNames []TypeName
		for _, n := range names {
			typeNames = append(typeNames, makeName(n))
		}
		return NewTypeNameSet(typeNames...)
	}

	roots := makeSet("res1", "res2")
	referrers := map[TypeName]TypeNameSet{
		makeName("G1"): nil,
		makeName("G2"): makeSet("G1"),
		makeName("A"):  makeSet("res1", "G2", "D"), // cyclic
		makeName("B"):  makeSet("A"),
		makeName("C"):  makeSet("A"),
		makeName("D"):  makeSet("C"),
	}

	cases := []struct {
		name   string
		result bool
	}{
		{"res1", true},
		{"res2", true},
		{"C", true},
		{"D", true},
		{"B", true},
		{"A", true},
		{"G1", false},
		{"G2", false},
	}

	// Check that things work going forward...
	checker := newConnectionChecker(roots, referrers)
	for _, c := range cases {
		t.Logf("forward-%s", c.name)
		g.Expect(checker.connected(makeName(c.name))).To(Equal(c.result))
	}

	// And backward...
	// Recreate the checker so that the memo is clean.
	checker = newConnectionChecker(roots, referrers)
	for i := len(cases) - 1; i >= 0; i-- {
		c := cases[i]
		t.Logf("backward-%s", c.name)
		g.Expect(checker.connected(makeName(c.name))).To(Equal(c.result))
	}
}
