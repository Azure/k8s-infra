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

const resourcesPropertyName = astmodel.PropertyName("Resources")

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
			specDef, err := getResourceSpecDefinition(definitions, def.Name(), resourceType)
			if err != nil {
				return nil, err
			}

			specType, err := resourceSpecTypeAsObject(specDef)
			if err != nil {
				return nil, errors.Wrapf(err, "Couldn't extract resource %s spec type as object", def.Name())
			}

			childResourcePropertyTypeDef, err := extractChildResourcePropertyTypeDef(
				definitions,
				def.Name(),
				specDef.Name(),
				specType)
			if err != nil {
				return nil, err
			}
			if childResourcePropertyTypeDef == nil {
				continue // This just means skip
			}

			childResourceTypeNames, err := extractChildResourceTypeNames(*childResourcePropertyTypeDef)
			if err != nil {
				return nil, err
			}

			err = createdUpdatedChildResourceDefinitionsWithOwner(definitions, childResourceTypeNames, def.Name(), updatedDefs)
			if err != nil {
				return nil, err
			}

			// Remove the resources property from the owning resource spec
			specDef = specDef.WithType(specType.WithoutProperty(resourcesPropertyName))

			updatedDefs[specDef.Name()] = specDef
		}
	}

	setResourceGroupOwnerForResourcesWithNoOwner(definitions, updatedDefs)
	return combineNewAndExistingDefs(definitions, updatedDefs), nil
}

func resourceSpecTypeAsObject(resourceSpecDef astmodel.TypeDefinition) (*astmodel.ObjectType, error) {
	// There's an expectation here that the spec is a typename pointing to an object. Even if the resource
	// uses AnyOf/OneOf to model some sort of inheritance at this point that will be rendered
	// as an object (with properties, etc)
	specType, ok := resourceSpecDef.Type().(*astmodel.ObjectType)
	if !ok {
		return nil, errors.Errorf(
			"spec (%s) type is %T, not *astmodel.ObjectType",
			resourceSpecDef.Name(),
			resourceSpecDef.Type())
	}

	return specType, nil
}

func extractChildResourcePropertyTypeDef(
	definitions astmodel.Types,
	resourceName astmodel.TypeName,
	resourceSpecName astmodel.TypeName,
	specType *astmodel.ObjectType) (*astmodel.TypeDefinition, error) {

	// We're looking for a magical "Resources" property - if we don't find
	// one just move on
	resourcesProp := findResourcesProperty(specType)
	if resourcesProp == nil {
		return nil, nil
	}

	// The resources property should be an array
	resourcesPropArray, ok := resourcesProp.PropertyType().(*astmodel.ArrayType)
	if !ok {
		return nil, errors.Errorf(
			"Resource %s has spec %s with Resources property whose type is %T not array",
			resourceName,
			resourceSpecName,
			resourcesProp.PropertyType())
	}

	// We're really interested in the type of this array
	resourcesPropertyTypeName, ok := resourcesPropArray.Element().(astmodel.TypeName)
	if !ok {
		return nil, errors.Errorf(
			"Resource %s has spec %s with Resources property whose type is array but whose inner type is not TypeName, instead it is %T",
			resourceName,
			resourceSpecName,
			resourcesPropArray.Element())
	}

	resourcesDef, ok := definitions[resourcesPropertyTypeName]
	if !ok {
		return nil, errors.Errorf("couldn't find definition Resources property type %s", resourcesPropertyTypeName)
	}

	return &resourcesDef, nil
}

func handleObjectTypeResourcesProperty(
	resourcesPropertyName astmodel.TypeName,
	resourcesPropertyType *astmodel.ObjectType) ([]astmodel.TypeName, error) {
	var results []astmodel.TypeName

	// Each property type is a subresource type
	for _, prop := range resourcesPropertyType.Properties() {
		// TODO: Do we need a recursive function here since this can also be a OneOf?
		optionalType, ok := prop.PropertyType().(*astmodel.OptionalType)
		if !ok {
			return nil, errors.Errorf(
				"OneOf type %s property %s not of type *astmodel.OptionalType",
				resourcesPropertyName.Name(),
				prop.PropertyName())
		}

		propTypeName, ok := optionalType.Element().(astmodel.TypeName)
		if !ok {
			return nil, errors.Errorf(
				"OneOf type %s optional property %s not of type astmodel.TypeName",
				resourcesPropertyName.Name(),
				prop.PropertyName())
		}
		results = append(results, propTypeName)
	}

	return results, nil
}

func extractChildResourceTypeNames(resourcesPropertyTypeDef astmodel.TypeDefinition) ([]astmodel.TypeName, error) {
	// This type should be ResourceType, or ObjectType if modelling a OneOf/AllOf
	_, isResource := resourcesPropertyTypeDef.Type().(*astmodel.ResourceType)
	resourcesPropertyTypeAsObject, isObject := resourcesPropertyTypeDef.Type().(*astmodel.ObjectType)
	if !isResource && !isObject {
		return nil, errors.Errorf(
			"Resources property type %s was not of type *astmodel.ObjectType or *astmodel.ResourceType, instead %T",
			resourcesPropertyTypeDef.Name(),
			resourcesPropertyTypeDef.Type())
	}

	// Determine if this is a OneOf/AllOf
	// TODO: Checking for the presence of the JSON marshal function is a bit of a hack...
	if isObject && resourcesPropertyTypeAsObject.HasFunctionWithName(astmodel.JSONMarshalFunctionName) {
		return handleObjectTypeResourcesProperty(resourcesPropertyTypeDef.Name(), resourcesPropertyTypeAsObject)
	} else {
		return []astmodel.TypeName{resourcesPropertyTypeDef.Name()}, nil
	}
}

func createdUpdatedChildResourceDefinitionsWithOwner(
	definitions astmodel.Types,
	childResourceTypeNames []astmodel.TypeName,
	owningResourceName astmodel.TypeName,
	updatedDefs astmodel.Types) error {

	for _, typeName := range childResourceTypeNames {
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
			// Bug in spec which there is a PR out for: https://github.com/Azure/azure-resource-manager-schemas/pull/1071
			// TODO: remove the below once PR is merged
			typeName.Name() == "ServersAdministrators" ||
			typeName.Name() == "ExtensionsChild" {
			continue
		}

		// Confirm the type really exists
		childResourceDef, ok := definitions[typeName]
		if !ok {
			return errors.Errorf("couldn't find child resource type %s", typeName)
		}

		// Update the definition of the child resource type to point to its owner
		childResource, ok := childResourceDef.Type().(*astmodel.ResourceType)
		if !ok {
			return errors.Errorf("child resource %s not of type *astmodel.ResourceType, instead %T", typeName, childResourceDef.Type())
		}

		childResourceDef = childResourceDef.WithType(childResource.WithOwner(&owningResourceName))

		// There can be overwrites here because updatedDefs contains all resources we're processing and it's possible
		// to have a resource graph where A owns B owns C, but we process B first, resulting in B having the "Resources" property
		// removed and being put into the updatedDefs, and then we process A which will try to add an owner to B which is already
		// in updatedDefs. Because order doesn't matter here we're okay with an overwrite.
		updatedDefs[typeName] = childResourceDef
	}

	return nil
}

func setResourceGroupOwnerForResourcesWithNoOwner(
	definitions astmodel.Types,
	updatedDefs astmodel.Types) {

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
			// This can overwrite because a resource with no owner may have had child resources,
			// and earlier on in this process we removed the resources property from the parent resource,
			// so it may already be in updatedDefs. In this case, that's okay so we allow it to overwrite.
			updatedDefs[def.Name()] = def.WithType(updatedType)
		}
	}
}

func combineNewAndExistingDefs(definitions astmodel.Types, updatedDefs astmodel.Types) astmodel.Types {
	results := make(astmodel.Types)
	for _, updatedDef := range updatedDefs {
		results.Add(updatedDef)
	}

	for name, def := range definitions {
		_, ok := updatedDefs[name]
		if ok {
			continue // Already included above, so skip here to avoid duplicates
		}

		results.Add(def)
	}

	return results
}

func findResourcesProperty(resourceSpec *astmodel.ObjectType) *astmodel.PropertyDefinition {
	for _, prop := range resourceSpec.Properties() {
		if prop.PropertyName() == resourcesPropertyName {
			return prop
		}
	}

	return nil
}
