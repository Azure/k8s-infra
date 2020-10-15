/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
)

// addForProviderProperty adds a "ForProvider" property as the sole property in every resource spec
func addForProviderProperty(idFactory astmodel.IdentifierFactory) PipelineStage {

	return MakePipelineStage(
		"addForProviderProperty",
		"Adds a for provider property on every spec",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {

			result := make(astmodel.Types)
			for _, typeDef := range types {
				if resource, ok := typeDef.Type().(*astmodel.ResourceType); ok {
					specName, ok := resource.SpecType().(astmodel.TypeName)
					if !ok {
						return nil, errors.Errorf("Resource %q spec was not of type TypeName, instead: %T", typeDef.Name(), typeDef.Type())
					}

					spec, ok := types[specName]
					if !ok {
						return nil, errors.Errorf("Couldn't find resource spec %q", specName)
					}

					specObject, ok := spec.Type().(*astmodel.ObjectType)
					if !ok {
						return nil, errors.Errorf("Spec %q was not of type ObjectType, instead %T", specName, spec.Type())
					}

					// Copy spec into a new Parameters object and track that
					specNamePrefix := strings.Split(specName.Name(), "_")[0]
					parametersName := astmodel.MakeTypeName(specName.PackageReference, specNamePrefix + "Parameters")
					parametersDef := astmodel.MakeTypeDefinition(parametersName, specObject)
					result.Add(parametersDef)

					// Change existing spec to have a single property pointing to the above parameters object
					newSpec := specObject.WithoutProperties().WithProperty(
						astmodel.NewPropertyDefinition(
							idFactory.CreatePropertyName("ForProvider", astmodel.Exported),
							idFactory.CreateIdentifier("ForProvider", astmodel.NotExported),
							parametersName))
					result.Add(astmodel.MakeTypeDefinition(specName, newSpec))

					// newDef := typeDef.WithType(resource.WithSpec(newSpec))
					result.Add(typeDef)
				}
			}

			for _, typeDef := range types {
				if !result.Contains(typeDef.Name()) {
					result.Add(typeDef)
				}
			}

			return result, nil
		})
}
