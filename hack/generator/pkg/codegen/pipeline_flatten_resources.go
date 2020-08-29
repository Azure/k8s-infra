/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
)

// convertAllOfAndOneOfToObjects reduces the AllOfType and OneOfType to ObjectType
func flattenResources() PipelineStage {
	return MakePipelineStage(
		"flatten-resources",
		"Flatten nested resource types",
		func(ctx context.Context, defs astmodel.Types) (astmodel.Types, error) {

			var flattenResource func(astmodel.TypeDefinition) (astmodel.TypeDefinition, error)
			flattenResource = func(def astmodel.TypeDefinition) (astmodel.TypeDefinition, error) {
				if resource, ok := def.Type().(*astmodel.ResourceType); ok {

					changed := false

					// resolve spec
					{
						specType, err := defs.FullyResolve(resource.SpecType())
						if err != nil {
							return astmodel.TypeDefinition{}, err
						}

						if specResource, ok := specType.(*astmodel.ResourceType); ok {
							resource = resource.WithSpec(specResource.SpecType())
							changed = true
						}
					}

					// resolve status
					{
						statusType, err := defs.FullyResolve(resource.StatusType())
						if err != nil {
							return astmodel.TypeDefinition{}, err
						}

						if statusResource, ok := statusType.(*astmodel.ResourceType); ok {
							resource = resource.WithStatus(statusResource.StatusType())
							changed = true
						}
					}

					if changed {
						def = def.WithType(resource)
						// do it again, can be multiply nested
						return flattenResource(def)
					}
				}

				return def, nil
			}

			result := make(astmodel.Types)

			for _, def := range defs {
				newDef, err := flattenResource(def)
				if err != nil {
					return nil, err
				}

				result.Add(newDef)
			}

			return result, nil
		})
}
