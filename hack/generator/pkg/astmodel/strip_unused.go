/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"k8s.io/klog/v2"
)

// TypeNameSet stores type names in no particular order without
// duplicates.
type TypeNameSet map[TypeName]bool

// NewTypeNameSet makes a TypeNameSet containing the specified
// names. If no elements are passed it might be nil.
func NewTypeNameSet(initial ...TypeName) TypeNameSet {
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
func (ts TypeNameSet) Add(val TypeName) TypeNameSet {
	if ts == nil {
		ts = make(TypeNameSet)
	}
	ts[val] = true
	return ts
}

// Remove gets rid of the element from the set. If the element was
// already not in the set, nothing changes. Returns the set for
// symmetry with Add.
func (ts TypeNameSet) Remove(val TypeName) TypeNameSet {
	if ts == nil {
		return nil
	}
	delete(ts, val)
	return ts
}

// Contains returns whether this name is in the set. Works for nil
// sets too.
func (ts TypeNameSet) Contains(val TypeName) bool {
	if ts == nil {
		return false
	}
	_, found := ts[val]
	return found
}

// SetUnion returns a new set with all of the names in s1 or s2.
func SetUnion(s1, s2 TypeNameSet) TypeNameSet {
	var result TypeNameSet
	for val := range s1 {
		result = result.Add(val)
	}
	for val := range s2 {
		result = result.Add(val)
	}
	return result
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
		for reference := range def.Type().References() {
			name := def.Name()
			if name == nil {
				klog.V(5).Infof("nil name for %#v", def)
			}
			referrers[reference] = referrers[reference].Add(*def.Name())
		}
	}

	klog.V(0).Infof("definitions: %d, roots: %d, referrers: %d", len(definitions), len(roots), len(referrers))

	checker := newConnectionChecker(roots, referrers)
	var newDefinitions []TypeDefiner
	for _, def := range definitions {
		if def.Name() == nil {
			klog.V(5).Infof("nil name for definition %#v", def)
			continue
		}
		if checker.connected(*def.Name()) {
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
			if def.Name() == nil {
				klog.V(5).Infof("nil name for %#v", def)
			}
			resources.Add(*def.Name())
		}
	}
	return resources
}

func newConnectionChecker(roots TypeNameSet, referrers map[TypeName]TypeNameSet) *connectionChecker {
	return &connectionChecker{
		roots:     roots,
		referrers: referrers,
	}
}

type connectionChecker struct {
	roots     TypeNameSet
	referrers map[TypeName]TypeNameSet

	// memo tracks results for typenames we've already seen.
	memo TypeNameSet
}

func (c *connectionChecker) connected(name TypeName) bool {
	return c.checkWithPath(name, nil)
}

func (c *connectionChecker) checkWithPath(name TypeName, path TypeNameSet) bool {
	if c.roots.Contains(name) {
		return true
	}
	// We don't need to recheck for a type we've seen is connected.
	if c.memo.Contains(name) {
		return true
	}
	path = path.Add(name)
	for referrer := range c.referrers[name] {
		if path.Contains(referrer) {
			// We've already visited this type, don't get caught
			// in a cycle.
			continue
		}
		if c.checkWithPath(referrer, path) {
			// If our referrer is connected to a root, then we are
			// too - track that in memo. Parent callers will
			// record for themselves, so we don't need to store
			// all of path
			c.memo = c.memo.Add(name)
			return true
		}
	}
	path.Remove(name)

	// Note: We can't memoise the negative result here because we
	// don't know whether the search failed because it would have
	// included a cycle. It's possible for this call to fail but a
	// search taking a different path to this node still to succeed
	// and prove that this is connected to a root. (It took me longer
	// than I'd like to understand this.)
	return false
}
