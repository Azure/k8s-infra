/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/klog/v2"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
)

// TODO: Make sure we have a golden file test for if we remove all remaining properties on an object.
// TODO: That object should then just be removed from the type graph and all references to it deleted.

// TODO: Fix comment below
// removeEmbeddedResources creates a pipeline stage that removes embedded resources. For example,
// a VNET is the parent resource of a subnet, but VNET also has a property in the spec pointing to a collection of
// subnets. Since management of those resources is delegated to another CRD (and thus another controller),
// we do not need any details about them the entity. The relationship between parent and child resource
// is managed on the child resource by specifying its owner. In the case of cross-resource references
func removeEmbeddedResources() PipelineStage {
	return MakePipelineStage(
		"removeEmbeddedResources",
		"Removes properties that point to embedded resources",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {
			// TODO: Update below comment
			// 1. Walk the set of resources, keeping in mind what resource is the "root"
			// 2. Find all child properties in the property tree whose types look like a resource (name, properties, type)
			//    note that for some reason some resources that are embedded only have name/properties.
			// 3. If found, duplicate the object and remove the name/properties/type properties. Must be duplicated because
			//    other resources may depend on this thing for _their_ spec/status and if they do we don't want to break them.

			result := make(astmodel.Types)

			resources := types.Where(func(def astmodel.TypeDefinition) bool {
				_, ok := astmodel.AsResourceType(def.Type())
				return ok
			})

			typeSuffix := "_Embedded"
			embeddedResourceFlag := astmodel.TypeFlag("embeddedResource") // TODO: Should we promote this?

			visitor := makeEmbeddedResourceRemovalVisitor()
			typeWalker := astmodel.NewTypeWalker(types, visitor)
			typeWalker.MakeContextFunc = makeEmbeddedResourceRemovalCtx
			// TODO: Below is duplicated
			typeWalker.AfterVisitFunc = func(original astmodel.TypeDefinition, updated astmodel.TypeDefinition, ctx interface{}) (astmodel.TypeDefinition, error) {
				if !original.Name().Equals(updated.Name()) {
					panic(fmt.Sprintf("Unexpeted name mismatch during type walk: %q -> %q", original.Name(), updated.Name()))
				}

				if !original.Type().Equals(updated.Type()) {

					// TODO: So there is an issue, we have two situations:
					// 1. A resource embeds another resource at a low level (5/6 props in). Since we change that embedded types name, the extra "_Embedded" percolates
					//    all the way up to the top resource type as well (and touches every type on the way).
					// 2. The resource itself (or its properties type) is embedded in another resource. So we remove the resource-like properties and make a new type with the
					//    "_Embedded" suffix.
					// The problem is that a type can BOTH look like a resource and also have other resources embedded in it. So while we really want 3 contexts (normal, used in its own
					// resource but with embedded subresources removed, and used as an embedded resource itself.

					// TODO: This is definitely a bit hacky
					originalOT, isOriginalObjectType := astmodel.AsObjectType(original.Type())
					updatedOT, isUpdatedObjectType := astmodel.AsObjectType(updated.Type())

					// Don't change name if we're just stripping resources embedded in one of our properties. The _Embedded suffix is for when
					// this resource is also embedded inside another resource.
					if !isUpdatedObjectType || !isOriginalObjectType || len(originalOT.Properties()) != len(updatedOT.Properties()) {
						newName := astmodel.MakeTypeName(original.Name().PackageReference, original.Name().Name()+typeSuffix)
						updated = updated.WithName(newName)
						updated = updated.WithType(embeddedResourceFlag.ApplyTo(updated.Type()))
					}
				}

				return updated, nil
			}

			for _, resource := range resources {
				updatedTypes, err := typeWalker.Walk(resource, embeddedResourceVisitorContext{name: resource.Name(), depth: 0})
				if err != nil {
					return nil, err
				}

				// It's possible that this type was already visited - as long as the resulting shape is the same we're ok
				for _, t := range updatedTypes {
					err := result.AddWithEqualityCheck(t)
					if err != nil {
						return nil, err
					}
				}
			}

			return cleanupTypeNames(result, embeddedResourceFlag, typeSuffix)
		})
}

type embeddedResourceVisitorContext struct {
	depth int
	name  astmodel.TypeName
}

func (e embeddedResourceVisitorContext) WithName(name astmodel.TypeName) embeddedResourceVisitorContext {
	e.name = name
	e.depth += 1
	return e
}

func makeEmbeddedResourceRemovalVisitor() astmodel.TypeVisitor {
	visitor := astmodel.MakeTypeVisitor()
	visitor.VisitObjectType = func(this *astmodel.TypeVisitor, it *astmodel.ObjectType, ctx interface{}) (astmodel.Type, error) {
		typedCtx := ctx.(embeddedResourceVisitorContext) // TODO: Possibly can drop name from this entirely?

		if typedCtx.depth > 2 {
			it = removeResourceLikeProperties(it)
		}

		return astmodel.IdentityVisitOfObjectType(this, it, ctx)
	}

	return visitor
}

func makeEmbeddedResourceRemovalCtx(name astmodel.TypeName, ctx interface{}) (interface{}, error) {
	typedCtx := ctx.(embeddedResourceVisitorContext)
	return typedCtx.WithName(name), nil
}

// removeResourceLikeProperties examines an astmodel.ObjectType and determines if it looks like an Azure resource.
// An object is "like" a resource if it has "name", "type" and "properties" properties.
func removeResourceLikeProperties(o *astmodel.ObjectType) *astmodel.ObjectType {
	mustRemove := []string{
		"Name",
		"Properties",

		// TODO: I think type actually needs to be required -- other kinds of embeds may not have a type but need to be treated
		// TODO: differently (see VNET -> Subnet)
		//"Type",
	}
	canRemove := []string{
		"Type",
		"Etag",
		"Location",
		"Tags",
	}

	hasRequiredProperties := true
	for _, propName := range mustRemove {
		_, hasProp := o.Property(astmodel.PropertyName(propName))
		hasRequiredProperties = hasRequiredProperties && hasProp
	}

	if !hasRequiredProperties {
		// Doesn't match the shape we're looking for -- no change
		return o
	}
	result := o
	for _, propName := range append(mustRemove, canRemove...) {
		result = result.WithoutProperty(astmodel.PropertyName(propName))
	}
	return result
}

// Only keep the suffix if the shorter (non-suffixed) name is also in use
func cleanupTypeNames(types astmodel.Types, flag astmodel.TypeFlag, suffix string) (astmodel.Types, error) {
	renames := make(map[astmodel.TypeName]astmodel.TypeName)
	result := make(astmodel.Types)

	for _, def := range types {
		if flag.IsOn(def.Type()) {
			nameWithoutSuffix := astmodel.MakeTypeName(def.Name().PackageReference, strings.TrimSuffix(def.Name().Name(), suffix))
			_, ok := types[nameWithoutSuffix]
			if !ok {
				renames[def.Name()] = nameWithoutSuffix

				// TODO: Log level
				klog.V(0).Infof("There are no usages of %q. Removing %q suffix from %q as collapsed name is available.", nameWithoutSuffix, suffix, def.Name())
			}
		}
	}

	renamingVisitor := astmodel.MakeTypeVisitor()
	renamingVisitor.VisitTypeName = func(this *astmodel.TypeVisitor, it astmodel.TypeName, ctx interface{}) (astmodel.Type, error) {
		if newName, ok := renames[it]; ok {
			return astmodel.IdentityVisitOfTypeName(this, newName, ctx)
		}
		return astmodel.IdentityVisitOfTypeName(this, it, ctx)
	}

	for _, def := range types {
		updatedDef, err := renamingVisitor.VisitDefinition(def, nil)
		if err != nil {
			return nil, err
		}
		updatedType, err := flag.RemoveFrom(updatedDef.Type()) // TODO: If we don't remove this, something is causing these types to not be emitted. Unsure what that is... should track it down
		if err != nil {
			return nil, err
		}
		result.Add(updatedDef.WithType(updatedType))
	}

	return result, nil
}
