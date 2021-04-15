/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package genruntime

// ResolvedReferences is a set of references which have been resolved for a particular resource.
// The special self field is the fully qualified ARM ID of the resource that this ResolvedReferences applies to.
type ResolvedReferences struct {
	// self is the fully qualified name of the resource in Azure ("a/b/c").
	self string
	// references is a map of ResourceReference to ARM ID.
	references map[ResourceReference]string
}

// MakeResolvedReferences creates a ResolvedReferences from the fully qualified ARM ID of the resource and
// and ARM IDs that the resource refers to.
func MakeResolvedReferences(name string, references map[ResourceReference]string) ResolvedReferences {
	if name == "" {
		panic("ResolvedReferences name is required")
	}

	return ResolvedReferences{
		self:       name,
		references: references,
	}
}

// ARMID looks up the fully qualified ARM ID for the given reference. If it cannot be found, false is returned for the second parameter.
func (r ResolvedReferences) ARMID(ref ResourceReference) (string, bool) {
	result, ok := r.references[ref]
	return result, ok
}

// Self returns the fully qualified ARM ID of the resource this ResolvedReferences is associated with.
func (r ResolvedReferences) Self() string {
	return r.self
}
