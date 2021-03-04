/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestCanMergeSameTypes(t *testing.T) {
	g := NewGomegaWithT(t)

	merger := NewTypeMerger(func(l, r Type) Type {
		t.Fatal("reached fallback")
		return nil
	})

	merger.Add(func(l, r *PrimitiveType) Type {
		return StringType
	})

	result := merger.Merge(IntType, BoolType)

	g.Expect(result).To(Equal(StringType))
}

func TestCanMergeDifferentTypes(t *testing.T) {
	g := NewGomegaWithT(t)

	merger := NewTypeMerger(func(l, r Type) Type {
		return BoolType
	})

	merger.Add(func(l *ObjectType, r *PrimitiveType) Type {
		return StringType
	})

	result := merger.Merge(NewObjectType(), IntType) // shouldn't hit fallback
	g.Expect(result).To(Equal(StringType))

	result2 := merger.Merge(IntType, NewObjectType()) // should hit fallback
	g.Expect(result2).To(Equal(BoolType))
}
