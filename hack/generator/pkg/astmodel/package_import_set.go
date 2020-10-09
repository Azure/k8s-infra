/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import "sort"

// PackageImportSet represents a set of distinct PackageImport references
type PackageImportSet struct {
	imports map[PackageImport]struct{}
}

// EmptyPackageImportSet creates a new empty set of PackageImport references
func EmptyPackageImportSet() *PackageImportSet {
	return &PackageImportSet{
		imports: make(map[PackageImport]struct{}),
	}
}

// AddImport ensures the set includes the specified import
// Adding an import already present in the set is fine.
func (set *PackageImportSet) AddImport(packageImport PackageImport) {
	set.imports[packageImport] = struct{}{}
}

// AddReference ensures this set includes an import of the specified reference
// Adding a reference already in the set is fine.
func (set *PackageImportSet) AddReference(ref PackageReference) {
	set.AddImport(NewPackageImport(ref))
}

// Merge ensures that all imports specified in other are included
func (set *PackageImportSet) Merge(other *PackageImportSet) {
	for i := range other.imports {
		set.AddImport(i)
	}
}

// Remove ensures the specified item is not present
// Removing an item not in the set is not an error.
func (set *PackageImportSet) Remove(packageImport PackageImport) {
	delete(set.imports, packageImport)
}

// Contains allows checking to see if an import is included
func (set *PackageImportSet) ContainsImport(packageImport PackageImport) bool {
	_, found := set.imports[packageImport]
	return found
}

// ImportFor looks up a package reference and returns its import, if any
func (set *PackageImportSet) ImportFor(ref PackageReference) (PackageImport, bool) {
	for imp := range set.imports {
		if imp.PackageReference.Equals(ref) {
			return imp, true
		}
	}

	return PackageImport{}, false
}

// AsSlice() returns a sorted slice containing all the imports
func (set *PackageImportSet) AsSlice() []PackageImport {
	var result []PackageImport
	for i := range set.imports {
		result = append(result, i)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].String() < result[j].String()
	})

	return result
}

// Length returns the number of unique imports in this set
func (set *PackageImportSet) Length() int {
	return len(set.imports)
}

// ApplyName replaces any existing PackageImport for the specified reference with one using the
// specified name
func (set *PackageImportSet) ApplyName(ref PackageReference, name string) {
	found := false
	// Iterate over a slice as that freezes the list
	for _, imp := range set.AsSlice() {
		if imp.PackageReference.Equals(ref) {
			set.Remove(imp)
			found = true
		}
	}

	if found {
		set.AddImport(NewPackageImport(ref).WithName(name))
	}
}
