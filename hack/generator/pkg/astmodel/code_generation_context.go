/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

// CodeGenerationContext stores context about the location code-generation is occurring.
// This is required because some things (such as specific field types) are impacted by the context
// in which the field declaration occurs - for example in a file with two conflicting package references
// a disambiguation must occur and field types must ensure they correctly refer to the disambiguated types
type CodeGenerationContext struct {
	packageReferences map[PackageReference]struct{}
	currentPackage    *PackageReference
}

// TODO: How picky about immutability should we be -- should we put this in its own package?

// New CodeGenerationContext creates a new immutable code generation context
func NewCodeGenerationContext(currentPackage *PackageReference, packageReferences map[PackageReference]struct{}) *CodeGenerationContext {
	return &CodeGenerationContext{currentPackage: currentPackage, packageReferences: packageReferences}
}

// PackageReferences returns the set of package references in the current context
func (codeGenContext *CodeGenerationContext) PackageReferences() map[PackageReference]struct{} {
	// return a copy of the map to ensure immutability
	result := make(map[PackageReference]struct{})

	for key, value := range codeGenContext.packageReferences {
		result[key] = value
	}
	return result
}
