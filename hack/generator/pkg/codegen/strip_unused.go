/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"fmt"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
)

type TypeNameSet map[astmodel.TypeName]bool

func NewTypeNameSet(initial ...*astmodel.TypeName) TypeNameSet {
	var result TypeNameSet
	for _, name := range initial {
		result = result.Add(name)
	}
	return result
}

func (ts TypeNameSet) Add(val *astmodel.TypeName) TypeNameSet {
	if val == nil {
		return ts
	}
	if ts == nil {
		ts = make(TypeNameSet)
	}
	ts[*val] = true
	return ts
}

func (ts TypeNameSet) Contains(val *astmodel.TypeName) bool {
	if ts == nil || val == nil {
		return false
	}
	_, found := ts[*val]
	return found
}

// StripUnusedDefinitions removes all types that aren't top-level or
// referred to by fields in other types, for example types that are
// generated as a byproduct of an allOf element.
func StripUnusedDefinitions(
	definitions []astmodel.TypeDefiner,
) ([]astmodel.TypeDefiner, error) {
	// Build a referrers map for each type.
	referrers := make(map[astmodel.TypeName]TypeNameSet)
	roots := make(TypeNameSet)

	for _, def := range definitions {
		if _, ok := def.(*astmodel.ResourceDefinition); ok {
			roots.Add(def.Name())
		}
		for _, referee := range def.Type().Referees() {
			if referee == nil {
				return nil, fmt.Errorf("nil referee for %s", def.Name())
			}
			refereeVal := *referee
			referrers[refereeVal] = referrers[refereeVal].Add(def.Name())
		}
	}

	// var newDefinitions []astmodel.TypeDefiner
	// for _, def := range definitions {

	// }
	return definitions, nil
}
