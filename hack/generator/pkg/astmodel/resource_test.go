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
		{"Resource and nil values are inequal", nil, resourceA, false},
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
