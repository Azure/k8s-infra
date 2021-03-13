/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
)

// If something contains the top level "Properties" of a resource that it is also the parent of, just remove that property entirely!
// For example VNET -> Subnet has this with "SubnetPropertiesFormat"

func removeEmbeddedResources() PipelineStage {
	return MakePipelineStage(
		"removeEmbeddedResources",
		"Removes properties that point to embedded resources",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {

			resourceToSubResourceMap, err := findSubResourcePropertyTypeNames(types)
			if err != nil {
				return nil, errors.Wrap(err, "couldn't find subresource \"Properties\" type names")
			}

			resourcePropertiesTypes, err := findAllResourcePropertyTypes(types)
			if err != nil {
				return nil, errors.Wrap(err, "couldn't find resource \"Properties\" type names")
			}

			remover := makeEmbeddedResourceRemover(types, resourceToSubResourceMap, resourcePropertiesTypes)
			return remover.removeEmbeddedResourceProperties()
		})
}

// result is a map of resource name to subresource property type names
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

func findAllResourcePropertyTypes(types astmodel.Types) (astmodel.TypeNameSet, error) {
	resources := types.Where(func(def astmodel.TypeDefinition) bool {
		_, ok := astmodel.AsResourceType(def.Type())
		return ok
	})

	var errs []error
	result := make(astmodel.TypeNameSet)

	// Identify sub-resources and their "properties", associate them with parent resource
	// Look through parent resource for subresource properties
	for _, def := range resources {
		resource, ok := astmodel.AsResourceType(def.Type())
		if !ok {
			// Shouldn't be possible to get here
			panic(fmt.Sprintf("resource was somehow not a resource: %q", def.Name()))
		}

		specPropertiesTypeName, statusPropertiesTypeName, err := extractSpecAndStatusPropertiesTypeOrNil(types, resource)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "couldn't extract spec/status properties from %q", def.Name()))
			continue
		}
		if specPropertiesTypeName != nil {
			result = result.Add(*specPropertiesTypeName)
		}
		if statusPropertiesTypeName != nil {
			result = result.Add(*statusPropertiesTypeName)
		}
	}

	if len(errs) > 0 {
		return nil, kerrors.NewAggregate(errs)
	}

	return result, nil
}

func findAllResourceStatusTypes(types astmodel.Types) astmodel.TypeNameSet {
	resources := types.Where(func(def astmodel.TypeDefinition) bool {
		_, ok := astmodel.AsResourceType(def.Type())
		return ok
	})

	result := make(astmodel.TypeNameSet)

	// Identify sub-resources and their "properties", associate them with parent resource
	// Look through parent resource for subresource properties
	for _, def := range resources {
		resource, ok := astmodel.AsResourceType(def.Type())
		if !ok {
			// Shouldn't be possible to get here
			panic(fmt.Sprintf("resource was somehow not a resource: %q", def.Name()))
		}

		statusName, ok := astmodel.AsTypeName(resource.StatusType())
		if !ok {
			continue
		}

		result = result.Add(statusName)
	}

	return result
}

func followTypeNameExtractPropertiesType(types astmodel.Types, typeName astmodel.TypeName) (*astmodel.TypeName, error) {
	def, ok := types[typeName]
	if !ok {
		return nil, errors.Errorf("couldn't find %q type", typeName)
	}

	ot, ok := astmodel.AsObjectType(def.Type())
	if !ok {
		return nil, errors.Errorf("%q was not of type object", typeName)
	}

	propertiesProp, ok := ot.Property("Properties")
	if !ok {
		return nil, nil
	}

	propertiesTypeName, ok := astmodel.AsTypeName(propertiesProp.PropertyType())
	if !ok {
		return nil, nil
	}

	return &propertiesTypeName, nil
}

func extractStatusPropertiesTypeNameOrNil(types astmodel.Types, resource *astmodel.ResourceType) (*astmodel.TypeName, error) {
	statusName, ok := astmodel.AsTypeName(resource.StatusType())
	if !ok {
		return nil, nil
	}

	return followTypeNameExtractPropertiesType(types, statusName)
}

func extractSpecPropertiesTypeNameOrNil(types astmodel.Types, resource *astmodel.ResourceType) (*astmodel.TypeName, error) {
	specName, ok := astmodel.AsTypeName(resource.SpecType())
	if !ok {
		return nil, errors.Errorf("resource spec was not a TypeName")
	}
	return followTypeNameExtractPropertiesType(types, specName)
}

type resourceRemovalVisitorContext struct {
	resource      astmodel.TypeName
	depth         int
	modifiedTypes astmodel.Types
}

func (e resourceRemovalVisitorContext) WithMoreDepth() resourceRemovalVisitorContext {
	e.depth += 1
	return e
}

type embeddedResourceRemover struct {
	types                    astmodel.Types
	resourceToSubresourceMap map[astmodel.TypeName]astmodel.TypeNameSet
	resourcePropertiesTypes  astmodel.TypeNameSet
	resourceStatusTypes      astmodel.TypeNameSet
	typeSuffix               string
	typeFlag                 astmodel.TypeFlag
}

func makeEmbeddedResourceRemover(
	types astmodel.Types,
	resourceToSubresourceMap map[astmodel.TypeName]astmodel.TypeNameSet,
	resourcePropertiesTypes astmodel.TypeNameSet) embeddedResourceRemover {

	// TODO: Weird that we do this here but build the other collections in this caller
	resourceStatusTypes := findAllResourceStatusTypes(types)

	remover := embeddedResourceRemover{
		types:                    types,
		resourceToSubresourceMap: resourceToSubresourceMap,
		resourcePropertiesTypes:  resourcePropertiesTypes,
		resourceStatusTypes:      resourceStatusTypes,
		typeSuffix:               "SubResourceEmbedded",
		typeFlag:                 astmodel.TypeFlag("embeddedSubResource"), // TODO: Instead of flag we could just use a map here if we wanted
	}

	return remover
}

func (e embeddedResourceRemover) MakeEmbeddedResourceRemovalTypeVisitor() astmodel.TypeVisitor {
	visitor := astmodel.MakeTypeVisitor()
	visitor.VisitObjectType = func(this *astmodel.TypeVisitor, it *astmodel.ObjectType, ctx interface{}) (astmodel.Type, error) {
		typedCtx := ctx.(resourceRemovalVisitorContext)

		// TODO: Is this the best way to achieve this protection?
		if typedCtx.depth <= 2 {
			// If we are not at sufficient depth, don't bother checking for subresource references. The resource itself and its spec type
			// will not refer to a subresource. There are some instances of resources (such as Microsoft.Web v20160801 Sites) where the resource
			// and some child resources reuse the same "Properties" type, which could false some of the logic below
			return astmodel.IdentityVisitOfObjectType(this, it, ctx)
		}

		// TODO: This is a hack -- although I guess we don't need it now
		//_, hasLoadBalancerBackendAddressPools := it.Property("LoadBalancerBackendAddressPools")
		//_, hasLoadBalancerInboundNatRules := it.Property("LoadBalancerInboundNatRules")
		//if typedCtx.resource.Name() == "LoadBalancer" && hasLoadBalancerBackendAddressPools && hasLoadBalancerInboundNatRules {
		//	it = it.WithoutProperty("LoadBalancerBackendAddressPools").WithoutProperty("LoadBalancerInboundNatRules")
		//}

		// Before visiting, check if any properties are just referring to one of our sub-resources and remove them
		//subResources := resourceToSubResourceMap[typedCtx.resource]
		for _, prop := range it.Properties() {
			propTypeName, ok := astmodel.AsTypeName(prop.PropertyType())
			if !ok {
				continue
			}

			// TODO: We may need to re-enable the below at some point when the behavior we have is different between subresources (total removal)
			// TODO: and cross resource references. For now though the below statement is a superset of this.
			//if subResources.Contains(propTypeName) {
			//	klog.V(5).Infof("Removing resource %q reference to subresource %q on property %q", typedCtx.resource, propTypeName, prop.PropertyName())
			//	it = removeResourceLikeProperties(it)
			//}

			if e.resourcePropertiesTypes.Contains(propTypeName) {
				klog.V(5).Infof("Removing reference to resource %q on property %q", propTypeName, prop.PropertyName())
				it = removeResourceLikeProperties(it)
			}
		}

		return astmodel.IdentityVisitOfObjectType(this, it, ctx)
	}

	return visitor
}

func (e embeddedResourceRemover) NewResourceRemovalTypeWalker(visitor astmodel.TypeVisitor) *astmodel.TypeWalker {
	typeWalker := astmodel.NewTypeWalker(e.types, visitor)
	typeWalker.AfterVisitFunc = func(original astmodel.TypeDefinition, updated astmodel.TypeDefinition, ctx interface{}) (astmodel.TypeDefinition, error) {
		typedCtx := ctx.(resourceRemovalVisitorContext)

		if !original.Name().Equals(updated.Name()) {
			panic(fmt.Sprintf("Unexpeted name mismatch during type walk: %q -> %q", original.Name(), updated.Name()))
		}

		if !original.Type().Equals(updated.Type()) {
			flaggedType := e.typeFlag.ApplyTo(updated.Type())
			var newName astmodel.TypeName
			exists := false
			for count := 0; ; count++ {
				newName = makeEmbeddedResourceTypeName(embeddedResourceTypeName{original: original.Name(), context: typedCtx.resource.Name(), suffix: e.typeSuffix, count: count})
				existing, ok := typedCtx.modifiedTypes[newName]
				if !ok {
					break
				}
				if existing.Type().Equals(flaggedType) {
					exists = true
					// Shape matches what we have already, can proceed
					break
				}
			}
			updated = updated.WithName(newName)
			updated = updated.WithType(flaggedType)
			if !exists {
				typedCtx.modifiedTypes.Add(updated)
			}

			klog.V(5).Infof("Updating %q to %q", original.Name(), updated.Name())
		}

		return updated, nil
	}
	typeWalker.MakeContextFunc = func(_ astmodel.TypeName, ctx interface{}) (interface{}, error) {
		typedCtx := ctx.(resourceRemovalVisitorContext)
		return typedCtx.WithMoreDepth(), nil
	}
	typeWalker.WalkCycle = func(def astmodel.TypeDefinition, ctx interface{}) (astmodel.TypeName, error) {
		// If we're about to walk a cycle that is to a known resource type, just skip it entirely
		if e.resourcePropertiesTypes.Contains(def.Name()) || e.resourceStatusTypes.Contains(def.Name()) {
			return astmodel.TypeWalkerRemoveType, nil
		}

		if isTypeResourceLookalike(def.Type()) {
			klog.V(5).Infof("Type %q is a resource lookalike", def.Name())
			return astmodel.TypeWalkerRemoveType, nil
		}

		return def.Name(), nil // Leave other cycles for now
	}

	return typeWalker
}

func (e embeddedResourceRemover) removeEmbeddedResourceProperties() (astmodel.Types, error) {
	result := make(astmodel.Types)

	visitor := e.MakeEmbeddedResourceRemovalTypeVisitor()
	typeWalker := e.NewResourceRemovalTypeWalker(visitor)

	// TODO: probably remove this
	// TODO: Actually we need this I think in order to guarantee we also discover types in the same order (and thus they get the same number assigned to them if they are used in multiple embedding contexts)
	// TODO: Nope this doesn't solve that issue, at least not by itself
	// Force ordering for determinism to ease debugging
	//var orderedTypes []astmodel.TypeDefinition
	//for _, def := range e.types {
	//	orderedTypes = append(orderedTypes, def)
	//}
	//sort.Slice(orderedTypes, func(i, j int) bool {
	//	left := orderedTypes[i]
	//	right := orderedTypes[j]
	//
	//	if left.Name().PackageReference.Equals(right.Name().PackageReference) {
	//		return left.Name().Name() < right.Name().Name()
	//	}
	//
	//	return left.Name().PackageReference.String() < right.Name().PackageReference.String()
	//})

	for _, def := range e.types {
		if astmodel.IsResourceDefinition(def) {
			ctx := resourceRemovalVisitorContext{resource: def.Name(), depth: 0, modifiedTypes: make(astmodel.Types)}
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

	return cleanupTypeNames(result, e.typeFlag)
}

// TODO: Method too long
// Only keep the suffix if the shorter (non-suffixed) name is also in use
func cleanupTypeNames(types astmodel.Types, flag astmodel.TypeFlag) (astmodel.Types, error) {
	renames := make(map[astmodel.TypeName]astmodel.TypeName)

	// Find all of the type names that have the flag we're interested in
	updatedNames := make(map[astmodel.TypeName]astmodel.TypeNameSet)
	for _, def := range types {
		if flag.IsOn(def.Type()) {
			embeddedName, err := splitEmbeddedResourceTypeName(def.Name())
			if err != nil {
				return nil, err
			}

			associatedNames := updatedNames[embeddedName.original]
			associatedNames = associatedNames.Add(def.Name())
			updatedNames[embeddedName.original] = associatedNames
		}
	}

	for original, associatedNames := range updatedNames {
		_, originalExists := types[original]

		// If original name doesn't exist and there is only one new name, rename new name to original name
		if !originalExists && len(associatedNames) == 1 {
			renames[associatedNames.Single()] = original

			klog.V(4).Infof("There are no usages of %q. Collapsing %q into the original for simplicity.", original, associatedNames.Single())
			continue
		}

		// TODO: This may be a duplicate of below?
		// If original name does exist and there is only one new name, remove context part of the new name (the resource name), but keep the suffix
		if originalExists && len(associatedNames) == 1 {
			associated := associatedNames.Single()
			embeddedName, err := splitEmbeddedResourceTypeName(associated)
			if err != nil {
				return nil, err
			}
			embeddedName.context = ""
			embeddedName.count = 0
			renames[associated] = makeCollapsedEmbeddedResourceTypeName(embeddedName)

			klog.V(4).Infof("There is only a single context %q is used in. Renaming it to %q for simplicity.", associated, renames[associated])
			continue
		}

		// Gather information about the associated types
		associatedCountPerContext := make(map[string]int)
		for associated := range associatedNames {
			embeddedName, err := splitEmbeddedResourceTypeName(associated)
			if err != nil {
				return nil, err
			}
			associatedCountPerContext[embeddedName.context] = associatedCountPerContext[embeddedName.context] + 1
		}

		// If all updated names share the same context, the context is not adding any disambiguation value so we can remove it
		if len(associatedCountPerContext) == 1 {
			for associated := range associatedNames {
				embeddedName, err := splitEmbeddedResourceTypeName(associated)
				if err != nil {
					return nil, err
				}
				embeddedName.context = ""
				renames[associated] = makeCollapsedEmbeddedResourceTypeName(embeddedName)
			}
			continue
		}

		// remove _0, which especially for the cases where there's only a single
		// kind of usage will make the type name much clearer
		for associated := range associatedNames {
			embeddedName, err := splitEmbeddedResourceTypeName(associated)
			if err != nil {
				return nil, err
			}

			possibleRename := makeCollapsedEmbeddedResourceTypeName(embeddedName)
			if !possibleRename.Equals(associated) {
				renames[associated] = possibleRename
			}
		}
	}

	return performRenames(types, renames, flag)
}

func performRenames(types astmodel.Types, renames map[astmodel.TypeName]astmodel.TypeName, flag astmodel.TypeFlag) (astmodel.Types, error) {
	result := make(astmodel.Types)

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

// TODO: We could push some of this stuff down into its own package I think... only advantage of that would be to reduce huge count of lines in this file. Thoughts?

type embeddedResourceTypeName struct {
	original astmodel.TypeName
	context  string // TODO: Rename this to resource?
	suffix   string
	count    int
}

// TODO: ugly name...
// TODO: really all of this is kinda ugly
func makeEmbeddedResourceTypeNameInternal(original astmodel.TypeName, context string, suffix string, count string) astmodel.TypeName {
	return astmodel.MakeTypeName(original.PackageReference, original.Name()+context+suffix+count)
}

func makeEmbeddedResourceTypeName(name embeddedResourceTypeName) astmodel.TypeName {
	if name.context == "" {
		panic("context cannot be empty when making embedded resource type name")
	}

	if name.suffix == "" {
		panic("suffix cannot be empty when making embedded resource type name")
	}

	nameContext := "_" + name.context
	suffix := "_" + name.suffix
	countString := fmt.Sprintf("_%d", name.count)
	return makeEmbeddedResourceTypeNameInternal(name.original, nameContext, suffix, countString)
}

func makeCollapsedEmbeddedResourceTypeName(name embeddedResourceTypeName) astmodel.TypeName {
	nameContext := ""
	if name.context != "" {
		nameContext = "_" + name.context
	}
	countString := ""
	if name.count > 0 {
		countString = fmt.Sprintf("_%d", name.count)
	}
	suffix := ""
	if name.suffix != "" {
		suffix = "_" + name.suffix
	}
	return makeEmbeddedResourceTypeNameInternal(name.original, nameContext, suffix, countString)
}

func splitEmbeddedResourceTypeName(name astmodel.TypeName) (embeddedResourceTypeName, error) {
	split := strings.Split(name.Name(), "_")
	if len(split) < 4 {
		return embeddedResourceTypeName{}, errors.Errorf("can't split embedded resource type name: %q didn't have 4 sections", name)
	}

	original := strings.Join(split[:len(split)-3], "_")
	resource := split[len(split)-3]
	suffix := split[len(split)-2]
	count, err := strconv.Atoi(split[len(split)-1])
	if err != nil {
		return embeddedResourceTypeName{}, err
	}

	return embeddedResourceTypeName{
		original: astmodel.MakeTypeName(name.PackageReference, original),
		context:  resource,
		suffix:   suffix,
		count:    count,
	}, nil
}

// TODO: Is this cleaner as module vars?
func requiredResourceProperties() []string {
	return []string{
		"Name",
		"Properties",

		// TODO: I think type actually needs to be required -- other kinds of embeds may not have a type but need to be treated
		// TODO: differently (see VNET -> Subnet)
		//"Type",
	}
}

func optionalResourceProperties() []string {
	return []string{
		"Type",
		"Etag",
		"Location",
		"Tags",
	}
}

func isTypeResourceLookalike(t astmodel.Type) bool {
	o, ok := astmodel.AsObjectType(t)
	if !ok {
		return false
	}

	return isObjectResourceLookalike(o)
}

func isObjectResourceLookalike(o *astmodel.ObjectType) bool {
	hasRequiredProperties := true
	for _, propName := range requiredResourceProperties() {
		_, hasProp := o.Property(astmodel.PropertyName(propName))
		hasRequiredProperties = hasRequiredProperties && hasProp
	}

	return hasRequiredProperties
}

// removeResourceLikeProperties examines an astmodel.ObjectType and determines if it looks like an Azure resource.
// An object is "like" a resource if it has "name", "type" and "properties" properties.
func removeResourceLikeProperties(o *astmodel.ObjectType) *astmodel.ObjectType {
	if !isObjectResourceLookalike(o) {
		// Doesn't match the shape we're looking for -- no change
		return o
	}

	result := o
	required := requiredResourceProperties()
	optional := optionalResourceProperties()

	for _, propName := range append(required, optional...) {
		result = result.WithoutProperty(astmodel.PropertyName(propName))
	}
	return result
}
