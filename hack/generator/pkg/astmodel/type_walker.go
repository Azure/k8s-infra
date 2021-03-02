/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"

	"github.com/pkg/errors"
)

// TODO: This shares a lot of concept with ReferenceGraph -- combine somehow?

// TODO: One awkward thing about this is that while it applies the visitor to each item
// TODO: in the graph, it doesn't give us a good way to modify type names (so if we change the structure of an
// TODO: object, we need to after-the-fact go and fix up references to that object. It might be better if we
// TODO: instead did this DFS style with a call in the type visitor VisitTypeName, which could then modify the
// TODO: TypeName it returned in the case that the type it examined was modified. If we do this, we will need
// TODO: external state to track what new types were added/modified


type typeWalkerQueueItem struct {
	typeName TypeName
	ctx      interface{}
}

// TypeWalker applies a TypeVisitor to each type in a type graph. Types referenced from the root
// type are walked in breadth-first order.
type TypeWalker struct {
	allTypes Types
	visitor TypeVisitor

	walkQueue []typeWalkerQueueItem

	WalkFunc func(this *TypeWalker, def TypeDefinition, ctx interface{}) (TypeDefinition, error)
	EnqueueContextFunc func(it TypeName, ctx interface{}) (interface{}, error) // Used to provide per-queue-item context
}

// NewTypeWalker returns a TypeWalker. The provided visitor VisitTypeName function must return a TypeName or calls to
// Walk will panic.
func NewTypeWalker(allTypes Types, visitor TypeVisitor) *TypeWalker {
	typeWalker := TypeWalker{
		allTypes: allTypes,
	}
	// visitor is a copy so modifications are safe here and won't impact passed visitor
	originalVisitTypeName := visitor.VisitTypeName
	visitor.VisitTypeName = func(this *TypeVisitor, it TypeName, ctx interface{}) (Type, error) {
		result, err := originalVisitTypeName(this, it, ctx) // TODO: Type name changing here would also be... awkward. What to do?
		if err != nil {
			return nil, err
		}

		// There's an expectation here that this function returns a type name
		typeName, ok := result.(TypeName)
		if !ok {
			panic(fmt.Sprintf("VisitTypeName did not return a TypeName, instead %q -> %T", it, result))
		}

		updatedCtx, err := typeWalker.EnqueueContextFunc(it, ctx)
		if err != nil {
			return nil, err
		}
		typeWalker.walkQueue = append(typeWalker.walkQueue, typeWalkerQueueItem{typeName: typeName, ctx: updatedCtx})

		return result, nil
	}
	typeWalker.visitor = visitor
	// TODO: Are these really "identity"? Maybe the walk is but the other one maybe not...
	typeWalker.WalkFunc = IdentityWalk
	typeWalker.EnqueueContextFunc = IdentityEnqueueContext

	return &typeWalker
}

// Walk performs a breadth-first walk of the provided type definition and all types referenced by it.
// Each walked definition has the TypeWalker visitor applied to it.
// A type definition is visited once for each occurrence at a unique location in the type graph, but
// can only appear in the result once. Before adding a definition to the result, a check is performed to see if
// a definition with the same name already exists, and an error is returned if a type has the same name as one
// already in the result set but a different structure.
// Cycles in the type graph are allowed. Each type in a cycle is visited only once.
func (t *TypeWalker) Walk(def TypeDefinition, ctx interface{}) (Types, error) {
	result := make(Types)
	visited := make(map[Type]struct{})

	t.walkQueue = append(t.walkQueue, typeWalkerQueueItem{typeName: def.Name(), ctx: ctx})
	for len(t.walkQueue) > 0 {
		// Note: items are added to walkQueue by the visitor when visiting TypeName
		queueItem := t.walkQueue[0]
		t.walkQueue = t.walkQueue[1:]

		def, ok := t.allTypes[queueItem.typeName]
		if !ok {
			return nil, errors.Errorf("unable to find type named %q", queueItem.typeName)
		}

		// Prevent loops to the exact same type-instance (note this isn't Equals based
		// it's reference based). This prevents loops while allowing multiple instances
		// of the same from different parts of the graph to be walked.
		if _, ok := visited[def.Type()]; ok {
			continue
		}

		updatedDef, err := t.WalkFunc(t, def, queueItem.ctx)
		if err != nil {
			return nil, errors.Wrap(err, "walkFunc returned an error")
		}

		visited[def.Type()] = struct{}{}

		err = result.AddWithEqualityCheck(updatedDef)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func IdentityWalk(this *TypeWalker, def TypeDefinition, ctx interface{}) (TypeDefinition, error) {
	updatedType, err := this.visitor.Visit(def.Type(), ctx)
	if err != nil {
		return TypeDefinition{}, err
	}

	return def.WithType(updatedType), nil
}

func IdentityEnqueueContext(_ TypeName, _ interface{}) (interface{}, error) {
	return nil, nil
}
