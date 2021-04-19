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

// removeEmbeddedResources uses a variety of heuristics to remove resources that are embedded inside other resources.
// There are a number of different kinds of embeddings:
// 1. A "Properties" embedding. When we process the Azure JSON schema/Swagger we manufacture a "Spec"
//    type that doesn't exist in the JSON schema/Swagger. In the JSON schema the resource itself must comply with ARM
//    resource requirements, meaning that all of the RP specific properties are stored in the "Properties"
//    property which for the sake of example we will say has type "R1Properties".
//    Other resources which have a property somewhere in their type hierarchy with that same "R1Properties"
//    type are actually embedding the R1 resource entirely inside themselves. Since the R1 resource is its own
//    resource it doesn't make sense to have it embedded inside another resource in Kubernetes. These embeddings
//    should really just be cross resource references. This pipeline finds such embeddings and removes them. A concrete
//    example of one such embedding is
//    v20181001 Microsoft.Networking Connection.Spec.Properties.LocalNetworkGateway2.Properties.
//    The LocalNetworkGateway2 property is of type "LocalNetworkGateway" which is itself a resource.
//    The ideal shape of Connection.Spec.Properties.LocalNetworkGate2 would just be a reference to a
//    LocalNetworkGateway resource. TODO: Talk about how sure we are of this
// 2. A subresource embedding. For the same reasons above, embedded subresources don't make sense in Kubernetes.
//    In the case of embedded subresources, the ideal shape would be a complete removal of the reference. We forbid
//    parent resources directly referencing child resources as it complicates the Watches scenario for each resource
//    reconciler. It's also not a common pattern in Kubernetes - usually you can identify children for a
//    given parent via a label. An example of this type of embedding is
//    v20180601 Microsoft.Networking RouteTable.Spec.Properties.Routes. The Routes property is of type RouteTableRoutes
//    which is a child resource of RouteTable.
// Note that even though the above examples do not include Status types, the same rules apply to Status types, with
// the only difference being that for Status types the resource reference in Swagger (the source of the Status types)
// is to the Status type (as opposed to the "Properties" type for Spec).
func removeEmbeddedResources() PipelineStage {
	return MakePipelineStage(
		"removeEmbeddedResources",
		"Removes properties that point to embedded resources",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {

			remover, err := makeEmbeddedResourceRemover(types)
			if err != nil {
				return nil, err
			}
			return remover.removeEmbeddedResourceProperties()
		})
}

// findSubResourcePropertiesTypeNames finds the "Properties" type of each subresource and returns a map of
// parent resource to subresource "Properties" type names.
func findSubResourcePropertiesTypeNames(types astmodel.Types) (map[astmodel.TypeName]astmodel.TypeNameSet, error) {
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

// findAllResourcePropertiesTypes finds the "Properties" type for each resource. The result is a astmodel.TypeNameSet containing
// each resources "Properties" type.
func findAllResourcePropertiesTypes(types astmodel.Types) (astmodel.TypeNameSet, error) {
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

// findAllResourceStatusTypes finds the astmodel.TypeName's of each resources Status type. If the resource does not have a Status type then
// that TypeName is not included in the resulting astmodel.TypeNameSet (obviously).
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

func makeEmbeddedResourceRemover(types astmodel.Types) (embeddedResourceRemover, error) {
	resourceStatusTypes := findAllResourceStatusTypes(types)
	resourceToSubresourceMap, err := findSubResourcePropertiesTypeNames(types)
	if err != nil {
		return embeddedResourceRemover{}, errors.Wrap(err, "couldn't find subresource \"Properties\" type names")
	}

	resourcePropertiesTypes, err := findAllResourcePropertiesTypes(types)
	if err != nil {
		return embeddedResourceRemover{}, errors.Wrap(err, "couldn't find resource \"Properties\" type names")
	}

	remover := embeddedResourceRemover{
		types:                    types,
		resourceToSubresourceMap: resourceToSubresourceMap,
		resourcePropertiesTypes:  resourcePropertiesTypes,
		resourceStatusTypes:      resourceStatusTypes,
		typeSuffix:               "SubResourceEmbedded",
		typeFlag:                 astmodel.TypeFlag("embeddedSubResource"), // TODO: Instead of flag we could just use a map here if we wanted
	}

	return remover, nil
}

func (e embeddedResourceRemover) MakeEmbeddedResourceRemovalTypeVisitor() astmodel.TypeVisitor {
	visitor := astmodel.MakeTypeVisitor()
	visitor.VisitObjectType = func(this *astmodel.TypeVisitor, it *astmodel.ObjectType, ctx interface{}) (astmodel.Type, error) {
		typedCtx := ctx.(resourceRemovalVisitorContext)

		if typedCtx.depth <= 2 {
			// If we are not at sufficient depth, don't bother checking for subresource references. The resource itself and its spec type
			// will not refer to a subresource. There are some instances of resources (such as Microsoft.Web v20160801 Sites) where the resource
			// and some child resources reuse the same "Properties" type, which could false some of the logic below
			return astmodel.IdentityVisitOfObjectType(this, it, ctx)
		}

		// Before visiting, check if any properties are just referring to one of our sub-resources and remove them
		subResources := e.resourceToSubresourceMap[typedCtx.resource]
		for _, prop := range it.Properties() {
			propTypeName, ok := astmodel.AsTypeName(prop.PropertyType())
			if !ok {
				continue
			}

			// TODO: This is currently no different than the below, but it likely will evolve to be different over time
			if subResources.Contains(propTypeName) {
				klog.V(5).Infof("Removing resource %q reference to subresource %q on property %q", typedCtx.resource, propTypeName, prop.PropertyName())
				it = removeResourceLikeProperties(it)
				continue
			}

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
	typeWalker.AfterVisit = func(original astmodel.TypeDefinition, updated astmodel.TypeDefinition, ctx interface{}) (astmodel.TypeDefinition, error) {
		typedCtx := ctx.(resourceRemovalVisitorContext)

		if !original.Name().Equals(updated.Name()) {
			panic(fmt.Sprintf("Unexpected name mismatch during type walk: %q -> %q", original.Name(), updated.Name()))
		}

		if !original.Type().Equals(updated.Type()) {
			flaggedType := e.typeFlag.ApplyTo(updated.Type())
			var newName astmodel.TypeName
			exists := false
			for count := 0; ; count++ {
				newName = embeddedResourceTypeName{original: original.Name(), context: typedCtx.resource.Name(), suffix: e.typeSuffix, count: count}.ToTypeName()
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

	typeWalker.ShouldRemoveCycle = func(def astmodel.TypeDefinition, ctx interface{}) (bool, error) {
		// If we're about to walk a cycle that is to a known resource type, just skip it entirely
		if e.resourcePropertiesTypes.Contains(def.Name()) || e.resourceStatusTypes.Contains(def.Name()) {
			return true, nil
		}

		if isTypeResourceLookalike(def.Type()) {
			klog.V(5).Infof("Type %q is a resource lookalike", def.Name())
			return true, nil
		}

		return false, nil // Leave other cycles for now
	}

	return typeWalker
}

func (e embeddedResourceRemover) removeEmbeddedResourceProperties() (astmodel.Types, error) {
	result := make(astmodel.Types)

	visitor := e.MakeEmbeddedResourceRemovalTypeVisitor()
	typeWalker := e.NewResourceRemovalTypeWalker(visitor)

	for _, def := range e.types {
		if astmodel.IsResourceDefinition(def) {
			// TODO: Bit awkward that we're modifying typewalker here... maybe should bring back initial ctx param?
			typeWalker.MakeContext = func(it astmodel.TypeName, ctx interface{}) (interface{}, error) {
				if ctx == nil {
					return resourceRemovalVisitorContext{resource: def.Name(), depth: 0, modifiedTypes: make(astmodel.Types)}, nil
				}
				typedCtx := ctx.(resourceRemovalVisitorContext)
				return typedCtx.WithMoreDepth(), nil
			}

			updatedTypes, err := typeWalker.Walk(def)
			if err != nil {
				return nil, err
			}

			for _, newDef := range updatedTypes {
				err := result.AddAllowDuplicates(newDef)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return simplifyTypeNames(result, e.typeFlag)
}

type renamer struct {
	types astmodel.Types
}

func (r renamer) singleNameToOriginalName(original astmodel.TypeName, associatedNames astmodel.TypeNameSet) (map[astmodel.TypeName]astmodel.TypeName, error) {
	_, originalExists := r.types[original]
	if originalExists || len(associatedNames) != 1 {
		return nil, nil
	}

	klog.V(4).Infof("There are no usages of %q. Collapsing %q into the original for simplicity.", original, associatedNames.Single())
	renames := make(map[astmodel.TypeName]astmodel.TypeName)
	renames[associatedNames.Single()] = original

	return renames, nil
}

func (r renamer) associatedNameUsedInSingleContext(_ astmodel.TypeName, associatedNames astmodel.TypeNameSet) (map[astmodel.TypeName]astmodel.TypeName, error) {
	if len(associatedNames) != 1 {
		return nil, nil
	}

	renames := make(map[astmodel.TypeName]astmodel.TypeName)
	associated := associatedNames.Single()
	embeddedName, err := parseContextualTypeName(associated)
	if err != nil {
		return nil, err
	}
	embeddedName.context = ""
	embeddedName.count = 0
	renames[associated] = embeddedName.ToSimplifiedTypeName()
	klog.V(4).Infof("There is only a single context %q is used in. Renaming it to %q for simplicity.", associated, renames[associated])

	return renames, nil
}

func (r renamer) associatedNameUsedInSingleContextMultipleCounts(_ astmodel.TypeName, associatedNames astmodel.TypeNameSet) (map[astmodel.TypeName]astmodel.TypeName, error) {
	// Gather information about the associated types
	associatedCountPerContext := make(map[string]int)
	for associated := range associatedNames {
		embeddedName, err := parseContextualTypeName(associated)
		if err != nil {
			return nil, err
		}
		associatedCountPerContext[embeddedName.context] = associatedCountPerContext[embeddedName.context] + 1
	}

	if len(associatedCountPerContext) != 1 {
		return nil, nil
	}

	// If all updated names share the same context, the context is not adding any disambiguation value so we can remove it
	renames := make(map[astmodel.TypeName]astmodel.TypeName)
	for associated := range associatedNames {
		embeddedName, err := parseContextualTypeName(associated)
		if err != nil {
			return nil, err
		}
		embeddedName.context = ""
		renames[associated] = embeddedName.ToSimplifiedTypeName()
	}

	return renames, nil
}

func (r renamer) simplifyRemainingAssociatedNames(_ astmodel.TypeName, associatedNames astmodel.TypeNameSet) (map[astmodel.TypeName]astmodel.TypeName, error) {
	// remove _0, which especially for the cases where there's only a single
	// kind of usage will make the type name much clearer
	renames := make(map[astmodel.TypeName]astmodel.TypeName)
	for associated := range associatedNames {
		embeddedName, err := parseContextualTypeName(associated)
		if err != nil {
			return nil, err
		}

		possibleRename := embeddedName.ToSimplifiedTypeName()
		if !possibleRename.Equals(associated) {
			renames[associated] = possibleRename
		}
	}

	return renames, nil
}

func (r renamer) performRenames(
	renames map[astmodel.TypeName]astmodel.TypeName,
	flag astmodel.TypeFlag) (astmodel.Types, error) {

	result := make(astmodel.Types)

	renamingVisitor := astmodel.MakeTypeVisitor()
	renamingVisitor.VisitTypeName = func(this *astmodel.TypeVisitor, it astmodel.TypeName, ctx interface{}) (astmodel.Type, error) {
		if newName, ok := renames[it]; ok {
			return astmodel.IdentityVisitOfTypeName(this, newName, ctx)
		}
		return astmodel.IdentityVisitOfTypeName(this, it, ctx)
	}

	for _, def := range r.types {
		updatedDef, err := renamingVisitor.VisitDefinition(def, nil)
		if err != nil {
			return nil, err
		}
		// TODO: If we don't remove this, something is causing these types to not be emitted.
		// TODO: Unsure what that is... should track it down
		updatedType, err := flag.RemoveFrom(updatedDef.Type())
		if err != nil {
			return nil, err
		}
		result.Add(updatedDef.WithType(updatedType))
	}

	return result, nil
}

// simplifyTypeNames simplifies contextual type names if possible.
func simplifyTypeNames(types astmodel.Types, flag astmodel.TypeFlag) (astmodel.Types, error) {
	renames := make(map[astmodel.TypeName]astmodel.TypeName)

	// Find all of the type names that have the flag we're interested in
	updatedNames := make(map[astmodel.TypeName]astmodel.TypeNameSet)
	for _, def := range types {
		if flag.IsOn(def.Type()) {
			embeddedName, err := parseContextualTypeName(def.Name())
			if err != nil {
				return nil, err
			}

			associatedNames := updatedNames[embeddedName.original]
			associatedNames = associatedNames.Add(def.Name())
			updatedNames[embeddedName.original] = associatedNames
		}
	}

	renamer := renamer{types: types}
	renameActions := []func(_ astmodel.TypeName, associatedNames astmodel.TypeNameSet) (map[astmodel.TypeName]astmodel.TypeName, error){
		renamer.singleNameToOriginalName,
		renamer.associatedNameUsedInSingleContext,
		renamer.associatedNameUsedInSingleContextMultipleCounts,
		renamer.simplifyRemainingAssociatedNames,
	}

	for original, associatedNames := range updatedNames {
		for _, action := range renameActions {
			result, err := action(original, associatedNames)
			if err != nil {
				return nil, err
			}

			if result != nil {
				// Add renames
				for oldName, newName := range result {
					renames[oldName] = newName
				}
				break
			}
		}
	}

	return renamer.performRenames(renames, flag)
}

type embeddedResourceTypeName struct {
	original astmodel.TypeName
	context  string
	suffix   string
	count    int
}

func (e embeddedResourceTypeName) ToTypeName() astmodel.TypeName {
	if e.context == "" {
		panic("context cannot be empty when making embedded resource type name")
	}

	if e.suffix == "" {
		panic("suffix cannot be empty when making embedded resource type name")
	}

	nameContext := "_" + e.context
	suffix := "_" + e.suffix
	countString := fmt.Sprintf("_%d", e.count)
	return makeContextualTypeName(e.original, nameContext, suffix, countString)
}

func (e embeddedResourceTypeName) ToSimplifiedTypeName() astmodel.TypeName {
	nameContext := ""
	if e.context != "" {
		nameContext = "_" + e.context
	}
	countString := ""
	if e.count > 0 {
		countString = fmt.Sprintf("_%d", e.count)
	}
	suffix := ""
	if e.suffix != "" {
		suffix = "_" + e.suffix
	}
	return makeContextualTypeName(e.original, nameContext, suffix, countString)
}

func makeContextualTypeName(original astmodel.TypeName, context string, suffix string, count string) astmodel.TypeName {
	return astmodel.MakeTypeName(original.PackageReference, original.Name()+context+suffix+count)
}

func parseContextualTypeName(name astmodel.TypeName) (embeddedResourceTypeName, error) {
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
