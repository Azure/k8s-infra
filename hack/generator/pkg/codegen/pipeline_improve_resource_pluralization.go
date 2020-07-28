/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"sort"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
)

// improveResourcePluralization improves pluralization for resources
func improveResourcePluralization() PipelineStage {

	return PipelineStage{
		Name: "Improve resource pluralization",
		Action: func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {

			result := make(astmodel.Types)
			var resources []astmodel.TypeDefinition
			for _, typeDef := range types {
				if _, ok := typeDef.Type().(*astmodel.ResourceType); ok {
					resources = append(resources, typeDef)
				} else {
					result.Add(typeDef)
				}
			}

			// now sort resources so that plurals will come after any singulars that
			// already exist, so we don't conflict
			sort.Slice(resources, func(i, j int) bool {
				return resources[i].Name().Name() < resources[j].Name().Name()
			})

			for _, resource := range resources {
				newTypeName := resource.Name().Singular()
				if _, ok := result[newTypeName]; !ok {
					result.Add(resource.WithName(newTypeName))
				} else {
					// non-plural form already exists, don't depluralize
					result.Add(resource)
				}
			}

			return result, nil
		},
	}
}
