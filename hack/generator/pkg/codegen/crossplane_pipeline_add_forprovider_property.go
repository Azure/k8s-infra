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

// addCrossplaneForProvider adds a "ForProvider" property as the sole property in every resource spec
func addCrossplaneForProvider(idFactory astmodel.IdentifierFactory) PipelineStage {

	return MakePipelineStage(
		"addForProviderProperty",
		"Adds a for provider property on every spec",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {

			result := make(astmodel.Types)
			for _, typeDef := range types {
				if _, ok := typeDef.Type().(*astmodel.ResourceType); ok {
					forProviderTypes, err := createForProviderTypeDefs(
						idFactory, types, typeDef)
					if err != nil {
						return nil, errors.Wrapf(err, "creating ForProvider types")
					}

					result.AddAll(forProviderTypes)
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

// TODO: Some duplicate with above, refactor
// addCrossplaneAtProvider adds an "AtProvider" property as the sole property in every resource status
func addCrossplaneAtProvider(idFactory astmodel.IdentifierFactory) PipelineStage {

	return MakePipelineStage(
		"addAtProviderProperty",
		"Adds an at provider property on every status",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {

			result := make(astmodel.Types)
			for _, typeDef := range types {
				if _, ok := typeDef.Type().(*astmodel.ResourceType); ok {
					atProviderTypes, err := createAtProviderTypeDefs(
						idFactory, types, typeDef)
					if err != nil {
						return nil, errors.Wrapf(err, "creating AtProvider types")
					}
					result.AddAll(atProviderTypes)
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

func createForProviderTypeDefs(
	idFactory astmodel.IdentifierFactory,
	types astmodel.Types,
	typeDef astmodel.TypeDefinition) ([]astmodel.TypeDefinition, error) {

	resource := typeDef.Type().(*astmodel.ResourceType)
	resourceName := typeDef.Name()

	specName, ok := resource.SpecType().(astmodel.TypeName)
	if !ok {
		return nil, errors.Errorf("Resource %q spec was not of type TypeName, instead: %T", resourceName, resource.SpecType())
	}

	spec, ok := types[specName]
	if !ok {
		return nil, errors.Errorf("Couldn't find resource spec %q", specName)
	}

	specObject, ok := spec.Type().(*astmodel.ObjectType)
	if !ok {
		return nil, errors.Errorf("Spec %q was not of type ObjectType, instead %T", specName, spec.Type())
	}

	var result []astmodel.TypeDefinition

	// Copy spec into a new Parameters object and track that
	specNamePrefix := strings.Split(specName.Name(), "_")[0]
	parametersName := astmodel.MakeTypeName(specName.PackageReference, specNamePrefix + "Parameters")
	parametersDef := astmodel.MakeTypeDefinition(parametersName, specObject)
	result = append(result, parametersDef)

	// Change existing spec to have a single property pointing to the above parameters object
	newSpec := specObject.WithoutProperties().WithProperty(
		astmodel.NewPropertyDefinition(
			idFactory.CreatePropertyName("ForProvider", astmodel.Exported),
			idFactory.CreateIdentifier("ForProvider", astmodel.NotExported),
			parametersName))
	result = append(result, astmodel.MakeTypeDefinition(specName, newSpec))
	result = append(result, typeDef)

	return result, nil
}

// TODO: This code could be shared with above, and just pass the Spec/Status and the name transform
func createAtProviderTypeDefs(
	idFactory astmodel.IdentifierFactory,
	types astmodel.Types,
	typeDef astmodel.TypeDefinition) ([]astmodel.TypeDefinition, error) {

	resource := typeDef.Type().(*astmodel.ResourceType)
	resourceName := typeDef.Name()

	statusType := astmodel.IgnoringErrors(resource.StatusType())
	if statusType == nil {
		return nil, nil // TODO: Some types don't have status yet
	}

	statusName, ok := resource.StatusType().(astmodel.TypeName)
	if !ok {
		return nil, errors.Errorf("Resource %q status was not of type TypeName, instead: %T", resourceName, resource.StatusType())
	}

	status, ok := types[statusName]
	if !ok {
		return nil, errors.Errorf("Couldn't find resource status %q", statusName)
	}

	statusObject, ok := status.Type().(*astmodel.ObjectType)
	if !ok {
		return nil, errors.Errorf("Status %q was not of type ObjectType, instead %T", statusName, status.Type())
	}

	var result []astmodel.TypeDefinition

	// Copy spec into a new Parameters object and track that
	statusNamePrefix := resourceName.Name()
	observationName := astmodel.MakeTypeName(statusName.PackageReference, statusNamePrefix + "Observation")
	observationDef := astmodel.MakeTypeDefinition(observationName, statusObject)
	result = append(result, observationDef)

	// Change existing spec to have a single property pointing to the above parameters object
	newSpec := statusObject.WithoutProperties().WithProperty(
		astmodel.NewPropertyDefinition(
			idFactory.CreatePropertyName("AtProvider", astmodel.Exported),
			idFactory.CreateIdentifier("AtProvider", astmodel.NotExported),
			observationName))
	result = append(result, astmodel.MakeTypeDefinition(statusName, newSpec))
	result = append(result, typeDef)

	return result, nil
}
