/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	. "github.com/onsi/gomega"
	"testing"
)

/*
 * Equals() tests
 */

func Test_ResourceEquals(t *testing.T) {
	resourceA := NewResourceType(StringType, StringType)
	resourceB := NewResourceType(StringType, IntType)
	resourceC := NewResourceType(IntType, StringType)
	resourceD := NewResourceType(StringType, StringType)

	cases := []struct {
		name          string
		left          *ResourceType
		right         *ResourceType
		expectedEqual bool
	}{
		// Always equal to same reference, and to same underlying types
		{"Equal to self", resourceA, resourceA, true},
		{"Equal to self", resourceB, resourceB, true},
		{"Equal to same values", resourceA, resourceD, true},
		// nil comparisons
		{"Two nil values are equal", nil, nil, true},
		{"Resource and nil values are inequal", resourceA, nil, false},
		{"Nil and resource values are also inequal", nil, resourceA, false},
		// Sensitive to value differences
		{"Sensitive to spec type", resourceA, resourceC, false},
		{"Sensitive to status type", resourceA, resourceB, false},
		{"Sensitive to storage version", resourceA, resourceA.MarkAsStorageVersion(), false},
		{"Sensitive to flag", resourceA, resourceA.AddFlag("foo"), false},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)

			areEqual := c.left.Equals(c.right)

			g.Expect(areEqual).To(Equal(c.expectedEqual))
		})
	}
}

/*
 * AddFlag() tests
 */

func Test_ResourceAddFlag_GivenFlag_ReturnsDifferentInstance(t *testing.T) {
	g := NewGomegaWithT(t)

	resource := NewResourceType(StringType, StringType)
	flagged := resource.AddFlag("foo")
	g.Expect(flagged).NotTo(BeIdenticalTo(resource))
}

/*
 * HasFlag() tests
 */

func Test_ResourceHasFlag_GivenResourceWithFlag_ReturnsTrue(t *testing.T) {
	g := NewGomegaWithT(t)

	resource := NewResourceType(StringType, StringType).AddFlag("foo")

	g.Expect(resource.HasFlag("foo")).To(BeTrue())
}

func Test_ResourceHasFlag_GivenResourceWithTwoFlags_ReturnsTrueForBoth(t *testing.T) {
	g := NewGomegaWithT(t)

	resource := NewResourceType(StringType, StringType).AddFlag("foo").AddFlag("bar")

	g.Expect(resource.HasFlag("foo")).To(BeTrue())
	g.Expect(resource.HasFlag("bar")).To(BeTrue())
}
