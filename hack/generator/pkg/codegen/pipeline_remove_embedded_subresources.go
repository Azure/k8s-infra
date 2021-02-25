/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
)

// TODO: Need golden files tests for this

// If something contains the top level "Properties" of a resource that it is also the parent of, just remove that property entirely!
// For example VNET -> Subnet has this with "SubnetPropertiesFormat"

func removeEmbeddedSubResources() PipelineStage {
	return MakePipelineStage(
		"removeEmbeddedSubResources",
		"Removes properties that point to embedded sub-resources",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {

			resourceToSubResourceMap, err := findSubResourcePropertyTypeNames(types)
			if err != nil {
				return nil, errors.Wrap(err, "couldn't find subresource \"Properties\" type names")
			}

			return removeEmbeddedSubResourceProperties(types, resourceToSubResourceMap)
		})
}

func findSubResourcePropertyTypeNames(types astmodel.Types) (map[astmodel.TypeName]astmodel.TypeNameSet, error) {
	resources := types.Where(func(def astmodel.TypeDefinition) bool {
		_, ok := astmodel.AsResourceType(def.Type())
		return ok
	})

	var errs []error
	result := make(map[astmodel.TypeName]astmodel.TypeNameSet)

	// Identify sub-resources and their "properties", associate them with parent resource
	// Look through parent resource for subresource properties
	for _, def := range resources {
		resource, ok := astmodel.AsResourceType(def.Type())
		if !ok {
			// Shouldn't be possible to get here
			panic(fmt.Sprintf("resource was somehow not a resource: %q", def.Name()))
		}

		if resource.Owner() != nil {
			owner := *resource.Owner()
			if result[owner] == nil {
				result[owner] = make(astmodel.TypeNameSet)
			}

			specPropertiesTypeName, statusPropertiesTypeName, err := extractSpecAndStatusPropertiesTypeOrNil(types, resource)
			if err != nil {
				errs = append(errs, errors.Wrapf(err, "couldn't extract spec/status properties from %q", def.Name()))
				continue
			}
			if specPropertiesTypeName != nil {
				result[owner] = result[owner].Add(*specPropertiesTypeName)
			}
			if statusPropertiesTypeName != nil {
				result[owner] = result[owner].Add(*statusPropertiesTypeName)
			}
		}
	}

	if len(errs) > 0 {
		return nil, kerrors.NewAggregate(errs)
	}

	return result, nil
}

func extractSpecAndStatusPropertiesTypeOrNil(types astmodel.Types, resource *astmodel.ResourceType) (*astmodel.TypeName, *astmodel.TypeName, error) {
	specPropertiesTypeName, err := extractSpecPropertiesTypeNameOrNil(types, resource)
	if err != nil {
		return nil, nil, errors.Wrap(err, "couldn't extract spec properties")
	}

	statusPropertiesTypeName, err := extractStatusPropertiesTypeNameOrNil(types, resource)
	if err != nil {
		return nil, nil, errors.Wrap(err, "couldn't extract status properties")
	}

	return specPropertiesTypeName, statusPropertiesTypeName, nil
}

// TODO: Big duplication on _Status detection w/ spec. Combine somehow
func extractStatusPropertiesTypeNameOrNil(types astmodel.Types, resource *astmodel.ResourceType) (*astmodel.TypeName, error) {
	statusName, ok := astmodel.AsTypeName(resource.StatusType())
	if !ok {
		return nil, nil
	}

	specDef, ok := types[statusName]
	if !ok {
		return nil, errors.Errorf("couldn't find %q type", statusName)
	}

	specObj, ok := astmodel.AsObjectType(specDef.Type())
	if !ok {
		return nil, errors.Errorf("%q was not of type object", statusName)
	}

	propertiesProp, ok := specObj.Property("Properties")
	if !ok {
		return nil, nil
	}

	propertiesTypeName, ok := astmodel.AsTypeName(propertiesProp.PropertyType())
	if !ok {
		return nil, nil
	}

	return &propertiesTypeName, nil
}

func extractSpecPropertiesTypeNameOrNil(types astmodel.Types, resource *astmodel.ResourceType) (*astmodel.TypeName, error) {
	specName, ok := astmodel.AsTypeName(resource.SpecType())
	if !ok {
		return nil, errors.Errorf("resource spec was not a TypeName")
	}

	specDef, ok := types[specName]
	if !ok {
		return nil, errors.Errorf("couldn't find %q type", specName)
	}

	specObj, ok := astmodel.AsObjectType(specDef.Type())
	if !ok {
		return nil, errors.Errorf("%q was not of type object", specName)
	}

	propertiesProp, ok := specObj.Property("Properties")
	if !ok {
		return nil, nil
	}

	propertiesTypeName, ok := astmodel.AsTypeName(propertiesProp.PropertyType())
	if !ok {
		return nil, nil
	}

	return &propertiesTypeName, nil
}

// TODO: Method is too long
func removeEmbeddedSubResourceProperties(types astmodel.Types, resourceToSubResourceMap map[astmodel.TypeName]astmodel.TypeNameSet) (astmodel.Types, error) {
	result := make(astmodel.Types)
	typeSuffix := "_SubResourceEmbedded"

	// TODO: Instead of flag we could just use a map here
	embeddedSubResourceType := astmodel.TypeFlag("embeddedSubResource") // TODO: Should we promote this?

	visitor := astmodel.MakeTypeVisitor()
	visitor.VisitObjectType = func(this *astmodel.TypeVisitor, it *astmodel.ObjectType, ctx interface{}) (astmodel.Type, error) {
		typedCtx := ctx.(embeddedSubResourceVisitorContext)

		// TODO: Is this the best way to achieve this protection?
		if typedCtx.depth <= 2 {
			// If we are not at sufficient depth, don't bother checking for subresource references. The resource itself and its spec type
			// will not refer to a subresource. There are some instances of resources (such as Microsoft.Web v20160801 Sites) where the resource
			// and some child resources reuse the same "Properties" type, which could false some of the logic below
			return astmodel.IdentityVisitOfObjectType(this, it, ctx)
		}

		// Before visiting, check if any properties are just referring to one of our sub-resources and remove them
		subResources := resourceToSubResourceMap[typedCtx.resource]
		for _, prop := range it.Properties() {
			propTypeName, ok := astmodel.AsTypeName(prop.PropertyType())
			if !ok {
				continue
			}

			if subResources.Contains(propTypeName) {
				klog.V(0).Infof("Removing resource %q reference to subresource %q on property %q", typedCtx.resource, propTypeName, prop.PropertyName())
				it = removeResourceLikeProperties(it)
			}
		}

		return astmodel.IdentityVisitOfObjectType(this, it, ctx)
	}

	typeWalker := astmodel.NewTypeWalker(types, visitor)
	typeWalker.AfterVisitFunc = func(original astmodel.TypeDefinition, updated astmodel.TypeDefinition, ctx interface{}) (astmodel.TypeDefinition, error) {
		if !original.Name().Equals(updated.Name()) {
			panic(fmt.Sprintf("Unexpeted name mismatch during type walk: %q -> %q", original.Name(), updated.Name()))
		}

		if !original.Type().Equals(updated.Type()) {
			newName := astmodel.MakeTypeName(original.Name().PackageReference, original.Name().Name()+typeSuffix)
			updated = updated.WithName(newName)
			updated = updated.WithType(embeddedSubResourceType.ApplyTo(updated.Type()))

			// TODO: Reduce log verbosity
			klog.V(0).Infof("Updating %q to %q", original.Name(), updated.Name())
		}

		return updated, nil
	}
	typeWalker.MakeContextFunc = func(_ astmodel.TypeName, ctx interface{}) (interface{}, error) {
		typedCtx := ctx.(embeddedSubResourceVisitorContext)
		return typedCtx.WithMoreDepth(), nil
	}

	for _, def := range types {
		if astmodel.IsResourceDefinition(def) {
			ctx := embeddedSubResourceVisitorContext{resource: def.Name(), depth: 0}
			updatedTypes, err := typeWalker.Walk(def, ctx)
			if err != nil {
				return nil, err
			}

			for _, newDef := range updatedTypes {
				err := result.AddWithEqualityCheck(newDef)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return cleanupTypeNames(result, embeddedSubResourceType, typeSuffix)
}

type embeddedSubResourceVisitorContext struct {
	resource astmodel.TypeName
	depth    int
}

func (e embeddedSubResourceVisitorContext) WithMoreDepth() embeddedSubResourceVisitorContext {
	e.depth += 1
	return e
}
