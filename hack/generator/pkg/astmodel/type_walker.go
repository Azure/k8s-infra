/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"

	"github.com/pkg/errors"
)

// TODO: This is conceptually kinda close to ReferenceGraph except more powerful

// typeWalkerRemoveType is a special TypeName that informs the type walker to remove the property containing this TypeName
// entirely.
var typeWalkerRemoveType = MakeTypeName(LocalPackageReference{}, "TypeWalkerRemoveProperty")

// TODO: it's awkward to have so much configuration on this thing (3 separate funcs that apply at different places in the walking process?)
// TODO: but unsure if there's a better way... bring it up in code review.

// TypeWalker performs a depth first traversal across the types provided, applying the visitor to each TypeDefinition.
// MakeContext is called before each visit, and AfterVisit is called after each visit. RemoveCycle is called
// if a cycle is detected.
type TypeWalker struct {
	allTypes Types
	visitor  TypeVisitor

	// MakeContext is called before any type but the root type is visited. It is given the current context and returns
	// a new context for use in the upcoming Visit.
	MakeContext func(it TypeName, ctx interface{}) (interface{}, error)
	// AfterVisit is called after the type walker has applied the visitor to a TypeDefinition.
	AfterVisit func(original TypeDefinition, updated TypeDefinition, ctx interface{}) (TypeDefinition, error)
	// RemoveCycle is called if a cycle is detected. If true is returned the cycle will be pruned, otherwise it will be preserved as-is.
	RemoveCycle func(def TypeDefinition, ctx interface{}) (bool, error)

	state                   typeWalkerState
	originalVisitTypeName   func(this *TypeVisitor, it TypeName, ctx interface{}) (Type, error)
	originalVisitObjectType func(this *TypeVisitor, it *ObjectType, ctx interface{}) (Type, error)
}

type typeWalkerState struct {
	result     Types
	processing map[TypeName]struct{}
}

// NewTypeWalker returns a TypeWalker.
// The provided visitor VisitTypeName function must return a TypeName and VisitObjectType must return an ObjectType or calls to
// Walk will panic.
func NewTypeWalker(allTypes Types, visitor TypeVisitor) *TypeWalker {
	typeWalker := TypeWalker{
		allTypes:                allTypes,
		originalVisitTypeName:   visitor.VisitTypeName,
		originalVisitObjectType: visitor.VisitObjectType,
		AfterVisit:              DefaultAfterVisit,
		MakeContext:             IdentityMakeContext,
		RemoveCycle:             IdentityRemoveCycle,
	}

	// visitor is a copy - modifications won't impact passed visitor
	visitor.VisitTypeName = typeWalker.visitTypeName
	visitor.VisitObjectType = typeWalker.visitObjectType

	typeWalker.visitor = visitor

	return &typeWalker
}

func (t *TypeWalker) visitTypeName(this *TypeVisitor, it TypeName, ctx interface{}) (Type, error) {
	updatedCtx, err := t.MakeContext(it, ctx)
	if err != nil {
		return nil, err
	}

	visitedTypeName, err := t.originalVisitTypeName(this, it, updatedCtx)
	if err != nil {
		return nil, err
	}
	it, ok := visitedTypeName.(TypeName)
	if !ok {
		panic(fmt.Sprintf("TypeWalker visitor VisitTypeName must return a TypeName, instead returned %T", visitedTypeName))
	}

	def, ok := t.allTypes[it]
	if !ok {
		return nil, errors.Errorf("couldn't find type %q", it)
	}

	// Prevent loops by bypassing this type if it's currently being processed.
	if _, ok := t.state.processing[def.Name()]; ok {
		remove, err := t.RemoveCycle(def, updatedCtx)
		if err != nil {
			return nil, err
		}
		if remove {
			return typeWalkerRemoveType, nil
		} else {
			// Preserve
			return def.Name(), nil
		}
	}
	t.state.processing[def.Name()] = struct{}{}

	updatedType, err := this.Visit(def.Type(), updatedCtx)
	if err != nil {
		return nil, errors.Wrapf(err, "error visiting type %q", def.Name())
	}
	updatedDef, err := t.AfterVisit(def, def.WithType(updatedType), updatedCtx)
	if err != nil {
		return nil, err
	}

	delete(t.state.processing, def.Name())

	err = t.state.result.AddWithEqualityCheck(updatedDef)
	if err != nil {
		return nil, err
	}

	return updatedDef.Name(), nil
}

func (t *TypeWalker) visitObjectType(this *TypeVisitor, it *ObjectType, ctx interface{}) (Type, error) {
	result, err := t.originalVisitObjectType(this, it, ctx)
	if err != nil {
		return nil, err
	}

	ot, ok := result.(*ObjectType)
	if !ok {
		panic(fmt.Sprintf("TypeWalker visitor VisitObjectType must return a *ObjectType, instead returned %T", result))
	}

	for _, prop := range ot.Properties() {
		shouldRemove := isRemoveType(prop.PropertyType())
		if shouldRemove {
			ot = ot.WithoutProperty(prop.PropertyName())
		}
	}

	return ot, nil
}

// Walk returns a Types collection constructed by applying the Visitor to each type in the graph of types reachable
// from the provided TypeDefinition 'def'. Types are visited in a depth-first order. Cycles are not followed
// (so each type in a cycle will be visited only once).
func (t *TypeWalker) Walk(def TypeDefinition) (Types, error) {
	t.state = typeWalkerState{
		result:     make(Types),
		processing: make(map[TypeName]struct{}),
	}

	// Visit our own name to start the walk.
	// Initial ctx is nil -- MakeContext will get called and fabricate a context if needed
	_, err := t.visitor.Visit(def.Name(), nil)
	if err != nil {
		return nil, err
	}

	return t.state.result, nil
}

// DefaultAfterVisit is the default AfterVisit function for TypeWalker. It returns the TypeDefinition from Visit unmodified
func DefaultAfterVisit(_ TypeDefinition, updated TypeDefinition, _ interface{}) (TypeDefinition, error) {
	return updated, nil
}

// IdentityMakeContext returns the context unmodified
func IdentityMakeContext(_ TypeName, ctx interface{}) (interface{}, error) {
	return ctx, nil
}

// IdentityRemoveCycle is the default cycle removal behavior. It preserves all cycles unmodified.
func IdentityRemoveCycle(_ TypeDefinition, _ interface{}) (bool, error) {
	return false, nil
}

func isRemoveType(t Type) bool {
	switch cast := t.(type) {
	case TypeName:
		return cast.Equals(typeWalkerRemoveType)
	case *PrimitiveType:
		return false
	case MetaType:
		return isRemoveType(cast.Unwrap())
	case *ArrayType:
		return isRemoveType(cast.Element())
	case *MapType:
		return isRemoveType(cast.KeyType()) || isRemoveType(cast.ValueType())
	}

	panic(fmt.Sprintf("Unknown Type: %T", t))
}
