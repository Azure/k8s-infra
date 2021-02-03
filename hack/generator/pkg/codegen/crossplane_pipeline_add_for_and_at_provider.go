/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
)

// addCrossplaneForProvider adds a "ForProvider" property as the sole property in every resource spec
// and moves everything that was at the spec level down a level into the ForProvider type
func addCrossplaneForProvider(idFactory astmodel.IdentifierFactory) PipelineStage {

	return MakePipelineStage(
		"addForProviderProperty",
		"Adds a for provider property on every spec",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {

			result := make(astmodel.Types)
			for _, typeDef := range types {
				if _, ok := typeDef.Type().(*astmodel.ResourceType); ok {
					forProviderTypes, err := nestSpecIntoForProvider(
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

// addCrossplaneAtProvider adds an "AtProvider" property as the sole property in every resource status
func addCrossplaneAtProvider(idFactory astmodel.IdentifierFactory) PipelineStage {

	return MakePipelineStage(
		"addAtProviderProperty",
		"Adds an at provider property on every status",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {

			result := make(astmodel.Types)
			for _, typeDef := range types {
				if _, ok := typeDef.Type().(*astmodel.ResourceType); ok {
					atProviderTypes, err := nestStatusIntoAtProvider(
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

// nestSpecIntoForProvider returns the type definitions required to nest the contents of the "Spec" type
// into a property named "ForProvider" whose type is "<name>Parameters"
func nestSpecIntoForProvider(
	idFactory astmodel.IdentifierFactory,
	types astmodel.Types,
	typeDef astmodel.TypeDefinition) ([]astmodel.TypeDefinition, error) {

	resource := typeDef.Type().(*astmodel.ResourceType)
	resourceName := typeDef.Name()

	specName, ok := resource.SpecType().(astmodel.TypeName)
	if !ok {
		return nil, errors.Errorf("Resource %q spec was not of type TypeName, instead: %T", resourceName, resource.SpecType())
	}

	nestedTypeName := resourceName.Name() + "Parameters"
	nestedPropertyName := "ForProvider"
	return nestType(idFactory, types, specName, nestedTypeName, nestedPropertyName)
}

// nestStatusIntoAtProvider returns the type definitions required to nest the contents of the "Status" type
// into a property named "AtProvider" whose type is "<name>Observation"
func nestStatusIntoAtProvider(
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

	nestedTypeName := resourceName.Name() + "Observation"
	nestedPropertyName := "AtProvider"
	return nestType(idFactory, types, statusName, nestedTypeName, nestedPropertyName)
}

// nestType nests the contents of the provided outerType into a property with the given nestedPropertyName whose
// type is the given nestedTypeName. The result is a type that looks something like the following:
//
// type <outerTypeName> struct {
//     <nestedPropertyName> <nestedTypeName> `yaml:"<nestedPropertyName>"`
// }
func nestType(
	idFactory astmodel.IdentifierFactory,
	types astmodel.Types,
	outerTypeName astmodel.TypeName,
	nestedTypeName string,
	nestedPropertyName string) ([]astmodel.TypeDefinition, error) {

	outerType, ok := types[outerTypeName]
	if !ok {
		return nil, errors.Errorf("couldn't find type %q", outerTypeName)
	}

	outerObject, ok := outerType.Type().(*astmodel.ObjectType)
	if !ok {
		return nil, errors.Errorf("type %q was not of type ObjectType, instead %T", outerTypeName, outerType.Type())
	}

	var result []astmodel.TypeDefinition

	// Copy outer type properties onto new "nesting type" with name nestedTypeName
	nestedDef := astmodel.MakeTypeDefinition(
		astmodel.MakeTypeName(outerTypeName.PackageReference, nestedTypeName),
		outerObject)
	result = append(result, nestedDef)

	// Change existing spec to have a single property pointing to the above parameters object
	updatedObject := outerObject.WithoutProperties().WithProperty(
		astmodel.NewPropertyDefinition(
			idFactory.CreatePropertyName(nestedPropertyName, astmodel.Exported),
			idFactory.CreateIdentifier(nestedPropertyName, astmodel.NotExported),
			nestedDef.Name()))
	result = append(result, astmodel.MakeTypeDefinition(outerTypeName, updatedObject))

	return result, nil
}
