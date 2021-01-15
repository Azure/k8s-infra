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

var RuntimeV1Alpha1PackageReference = astmodel.MakeExternalPackageReference("github.com/crossplane/crossplane-runtime/apis/core/v1alpha1")

// addCrossplaneEmbeddedResourceSpec puts an embedded runtimev1alpha1.ResourceSpec on every spec type
func addCrossplaneEmbeddedResourceSpec(idFactory astmodel.IdentifierFactory) PipelineStage {

	return MakePipelineStage(
		"addCrossplaneEmbeddedResourceSpec",
		"Puts an embedded runtimev1alpha1.ResourceSpec on every spec type",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {
			specTypeName := astmodel.MakeTypeName(
				RuntimeV1Alpha1PackageReference,
				idFactory.CreateIdentifier("ResourceSpec", astmodel.Exported))
			embeddedSpec := astmodel.NewPropertyDefinition("", ",inline", specTypeName)

			result := make(astmodel.Types)
			for _, typeDef := range types {

				if resource, ok := typeDef.Type().(*astmodel.ResourceType); ok {

					// TODO: This function should be shared in some common place?
					specDef, err := getResourceSpecDefinition(types, resource)
					if err != nil {
						return nil, errors.Wrapf(err, "getting resource spec definition")
					}

					// The assumption here is that specs are all Objects
					updatedDef, err := specDef.ApplyObjectTransformation(func(o *astmodel.ObjectType) (astmodel.Type, error) {
						return o.WithEmbeddedProperty(embeddedSpec)
					})
					if err != nil {
						return nil, errors.Wrapf(err, "adding embedded crossplane spec")
					}

					result.Add(typeDef)
					result.Add(updatedDef)
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

// TODO: Some duplication from above
// addCrossplaneEmbeddedResourceStatus puts an embedded runtimev1alpha1.ResourceStatus on every spec type
func addCrossplaneEmbeddedResourceStatus(idFactory astmodel.IdentifierFactory) PipelineStage {

	return MakePipelineStage(
		"addCrossplaneEmbeddedResourceStatus",
		"Puts an embedded runtimev1alpha1.ResourceStatus on every status type",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {
			specTypeName := astmodel.MakeTypeName(
				RuntimeV1Alpha1PackageReference,
				idFactory.CreateIdentifier("ResourceStatus", astmodel.Exported))
			embeddedSpec := astmodel.NewPropertyDefinition("", ",inline", specTypeName)

			result := make(astmodel.Types)
			for _, typeDef := range types {

				if resource, ok := typeDef.Type().(*astmodel.ResourceType); ok {

					if resource.StatusType() == nil {
						continue
					}

					// TODO: This function should be shared in some common place?
					statusDef, err := getResourceStatusDefinition(types, resource)
					if err != nil {
						return nil, errors.Wrapf(err, "getting resource status definition")
					}

					// The assumption here is that specs are all Objects
					updatedDef, err := statusDef.ApplyObjectTransformation(func(o *astmodel.ObjectType) (astmodel.Type, error) {
						return o.WithEmbeddedProperty(embeddedSpec)
					})
					if err != nil {
						return nil, errors.Wrapf(err, "adding embedded crossplane status")
					}

					result.Add(typeDef)
					result.Add(updatedDef)
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

func getResourceStatusDefinition(
	definitions astmodel.Types,
	resourceType *astmodel.ResourceType) (astmodel.TypeDefinition, error) {

	statusName, ok := resourceType.StatusType().(astmodel.TypeName)
	if !ok {
		return astmodel.TypeDefinition{}, errors.Errorf("status was not of type TypeName, instead: %T", resourceType.SpecType())
	}

	resourceStatusDef, ok := definitions[statusName]
	if !ok {
		return astmodel.TypeDefinition{}, errors.Errorf("couldn't find status %v", statusName)
	}

	// preserve outer spec name
	return resourceStatusDef.WithName(statusName), nil
}