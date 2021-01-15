/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
)

// addCrossplaneOwnerProperties adds the 3-tuple of (xName, xNameRef, xNameSelector) for each owning resource
func addCrossplaneOwnerProperties(idFactory astmodel.IdentifierFactory) PipelineStage {

	return MakePipelineStage(
		"addCrossplaneOwnerProperties",
		"Adds the 3-tuple of (xName, xNameRef, xNameSelector) for each owning resource",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {
			referenceTypeName := astmodel.MakeTypeName(
				RuntimeV1Alpha1PackageReference,
				idFactory.CreateIdentifier("Reference", astmodel.Exported))
			selectorTypeName := astmodel.MakeTypeName(
				RuntimeV1Alpha1PackageReference,
				idFactory.CreateIdentifier("Selector", astmodel.Exported))

			result := make(astmodel.Types)
			for _, typeDef := range types {

				if resource, ok := typeDef.Type().(*astmodel.ResourceType); ok {

					owners, err := lookupOwners(types, typeDef)
					if err != nil {
						return nil, errors.Wrapf(err, "failed to look up owners for %s", typeDef.Name())
					}

					// The right-most owner is this type, so remove it
					owners = owners[0:len(owners)-1]

					// This type has no owners so no modification needed
					if len(owners) == 0 {
						continue
					}

					// TODO: This function should be shared in some common place?
					specDef, err := getResourceSpecDefinition(types, resource)
					if err != nil {
						return nil, errors.Wrapf(err, "getting resource spec definition")
					}

					for _, owner := range owners {
						nameSubset := fmt.Sprintf("%sName", owner.Name())
						name := idFactory.CreatePropertyName(nameSubset, astmodel.Exported)
						nameRef := name + "Ref"
						nameSelector := name + "Selector"

						updatedDef, err := specDef.ApplyObjectTransformation(func(o *astmodel.ObjectType) (astmodel.Type, error) {
							nameProperty := astmodel.NewPropertyDefinition(
								name,
								idFactory.CreateIdentifier(string(name), astmodel.NotExported),
								astmodel.StringType)
							nameRefProperty := astmodel.NewPropertyDefinition(
								nameRef,
								idFactory.CreateIdentifier(string(nameRef), astmodel.NotExported),
								referenceTypeName).MakeOptional()
							nameSelectorProperty := astmodel.NewPropertyDefinition(
								nameSelector,
								idFactory.CreateIdentifier(string(nameSelector), astmodel.NotExported),
								selectorTypeName).MakeOptional()

							result := o.WithProperty(nameProperty).WithProperty(nameRefProperty).WithProperty(nameSelectorProperty)
							return result, nil
						})
						if err != nil {
							return nil, errors.Wrapf(err, "adding ownership properties to spec")
						}
						specDef = updatedDef
					}
					result.Add(typeDef)
					result.Add(specDef)
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

func lookupOwners(defs astmodel.Types, resourceDef astmodel.TypeDefinition) ([]astmodel.TypeName, error) {
	resourceType, ok := resourceDef.Type().(*astmodel.ResourceType)
	if !ok {
		return nil, errors.Errorf("type %s is not a resource", resourceDef.Name())
	}

	if resourceType.Owner() == nil {
		return []astmodel.TypeName {resourceDef.Name()}, nil
	}

	if resourceType.Owner().Name() == "ResourceGroup" {
		return []astmodel.TypeName {*resourceType.Owner(), resourceDef.Name()}, nil
	}

	owner := *resourceType.Owner()
	ownerDef, ok := defs[owner]
	if !ok {
		return nil, errors.Errorf("couldn't find definition for owner %s", owner)
	}

	result, err := lookupOwners(defs, ownerDef)
	if err != nil {
		return nil, err
	}

	return append(result, resourceDef.Name()), nil
}
