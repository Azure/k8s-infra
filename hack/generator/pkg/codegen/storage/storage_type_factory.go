/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package storage

import (
	"fmt"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
)

// Each StorageTypeFactory is used to create storage types for a specific service
type StorageTypeFactory struct {
	types               astmodel.Types
	propertyConversions []propertyConversion
	visitor             astmodel.TypeVisitor
}

// NewStorageTypeFactory creates a new instance of StorageTypeFactory ready for use
func NewStorageTypeFactory() *StorageTypeFactory {
	result := &StorageTypeFactory{
		types: make(astmodel.Types),
	}

	result.propertyConversions = []propertyConversion{
		result.preserveKubernetesResourceStorageProperties,
		result.convertPropertiesForStorage,
	}

	return result
}

func (f *StorageTypeFactory) Add(d astmodel.TypeDefinition) {
	f.types.Add(d)
}

// StorageTypes returns all the storage types created by the factory, also returning any errors
// that occurred during construction
func (f *StorageTypeFactory) StorageTypes() (astmodel.Types, error) {
	visitor := f.makeStorageTypesVisitor()
	vc := MakeStorageTypesVisitorContext()
	types := make(astmodel.Types)
	var errs []error
	for _, d := range f.types {
		d := d

		if astmodel.ArmFlag.IsOn(d.Type()) {
			// Skip ARM definitions, we don't need to create storage variants of those
			continue
		}

		if _, ok := types.ResolveEnumDefinition(&d); ok {
			// Skip Enum definitions as we use the base type for storage
			continue
		}

		def, err := visitor.VisitDefinition(d, vc)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		finalDef := def.WithDescription(f.descriptionForStorageVariant(d))
		types.Add(finalDef)
	}

	if len(errs) > 0 {
		err := kerrors.NewAggregate(errs)
		return nil, err
	}

	return types, nil
}

// makeStorageTypesVisitor returns a TypeVisitor to do the creation of dedicated storage types
func (f *StorageTypeFactory) makeStorageTypesVisitor() astmodel.TypeVisitor {

	result := astmodel.MakeTypeVisitor()
	result.VisitValidatedType = f.visitValidatedType
	result.VisitTypeName = f.visitTypeName
	result.VisitObjectType = f.visitObjectType
	result.VisitResourceType = f.visitResourceType
	result.VisitFlaggedType = f.visitFlaggedType

	f.visitor = result

	return result
}

// A property conversion accepts a property definition and optionally applies a conversion to make
// the property suitable for use on a storage type. Conversions return nil if they decline to
// convert, deferring the conversion to another.
type propertyConversion = func(property *astmodel.PropertyDefinition, ctx StorageTypesVisitorContext) (*astmodel.PropertyDefinition, error)

func (f *StorageTypeFactory) visitValidatedType(this *astmodel.TypeVisitor, v *astmodel.ValidatedType, ctx interface{}) (astmodel.Type, error) {
	// strip all type validations from storage types,
	// act as if they do not exist
	return this.Visit(v.ElementType(), ctx)
}

func (f *StorageTypeFactory) visitTypeName(_ *astmodel.TypeVisitor, name astmodel.TypeName, ctx interface{}) (astmodel.Type, error) {
	visitorContext := ctx.(StorageTypesVisitorContext)

	// Resolve the type name to the actual referenced type
	actualType, err := f.types.FullyResolve(name)
	if err != nil {
		return nil, errors.Wrapf(err, "visiting type name %q", name)
	}

	// Check for property specific handling
	if visitorContext.property != nil {
		if et, ok := astmodel.AsEnumType(actualType); ok {
			// Property type refers to an enum, so we use the base type instead
			return et.BaseType(), nil
		}
	}

	// Map the type name into our storage namespace
	localRef, ok := name.PackageReference.AsLocalPackage()
	if !ok {
		return name, nil
	}

	storageRef := astmodel.MakeStoragePackageReference(localRef)
	visitedName := astmodel.MakeTypeName(storageRef, name.Name())
	return visitedName, nil
}

func (f *StorageTypeFactory) visitResourceType(
	tv *astmodel.TypeVisitor,
	resource *astmodel.ResourceType,
	ctx interface{}) (astmodel.Type, error) {

	// storage resource types do not need defaulter interface, they have no webhooks
	rsrc := resource.WithoutInterface(astmodel.DefaulterInterfaceName)

	return astmodel.IdentityVisitOfResourceType(tv, rsrc, ctx)
}

func (f *StorageTypeFactory) visitObjectType(
	_ *astmodel.TypeVisitor,
	object *astmodel.ObjectType,
	ctx interface{}) (astmodel.Type, error) {
	visitorContext := ctx.(StorageTypesVisitorContext)
	objectContext := visitorContext.forObject(object)

	var errs []error
	properties := object.Properties()
	for i, prop := range properties {
		p, err := f.makeStorageProperty(prop, objectContext)
		if err != nil {
			errs = append(errs, err)
		} else {
			properties[i] = p
		}
	}

	if len(errs) > 0 {
		err := kerrors.NewAggregate(errs)
		return nil, err
	}

	objectType := astmodel.NewObjectType().WithProperties(properties...)
	return astmodel.StorageFlag.ApplyTo(objectType), nil
}

// makeStorageProperty applies a conversion to make a variant of the property for use when
// serializing to storage
func (f *StorageTypeFactory) makeStorageProperty(
	prop *astmodel.PropertyDefinition,
	objectContext StorageTypesVisitorContext) (*astmodel.PropertyDefinition, error) {
	for _, conv := range f.propertyConversions {
		p, err := conv(prop, objectContext.forProperty(prop))
		if err != nil {
			// Something went wrong, return the error
			return nil, err
		}
		if p != nil {
			// We have the conversion we need, return it promptly
			return p, nil
		}
	}

	return nil, fmt.Errorf("failed to find a conversion for property %v", prop.PropertyName())
}

// preserveKubernetesResourceStorageProperties preserves properties required by the KubernetesResource interface as they're always required
func (f *StorageTypeFactory) preserveKubernetesResourceStorageProperties(
	prop *astmodel.PropertyDefinition,
	_ StorageTypesVisitorContext) (*astmodel.PropertyDefinition, error) {
	if astmodel.IsKubernetesResourceProperty(prop.PropertyName()) {
		// Keep these unchanged
		return prop, nil
	}

	// No opinion, defer to another conversion
	return nil, nil
}

func (f *StorageTypeFactory) convertPropertiesForStorage(
	prop *astmodel.PropertyDefinition,
	objectContext StorageTypesVisitorContext) (*astmodel.PropertyDefinition, error) {
	propertyType, err := factory.visitor.Visit(prop.PropertyType(), objectContext)
	if err != nil {
		return nil, err
	}

	p := prop.WithType(propertyType).
		MakeOptional().
		WithDescription("")

	return p, nil
}

func (f *StorageTypeFactory) visitFlaggedType(
	tv *astmodel.TypeVisitor,
	flaggedType *astmodel.FlaggedType,
	ctx interface{}) (astmodel.Type, error) {
	if flaggedType.HasFlag(astmodel.ArmFlag) {
		// We don't want to do anything with ARM types
		return flaggedType, nil
	}

	return astmodel.IdentityVisitOfFlaggedType(tv, flaggedType, ctx)
}

func (f *StorageTypeFactory) descriptionForStorageVariant(definition astmodel.TypeDefinition) []string {
	pkg := definition.Name().PackageReference.PackageName()

	result := []string{
		fmt.Sprintf("Storage version of %v.%v", pkg, definition.Name().Name()),
	}
	result = append(result, definition.Description()...)

	return result
}
