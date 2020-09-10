/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"
	"go/ast"
	"strings"
)

// AllOfType represents something that is the union
// of all the given types
type AllOfType struct {
	// invariants:
	// - all types are unique
	// - length > 1
	// - no nested AllOfs (aside from indirectly via TypeName)
	types []Type
}

// MakeAllOfType is a smart constructor for AllOfType,
// maintaining the invariants
func MakeAllOfType(types []Type) Type {
	var uniqueTypes []Type
	for _, t := range types {
		if allOf, ok := t.(AllOfType); ok {
			for _, tInner := range allOf.types {
				uniqueTypes = appendIfUniqueType(uniqueTypes, tInner)
			}
		} else {
			uniqueTypes = appendIfUniqueType(uniqueTypes, t)
		}
	}

	if len(uniqueTypes) == 1 {
		return uniqueTypes[0]
	}

	// see if there are any OneOfs inside
	var oneOfs []OneOfType
	var notOneOfs []Type
	for _, t := range uniqueTypes {
		if oneOf, ok := t.(OneOfType); ok {
			oneOfs = append(oneOfs, oneOf)
		} else {
			notOneOfs = append(notOneOfs, t)
		}
	}

	if len(oneOfs) == 1 {
		// we want to push AllOf down so that:
		// 		allOf { x, y, oneOf { a, b } }
		// becomes
		//		oneOf { allOf { x, y, a }, allOf { x, y, b } }
		// the latter is much easier to deal with

		var ts []Type
		for _, t := range oneOfs[0].types {
			ts = append(ts, MakeAllOfType(append(notOneOfs, t)))
		}

		return MakeOneOfType(ts)
	}

	// 0 oneOf (nothing to do) or >1 oneOf (too hard)
	return AllOfType{uniqueTypes}
}

var _ Type = AllOfType{}

// Types returns what types the OneOf can be
func (allOf AllOfType) Types() []Type {
	return allOf.types
}

// References returns any type referenced by the AllOf types
func (allOf AllOfType) References() TypeNameSet {
	var result TypeNameSet
	for _, t := range allOf.types {
		result = SetUnion(result, t.References())
	}

	return result
}

// AsType always panics; AllOf cannot be represented by the Go AST and must be
// lowered to an object type
func (allOf AllOfType) AsType(codeGenerationContext *CodeGenerationContext) ast.Expr {
	panic("should have been replaced by generation time by 'convertAllOfAndOneOf' phase")
}

// AsDeclarations always panics; AllOf cannot be represented by the Go AST and must be
// lowered to an object type
func (allOf AllOfType) AsDeclarations(codeGenerationContext *CodeGenerationContext, name TypeName, description []string) []ast.Decl {
	panic("should have been replaced by generation time by 'convertAllOfAndOneOf' phase")
}

// RequiredImports returns the union of the required imports of all the oneOf types
func (allOf AllOfType) RequiredImports() []PackageReference {
	panic("should have been replaced by generation time by 'convertAllOfAndOneOf' phase")
}

// Equals returns true if the other Type is a AllOf that contains
// the same set of types
func (allOf AllOfType) Equals(t Type) bool {

	other, ok := t.(AllOfType)
	if !ok {
		return false
	}

	if len(allOf.types) != len(other.types) {
		return false
	}

	// compare regardless of ordering
	for _, t := range allOf.types {
		found := false
		for _, tOther := range other.types {
			if t.Equals(tOther) {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

// String implements fmt.Stringer
func (allOf AllOfType) String() string {

	var subStrings []string
	for _, t := range allOf.Types() {
		subStrings = append(subStrings, t.String())
	}

	return fmt.Sprintf("(allOf: %s)", strings.Join(subStrings, ", "))
}
