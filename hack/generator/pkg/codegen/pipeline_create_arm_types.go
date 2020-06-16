/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel/armconversion"
	"github.com/pkg/errors"
)

// TODO: There should probably actually be 2 passes here, one for Arm (creating types) and one for Kube (modifying existing types)
func createArmTypes(idFactory astmodel.IdentifierFactory) PipelineStage {
	return PipelineStage{
		id: "createArmTypes",
		description: "Create ARM types",
		Action: func(ctx context.Context, definitions astmodel.Types) (astmodel.Types, error) {
			return createArmTypesInternal(definitions, idFactory)
		},
	}
}

// createArmTypesInternal generates ARM types and adds them to the object graph
func createArmTypesInternal(
	definitions astmodel.Types,
	idFactory astmodel.IdentifierFactory) (astmodel.Types, error) {

	updatedDefs := make(astmodel.Types)

	// This function has multiple objectives:
	// 1. Create a duplicate hierarchy of ARM types.
	// 2. Remove ARM only properties (Name, Type) from the Kubernetes type. These properties will
	//    still exist on the ARM type.
	// 3. Add Kubernetes only properties (Owner, AzureName) to the Kubernetes type.
	// 4. Attach an interface to each Kubernetes type which contains methods to transform to/from
	//    the ARM types.

	// Do all the resources first - this ensures we avoid handling a spec before we've processed
	// its associated resource.
	for _, def := range definitions {
		// Special handling for resources because we need to modify their specs with extra properties
		if _, ok := def.Type().(*astmodel.ResourceType); ok {
			kubernetesSpecDef, armSpecDef, err := createArmAndKubeResourceSpecDefinitions(definitions, idFactory, def)
			if err != nil {
				return nil, err
			}

			updatedDefs.Add(*armSpecDef)
			updatedDefs.Add(*kubernetesSpecDef)
		}
	}

	// Process the remaining definitions
	for _, def := range definitions {
		// If it's a type which has already been handled (specs from above), skip it
		if _, ok := updatedDefs[def.Name()]; ok {
			continue
		}

		// Note: We would need to do something about type aliases here if they weren't already
		// removed earlier in the pipeline

		// Other types can be reused in both the Kube type graph and the ARM type graph, for
		// example enums which are effectively primitive types, all primitive types, etc.
		_, ok := def.Type().(*astmodel.ObjectType)
		if !ok {
			continue
		}

		kubernetesDef, armDef, err := createArmTypeDefsWithComplexPropertiesTypesUpdated(
			definitions,
			idFactory,
			def)
		if err != nil {
			return nil, err
		}
		updatedDefs.Add(*armDef)
		updatedDefs.Add(*kubernetesDef)
	}

	// Merge our updates in
	results := make(astmodel.Types)
	for _, updatedSpec := range updatedDefs {
		results.Add(updatedSpec)
	}
	for _, def := range definitions {
		_, ok := updatedDefs[def.Name()]
		if ok {
			continue // Already included above, so skip here to avoid duplicates
		}

		results.Add(def)
	}

	return results, nil
}

func removeValidations(t *astmodel.ObjectType) (*astmodel.ObjectType, error) {
	result := t

	for _, p := range t.Properties() {
		result = result.WithoutProperty(p.PropertyName())
	}

	for _, p := range t.Properties() {
		p = p.WithoutValidation()
		result = result.WithProperty(p)
	}

	return result, nil
}

func createArmTypeName(name astmodel.TypeName) astmodel.TypeName {
	return astmodel.MakeTypeName(name.PackageReference, name.Name()+"Arm")
}

type armKubeConversionHandler = func(t *astmodel.ObjectType) (*astmodel.ObjectType, error)

func createArmAndKubeDefinitions(
	idFactory astmodel.IdentifierFactory,
	typeDefinition astmodel.TypeDefinition,
	armStructHandlers []armKubeConversionHandler,
	kubeStructHandlers []armKubeConversionHandler,
	isResource bool) (*astmodel.TypeDefinition, *astmodel.TypeDefinition, error) {

	originalType, ok := typeDefinition.Type().(*astmodel.ObjectType)
	if !ok {
		return nil, nil, errors.Errorf("input type definition %q (%T) was not of expected type Object", typeDefinition.Name(), typeDefinition.Type())
	}

	armType := originalType
	var err error
	for _, handler := range armStructHandlers {
		armType, err = handler(armType)
		if err != nil {
			return nil, nil, err
		}
	}

	kubernetesType := originalType
	for _, handler := range kubeStructHandlers {
		kubernetesType, err = handler(kubernetesType)
		if err != nil {
			return nil, nil, err
		}
	}

	armTypeName := createArmTypeName(typeDefinition.Name())
	kubernetesTypeName := typeDefinition.Name()

	kubernetesType = kubernetesType.WithInterface(
		armconversion.NewArmTransformerImpl(
			armTypeName,
			armType,
			idFactory,
			isResource))

	armDef := astmodel.MakeTypeDefinition(
		armTypeName,
		armType)
	kubernetesDef := astmodel.MakeTypeDefinition(
		kubernetesTypeName,
		kubernetesType)

	armDef = armDef.WithDescription(typeDefinition.Description())
	kubernetesDef = kubernetesDef.WithDescription(typeDefinition.Description())

	return &kubernetesDef, &armDef, nil
}

func createArmAndKubeResourceSpecDefinitions(
	definitions astmodel.Types,
	idFactory astmodel.IdentifierFactory,
	resourceDef astmodel.TypeDefinition) (*astmodel.TypeDefinition, *astmodel.TypeDefinition, error) {

	resourceType := resourceDef.Type().(*astmodel.ResourceType)

	// The expectation is that the spec type is just a name
	specName, ok := resourceType.SpecType().(astmodel.TypeName)
	if !ok {
		return nil, nil, errors.Errorf("%s spec was not of type TypeName, instead: %T", resourceDef.Name(), resourceType.SpecType())
	}

	resourceSpecDef, ok := definitions[specName]
	if !ok {
		return nil, nil, errors.Errorf("couldn't find spec for resource %s", resourceDef.Name())
	}

	createOwnerProperty := func(t *astmodel.ObjectType) (*astmodel.ObjectType, error) {
		if resourceType.Owner() != nil {
			ownerField, err := createOwnerProperty(idFactory, resourceType.Owner())
			if err != nil {
				return nil, err
			}
			t = t.WithProperty(ownerField)
		}

		return t, nil
	}

	kubePropertyRemapper := func(t *astmodel.ObjectType) (*astmodel.ObjectType, error) {
		var hasName bool
		for _, prop := range t.Properties() {
			if prop.PropertyName() == astmodel.PropertyName("Name") {
				hasName = true
			}
		}

		// TODO: Right now the Kubernetes type has all of its standard requiredness (validations). If we want to allow
		// TODO: users to submit "just a name and owner" types we will have to strip some validation until
		// TODO: https://github.com/kubernetes-sigs/controller-tools/issues/461 is fixed
		kubernetesType := t.WithoutProperty(astmodel.PropertyName("Name")).WithoutProperty(astmodel.PropertyName("Type"))
		if hasName {
			kubernetesType = kubernetesType.WithProperty(armconversion.GetAzureNameField(idFactory))
		}

		return kubernetesType, nil
	}

	complexPropertyHandlerWrapper := func(t *astmodel.ObjectType) (*astmodel.ObjectType, error) {
		return complexPropertyHandler(t, definitions)
	}

	kubernetesDef, armDef, err := createArmAndKubeDefinitions(
		idFactory,
		resourceSpecDef,
		[]armKubeConversionHandler{removeValidations, complexPropertyHandlerWrapper},
		[]armKubeConversionHandler{kubePropertyRemapper, createOwnerProperty},
		true)
	if err != nil {
		return nil, nil, err
	}

	return kubernetesDef, armDef, nil
}

func convertArmPropertyTypeIfNeeded(definitions astmodel.Types, t astmodel.Type) (astmodel.Type, error) {
	switch concreteType := t.(type) {
	case astmodel.TypeName:
		def, ok := definitions[concreteType]
		if !ok {
			return nil, errors.Errorf("couldn't find type %v", concreteType)
		}

		if _, ok := def.Type().(*astmodel.ObjectType); ok {
			return createArmTypeName(def.Name()), nil
		} else {
			// We may or may not need to use an updated type name (i.e. if it's an aliased primitive type we can
			// just keep using that alias)
			updatedType, err := convertArmPropertyTypeIfNeeded(definitions, def.Type())
			if err != nil {
				return nil, err
			}

			if updatedType.Equals(def.Type()) {
				return t, nil
			} else {
				return createArmTypeName(def.Name()), nil
			}
		}

	case *astmodel.EnumType:
		return t, nil // Do nothing to enums
	case *astmodel.ResourceType:
		return t, nil // Do nothing to resources -- they have their own handling elsewhere
	case *astmodel.PrimitiveType:
		// No action required as the property is already the right name and type
		return t, nil
	case *astmodel.MapType:
		keyType, err := convertArmPropertyTypeIfNeeded(definitions, concreteType.KeyType())
		if err != nil {
			return nil, errors.Wrapf(err, "error converting to arm type for map key")
		}

		valueType, err := convertArmPropertyTypeIfNeeded(definitions, concreteType.ValueType())
		if err != nil {
			return nil, errors.Wrapf(err, "error converting to arm type for map value")
		}

		if keyType.Equals(concreteType.KeyType()) && valueType.Equals(concreteType.ValueType()) {
			return t, nil // No changes needed as both key/value haven't had special handling
		}

		return astmodel.NewMapType(keyType, valueType), nil
	case *astmodel.ArrayType:
		elementType, err := convertArmPropertyTypeIfNeeded(definitions, concreteType.Element())
		if err != nil {
			return nil, errors.Wrapf(err, "error converting to arm type for array")
		}

		if elementType.Equals(concreteType.Element()) {
			return t, nil // No changes needed as element wasn't changed
		}

		return astmodel.NewArrayType(elementType), nil
	case *astmodel.OptionalType:
		inner, err := convertArmPropertyTypeIfNeeded(definitions, concreteType.Element())
		if err != nil {
			return nil, err
		}

		if inner.Equals(concreteType.Element()) {
			return t, nil // No changes needed
		}

		return astmodel.NewOptionalType(inner), nil

	default:
		return nil, errors.Errorf("unhandled type %T", t)
	}
}

func complexPropertyHandler(t *astmodel.ObjectType, definitions astmodel.Types) (*astmodel.ObjectType, error) {
	result := t

	for _, prop := range result.Properties() {
		propType := prop.PropertyType()
		newType, err := convertArmPropertyTypeIfNeeded(definitions, propType)
		if err != nil {
			return nil, errors.Wrapf(err, "error converting type to arm type for property %s", prop.PropertyName())
		}

		if newType != propType {
			newProp := prop.WithType(newType)
			result = result.WithoutProperty(prop.PropertyName()).WithProperty(newProp)
		}
	}

	return result, nil
}

func createArmTypeDefsWithComplexPropertiesTypesUpdated( // TODO: ick, naming
	definitions astmodel.Types,
	idFactory astmodel.IdentifierFactory,
	objectDef astmodel.TypeDefinition) (*astmodel.TypeDefinition, *astmodel.TypeDefinition, error) {

	complexPropertyHandlerWrapper := func(t *astmodel.ObjectType) (*astmodel.ObjectType, error) {
		return complexPropertyHandler(t, definitions)
	}

	kubernetesDef, armDef, err := createArmAndKubeDefinitions(
		idFactory,
		objectDef,
		[]armKubeConversionHandler{removeValidations, complexPropertyHandlerWrapper},
		nil,
		false)

	return kubernetesDef, armDef, err
}

func createOwnerProperty(idFactory astmodel.IdentifierFactory, ownerTypeName *astmodel.TypeName) (*astmodel.PropertyDefinition, error) {

	knownResourceReferenceType := astmodel.MakeTypeName(
		astmodel.MakeGenRuntimePackageReference(),
		"KnownResourceReference")

	prop := astmodel.NewPropertyDefinition(
		idFactory.CreatePropertyName("owner", astmodel.Exported),
		"owner",
		knownResourceReferenceType)

	group, _, err := ownerTypeName.PackageReference.GroupAndPackage()
	if err != nil {
		return nil, err
	}

	prop = prop.WithTag("group", group).WithTag("kind", ownerTypeName.Name())
	prop = prop.WithValidation(astmodel.ValidateRequired()) // Owner is already required

	return prop, nil
}
