/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
)

// improveResourcePluralization improves pluralization for resources
func improveResourcePluralization() PipelineStage {

	return PipelineStage{
		Name: "Improve resource pluralization",
		Action: func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {

			result := make(astmodel.Types)
			for typeName, typeDef := range types {
				if _, ok := typeDef.Type().(*astmodel.ResourceType); ok {
					newTypeName := typeName.Singular()
					result.Add(typeDef.WithName(newTypeName))
				} else {
					result.Add(typeDef)
				}
			}

			return result, nil
		},
	}
}
