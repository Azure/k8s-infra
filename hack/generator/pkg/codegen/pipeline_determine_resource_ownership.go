/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/pkg/errors"
	"strings"
)

const resourcesPropertyName astmodel.PropertyName = astmodel.PropertyName("Resources")

func determineResourceOwnership() PipelineStage {
	return PipelineStage{
		id: "determineResourceOwnership",
		description: "Determine ARM resource relationships",
		Action: func(ctx context.Context, definitions astmodel.Types) (astmodel.Types, error) {
			return determineOwnership(definitions)
		},
	}
}

func determineOwnership(definitions astmodel.Types) (astmodel.Types, error) {

	updatedDefs := make(astmodel.Types)

	for _, def := range definitions {
		if resourceType, ok := def.Type().(*astmodel.ResourceType); ok {
			// There's an expectation here that the spec is a typename pointing to an object. Even if the resource
			// uses AnyOf/OneOf to model some sort of inheritance at this point that will be rendered
			// as an object (with properties, etc)
			specName, ok := resourceType.SpecType().(astmodel.TypeName)
			if !ok {
				return nil, errors.Errorf(
					"Resource %s has spec whose type is %T, not TypeName",
					def.Name(),
					resourceType.SpecType())
			}

			specDef, ok := definitions[specName]
			if !ok {
				return nil, errors.Errorf("couldn't find spec for resource %s", def.Name())
			}

			specType, ok := specDef.Type().(*astmodel.ObjectType)
			if !ok {
				return nil, errors.Errorf(
					"Resource %s has spec (%s) whose type is %T, not *astmodel.ObjectType",
					def.Name(),
					specDef.Name(),
					specDef.Type())
			}

			// We're looking for a magical "Resources" property - if we don't find
			// one we just move on
			resourcesProp := findResourcesProperty(specType)
			if resourcesProp == nil {
				continue
			}

			// The resources property should be an array
			resourcesPropArray, ok := resourcesProp.PropertyType().(*astmodel.ArrayType)
			if !ok {
				return nil, errors.Errorf(
					"Resource %s has spec %s with Resources property whose type is %T not array",
					def.Name(),
					specDef.Name(),
					resourcesProp.PropertyType())
			}

			// We're really interested in the type of this array
			resourcesPropertyTypeName, ok := resourcesPropArray.Element().(astmodel.TypeName)
			if !ok {
				return nil, errors.Errorf(
					"Resource %s has spec %s with Resources property whose type is array but whose inner type is not TypeName, instead it is %T",
					def.Name(),
					specDef.Name(),
					resourcesPropArray.Element())
			}

			resourcesDef, ok := definitions[resourcesPropertyTypeName]
			if !ok {
				return nil, errors.Errorf("couldn't find definition Resources property type %s", resourcesPropertyTypeName)
			}

			// This type should may be ResourceType, or ObjectType if modelling a OneOf/AllOf
			_, isResource := resourcesDef.Type().(*astmodel.ResourceType)
			resourcesPropertyTypeAsObject, isObject := resourcesDef.Type().(*astmodel.ObjectType)
			if !isResource && !isObject {
				return nil, errors.Errorf(
					"Resources property type %s was not of type *astmodel.ObjectType or *astmodel.ResourceType, instead %T",
					resourcesPropertyTypeName,
					resourcesDef.Type())
			}

			var candidateChildResourceTypeNames []astmodel.TypeName

			// Determine if this is a OneOf/AllOf
			// TODO: This is a bit of a hack
			if isObject && resourcesPropertyTypeAsObject.HasFunctionWithName(astmodel.JSONMarshalFunctionName) {
				// Each property type is a subresource type
				for _, prop := range resourcesPropertyTypeAsObject.Properties() {
					// TODO: Do we need a recursive function here since this can also be a OneOf?
					optionalType, ok := prop.PropertyType().(*astmodel.OptionalType)
					if !ok {
						return nil, errors.Errorf(
							"OneOf type %s property %s not of type *astmodel.OptionalType",
							resourcesPropertyTypeName,
							prop.PropertyName())
					}

					propTypeName, ok := optionalType.Element().(astmodel.TypeName)
					if !ok {
						return nil, errors.Errorf(
							"OneOf type %s optional property %s not of type astmodel.TypeName",
							resourcesPropertyTypeName,
							prop.PropertyName())
					}
					candidateChildResourceTypeNames = append(candidateChildResourceTypeNames, propTypeName)
				}
			} else {
				candidateChildResourceTypeNames = append(candidateChildResourceTypeNames, resourcesPropertyTypeName)
			}

			for _, typeName := range candidateChildResourceTypeNames {
				// If the typename ends in ChildResource, remove that
				if strings.HasSuffix(typeName.Name(), "ChildResource") {
					typeName = astmodel.MakeTypeName(typeName.PackageReference, strings.TrimSuffix(typeName.Name(), "ChildResource"))
				}

				// TODO: These are types that cause us trouble... a lot of them use allof inheritance.
				// TODO: I think for these we will need to walk the graph of types and do a structural
				// TODO: equality check to find the name of the actual resource, but we can't do that check
				// TODO: now because these types allOf inherit from resourceBase and the actual resources
				// TODO: being referenced do not. See:
				if typeName.Name() == "VirtualMachinesSpec_Resources" || // Uses allof inheritance
					typeName.Name() == "AccountSpec_Resources" || // Uses allof inheritance
					typeName.Name() == "SitesSpec_Resources" || // Uses allof inheritance
					typeName.Name() == "NamespacesSpec_Resources" || // Uses allof inheritance
					typeName.Name() == "VaultsSpec_Resources" || // Uses allof inheritance
					typeName.Name() == "ExtensionsChild" {
					continue
				}

				// Confirm the type really exists
				childResourceDef, ok := definitions[typeName]
				if !ok {
					return nil, errors.Errorf("couldn't find child resource type %s", typeName)
				}

				// Update the definition of the child resource type to point to its owner
				childResource, ok := childResourceDef.Type().(*astmodel.ResourceType)
				if !ok {
					return nil, errors.Errorf("child resource %s not of type *astmodel.ResourceType, instead %T", typeName, childResourceDef.Type())
				}

				ownerName := def.Name()
				childResourceDef = childResourceDef.WithType(childResource.WithOwner(&ownerName))
				updatedDefs[typeName] = childResourceDef
			}

			// Remove the resources property from the owning resource spec
			specDef = specDef.WithType(specType.WithoutProperty(resourcesPropertyName))

			updatedDefs[specDef.Name()] = specDef
		}
	}

	// TODO: Refactor
	// Go over all of the resource types and flag any that don't have an owner as having resource group as their owner
	for _, def := range definitions {
		// Check if we've already modified this type
		if updatedDef, ok := updatedDefs[def.Name()]; ok {
			def = updatedDef
		}

		resourceType, ok := def.Type().(*astmodel.ResourceType)
		if !ok {
			continue
		}

		if resourceType.Owner() == nil {
			ownerTypeName := astmodel.MakeTypeName(
				// Note that the version doesn't really matter here -- it's removed later. We just need to refer to the logical
				// resource group really
				astmodel.MakeLocalPackageReference("microsoft.resources", "v20191001"),
				"ResourceGroup")
			updatedType := resourceType.WithOwner(&ownerTypeName) // TODO: Note that right now... this type doesn't actually exist...
			// This can overwrite but that's okay
			updatedDefs[def.Name()] = def.WithType(updatedType)
		}
	}

	// TODO: Below code is duplicated and could be refactored
	results := make(astmodel.Types)
	for name, updatedSpec := range updatedDefs {
		results[name] = updatedSpec
	}

	for name, def := range definitions {
		_, ok := updatedDefs[name]
		if ok {
			continue // Already included above, so skip here to avoid duplicates
		}

		results[name] = def
	}

	return results, nil
}

func findResourcesProperty(resourceSpec *astmodel.ObjectType) *astmodel.PropertyDefinition {
	for _, prop := range resourceSpec.Properties() {
		if prop.PropertyName() == resourcesPropertyName {
			return prop
		}
	}

	return nil
}
