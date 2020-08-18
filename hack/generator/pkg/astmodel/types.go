/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import "fmt"

// Types is the set of all types being generated
type Types map[TypeName]TypeDefinition

// Add adds a type to the set, with safety check that it has not already been defined
func (types Types) Add(def TypeDefinition) {
	key := def.Name()
	if _, ok := types[key]; ok {
		panic(fmt.Sprintf("type already defined: %v", key))
	}

	types[key] = def
}

// AddAll adds multiple types to the set, with the same safety check as Add() to panic if a duplicate is included
func (types Types) AddAll(otherTypes Types) {
	for _, t := range otherTypes {
		types.Add(t)
	}
}

// Where returns a new set of types including only those that satisfy the predicate
func (types Types) Where(predicate func(definition TypeDefinition) bool) Types {
	result := make(Types)
	for _, t := range types {
		if predicate(t) {
			result[t.Name()] = t
		}
	}

	return result
}

// Except returns a new set of types including only those not defined in otherTypes
func (types Types) Except(otherTypes Types) Types {
	return types.Where(func(def TypeDefinition) bool {
		return !otherTypes.Contains(def.Name())
	})
}

// Contains returns true if the set contains a definition for the specified name
func (types Types) Contains(name TypeName) bool {
	_, ok := types[name]
	return ok
}
