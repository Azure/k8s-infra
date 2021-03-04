/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

func TestCanMergeSameTypes(t *testing.T) {
	g := NewGomegaWithT(t)

	merger := NewTypeMerger(func(_ interface{}, l, r Type) (Type, error) {
		return nil, errors.New("reached fallback")
	})

	merger.Add(func(l, r *PrimitiveType) (Type, error) {
		return StringType, nil
	})

	result, err := merger.Merge(IntType, BoolType)
	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(result).To(Equal(StringType))
}

func TestCanMergeDifferentTypes(t *testing.T) {
	g := NewGomegaWithT(t)

	merger := NewTypeMerger(func(_ interface{}, l, r Type) (Type, error) {
		return BoolType, nil
	})

	merger.Add(func(l *ObjectType, r *PrimitiveType) (Type, error) {
		return StringType, nil
	})

	result, err := merger.Merge(NewObjectType(), IntType) // shouldn't hit fallback
	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(result).To(Equal(StringType))

	result, err = merger.Merge(IntType, NewObjectType()) // should hit fallback
	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(result).To(Equal(BoolType))
}

func TestCanMergeWithGenericTypeArgument(t *testing.T) {
	g := NewGomegaWithT(t)

	merger := NewTypeMerger(func(_ interface{}, l, r Type) (Type, error) {
		return BoolType, nil
	})

	merger.Add(func(l *ObjectType, r Type) (Type, error) {
		return StringType, nil
	})

	result, err := merger.Merge(NewObjectType(), IntType) // shouldn't hit fallback
	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(result).To(Equal(StringType))

	result, err = merger.Merge(IntType, NewObjectType()) // should hit fallback
	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(result).To(Equal(BoolType))
}

func TestCanMergeWithUnorderedMerger(t *testing.T) {
	g := NewGomegaWithT(t)

	merger := NewTypeMerger(func(_ interface{}, l, r Type) (Type, error) {
		return BoolType, nil
	})

	merger.AddUnordered(func(l *ObjectType, r Type) (Type, error) {
		return StringType, nil
	})

	result, err := merger.Merge(NewObjectType(), IntType) // shouldn't hit fallback
	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(result).To(Equal(StringType))

	result, err = merger.Merge(IntType, NewObjectType()) // shouldn't hit fallback either
	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(result).To(Equal(StringType))
}
