/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import "github.com/pkg/errors"

// TypeNameSet stores type names in no particular order without
// duplicates.
type TypeNameSet map[TypeName]bool

// NewTypeNameSet makes a TypeNameSet containing the specified
// names. If no elements are passed it might be nil.
func NewTypeNameSet(initial ...*TypeName) TypeNameSet {
	var result TypeNameSet
	for _, name := range initial {
		result = result.Add(name)
	}
	return result
}

// Add includes the passed name in the set and returns the updated
// set, so that adding can work for a nil set - this makes it more
// convenient to add to sets kept in a map (in the way you might with
// a map of slices).
// TODO(babbageclunk): I like the idea of making the zero element
// useful so a map works nicely, and the analogue with appending to a
// slice, but the mutating and returning behaviour might be too weird.
func (ts TypeNameSet) Add(val *TypeName) TypeNameSet {
	if val == nil {
		return ts
	}
	if ts == nil {
		ts = make(TypeNameSet)
	}
	ts[*val] = true
	return ts
}

// Remove gets rid of the element from the set. If the element was
// already not in the set, nothing changes. Returns the set for
// symmetry with Add.
func (ts TypeNameSet) Remove(val *TypeName) TypeNameSet {
	if ts == nil {
		return nil
	}
	if val == nil {
		return ts
	}
	delete(ts, *val)
	return ts
}

// Contains returns whether this name is in the set. Works for nil
// sets too.
func (ts TypeNameSet) Contains(val *TypeName) bool {
	if ts == nil || val == nil {
		return false
	}
	_, found := ts[*val]
	return found
}

// StripUnusedDefinitions removes all types that aren't in roots or
// referred to by the types in roots, for example types that are
// generated as a byproduct of an allOf element.
func StripUnusedDefinitions(
	roots TypeNameSet,
	definitions []TypeDefiner,
) ([]TypeDefiner, error) {
	// Build a referrers map for each type.
	referrers := make(map[TypeName]TypeNameSet)

	for _, def := range definitions {
		for _, referee := range def.Type().Referees() {
			if referee == nil {
				return nil, errors.Errorf("nil referee for %s", def.Name())
			}
			refereeVal := *referee
			referrers[refereeVal] = referrers[refereeVal].Add(def.Name())
		}
	}

	checker := newConnectionChecker(roots, referrers)
	var newDefinitions []TypeDefiner
	for _, def := range definitions {
		if checker.connected(def.Name()) {
			newDefinitions = append(newDefinitions, def)
		}
	}
	return newDefinitions, nil
}

// CollectResourceDefinitions returns a TypeNameSet of all of the
// resource definitions in the definitions passed in.
func CollectResourceDefinitions(definitions []TypeDefiner) TypeNameSet {
	resources := make(TypeNameSet)
	for _, def := range definitions {
		if _, ok := def.(*ResourceDefinition); ok {
			resources.Add(def.Name())
		}
	}
	return resources
}

func newConnectionChecker(roots TypeNameSet, referrers map[TypeName]TypeNameSet) *connectionChecker {
	return &connectionChecker{
		roots:     roots,
		referrers: referrers,
		memo:      make(map[TypeName]bool),
	}
}

type connectionChecker struct {
	roots     TypeNameSet
	referrers map[TypeName]TypeNameSet

	// memo tracks results for typenames we've already seen - both
	// positive and negative, which is why it's not a TypeNameSet.
	// TODO(babbageclunk): see how much of a difference this
	// makes. Maybe the chains are all pretty shallow?
	memo map[TypeName]bool
}

func (c *connectionChecker) connected(name *TypeName) bool {
	return c.checkWithPath(name, nil)
}

func (c *connectionChecker) checkWithPath(name *TypeName, path TypeNameSet) bool {
	if name == nil {
		return false
	}
	if c.roots.Contains(name) {
		return true
	}
	// We don't need to recheck for a type we've seen before.
	if result, found := c.memo[*name]; found {
		return result
	}
	for referrer := range c.referrers[*name] {
		ref := &referrer
		if path.Contains(ref) {
			// We've already visited this type, don't get caught
			// in a cycle.
			continue
		}
		path = path.Add(ref)
		if c.checkWithPath(ref, path) {
			// If our referrer is connected to a root, then we are
			// too - track that in memo. Parent callers will
			// record for themselves, so we don't need to store
			// all of path.
			c.memo[*name] = true
			return true
		}
		path = path.Remove(ref)
	}
	c.memo[*name] = false
	return false
}
