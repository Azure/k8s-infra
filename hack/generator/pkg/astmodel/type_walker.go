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

// TypeWalkerRemoveType is a special TypeName that informs the type walker to remove the property containing this TypeName
// entirely.
var TypeWalkerRemoveType = MakeTypeName(LocalPackageReference{}, "TypeWalkerRemoveProperty")

// TODO: it's awkward to have so much configuration on this thing (3 separate funcs that apply at different places in the walking process?)
// TODO: but unsure if there's a better way... bring it up in code review.

// TypeWalker performs a depth first search across the types provided, applying the visitor to each TypeDefinition.
// MakeContextFunc is called before each visit, and AfterVisitFunc is called after each visit. WalkCycle is called
// if a cycle is detected.
type TypeWalker struct {
	allTypes Types
	visitor  TypeVisitor

	// MakeContextFunc is called before a type is visited.
	MakeContextFunc func(it TypeName, ctx interface{}) (interface{}, error)
	// AfterVisitFunc is called after the type walker has applied the visitor to a TypeDefinition.
	AfterVisitFunc func(original TypeDefinition, updated TypeDefinition, ctx interface{}) (TypeDefinition, error)
	// WalkCycle is called if a cycle is detected. It allows configurable behavior for how to handle cycles.
	WalkCycle func(def TypeDefinition, ctx interface{}) (TypeName, error)

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
		allTypes: allTypes,
	}
	typeWalker.originalVisitTypeName = visitor.VisitTypeName
	typeWalker.originalVisitObjectType = visitor.VisitObjectType

	// visitor is a copy - modifications won't impact passed visitor
	visitor.VisitTypeName = typeWalker.visitTypeName
	visitor.VisitObjectType = typeWalker.visitObjectType

	typeWalker.visitor = visitor
	typeWalker.AfterVisitFunc = DefaultAfterVisit
	typeWalker.MakeContextFunc = IdentityMakeContext
	typeWalker.WalkCycle = IdentityWalkCycle

	return &typeWalker
}

func (t *TypeWalker) visitTypeName(this *TypeVisitor, it TypeName, ctx interface{}) (Type, error) {
	updatedCtx, err := t.MakeContextFunc(it, ctx)
	if err != nil {
		return nil, err
	}

	visitedTypeName, err := t.originalVisitTypeName(this, it, updatedCtx)
	if err != nil {
		return nil, err
	}
	var ok bool
	it, ok = visitedTypeName.(TypeName)
	if !ok {
		panic(fmt.Sprintf("TypeWalker visitor VisitTypeName must return a TypeName, instead returned %T", visitedTypeName))
	}

	def, ok := t.allTypes[it]
	if !ok {
		return nil, errors.Errorf("couldn't find type %q", it)
	}

	// Prevent loops by bypassing this type if it's currently being processed. The processing
	// slice is basically the "path" taken to get to the current type.
	if _, ok := t.state.processing[def.Name()]; ok {
		return t.WalkCycle(def, updatedCtx)
	}
	t.state.processing[def.Name()] = struct{}{}

	updatedType, err := this.Visit(def.Type(), updatedCtx)
	if err != nil {
		return nil, errors.Wrapf(err, "error visiting type %q", def.Name())
	}
	updatedDef, err := t.AfterVisitFunc(def, def.WithType(updatedType), updatedCtx)
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
		isCycle := isRemoveType(prop.PropertyType())
		if isCycle {
			ot = ot.WithoutProperty(prop.PropertyName())
		}
	}

	return ot, nil
}

// Walk returns a Types collection constructed by applying the Visitor to each type in the graph of types reachable
// from the provided TypeDefinition 'def'. Types are visited in a depth-first order. Cycles are not visited.
func (t *TypeWalker) Walk(def TypeDefinition, ctx interface{}) (Types, error) {
	t.state = typeWalkerState{
		result:     make(Types),
		processing: make(map[TypeName]struct{}),
	}

	t.state.processing[def.Name()] = struct{}{}

	updatedType, err := t.visitor.Visit(def.Type(), ctx)
	if err != nil {
		return nil, err
	}
	updatedDef, err := t.AfterVisitFunc(def, def.WithType(updatedType), ctx)
	if err != nil {
		return nil, err
	}

	err = t.state.result.AddWithEqualityCheck(updatedDef)
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

// IdentityWalkCycle is the default cycle walking behavior. It returns the cycle TypeName unmodified (so the cycle is
// not removed or otherwise changed)
func IdentityWalkCycle(def TypeDefinition, _ interface{}) (TypeName, error) {
	return def.Name(), nil
}

func isRemoveType(t Type) bool {
	switch cast := t.(type) {
	case TypeName:
		return cast.Equals(TypeWalkerRemoveType)
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
