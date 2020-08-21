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

// OneOfType represents something that can be any
// one of a number of selected types
type OneOfType struct {
	// invariants:
	// - all types are unique
	// - length > 1
	// - no nested OneOfs (aside from indirectly via TypeName)
	types []Type
}

// MakeOneOfType is a smart constructor for a  OneOfType,
// maintaining the invariants
func MakeOneOfType(types []Type) Type {
	var uniqueTypes []Type
	for _, t := range types {
		if oneOf, ok := t.(OneOfType); ok {
			for _, tInner := range oneOf.types {
				uniqueTypes = appendIfUniqueType(uniqueTypes, tInner)
			}
		} else {
			uniqueTypes = appendIfUniqueType(uniqueTypes, t)
		}
	}

	if len(uniqueTypes) == 1 {
		return uniqueTypes[0]
	}

	return OneOfType{uniqueTypes}
}

func appendIfUniqueType(slice []Type, item Type) []Type {
	for _, r := range slice {
		if r.Equals(item) {
			return slice
		}
	}

	return append(slice, item)
}

var _ Type = OneOfType{}

// Types returns what types the OneOf can be
func (oneOf OneOfType) Types() []Type {
	return oneOf.types
}

// References returns any type referenced by the OneOf types
func (oneOf OneOfType) References() TypeNameSet {
	var result TypeNameSet
	for _, t := range oneOf.types {
		result = SetUnion(result, t.References())
	}

	return result
}

func (oneOf OneOfType) AsType(codeGenerationContext *CodeGenerationContext) ast.Expr {
	panic("should have been replaced by generation time by 'convertAllOfAndOneOf' phase")
}

func (oneOf OneOfType) AsDeclarations(codeGenerationContext *CodeGenerationContext, name TypeName, description []string) []ast.Decl {
	panic("should have been replaced by generation time by 'convertAllOfAndOneOf' phase")
}

// RequiredImports returns the union of the required imports of all the oneOf types
func (oneOf OneOfType) RequiredImports() []PackageReference {
	panic("should have been replaced by generation time by 'convertAllOfAndOneOf' phase")
}

// Equals returns true if the other Type is a OneOfType that contains
// the same set of types
func (oneOf OneOfType) Equals(t Type) bool {

	other, ok := t.(OneOfType)
	if !ok {
		return false
	}

	if len(oneOf.types) != len(other.types) {
		return false
	}

	// compare regardless of ordering
	for _, t := range oneOf.types {
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
func (oneOf OneOfType) String() string {

	var subStrings []string
	for _, t := range oneOf.Types() {
		subStrings = append(subStrings, t.String())
	}

	return fmt.Sprintf("(oneOf: %s)", strings.Join(subStrings, ", "))
}
