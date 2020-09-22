/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"fmt"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
)

// createStorageTypes returns a pipeline stage that creates dedicated storage types for each resource and nested object.
// Storage versions are created for *all* API versions to allow users of older versions of the operator to easily
// upgrade. This is of course a bit odd for the first release, but defining the approach from day one is useful.
func createStorageTypes() PipelineStage {
	return PipelineStage{
		id:          "createStorage",
		description: "Create storage versions of CRD types",
		Action: func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {

			storageTypes := make(astmodel.Types)
			visitor := makeStorageTypesVisitor(types)
			vc := makeStorageTypesVisitorContext()
			var errs []error
			for _, d := range types {
				d := d

				if types.IsArmDefinition(&d) {
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

				finalDef := def.WithDescription(descriptionForStorageVariant(d))
				storageTypes[finalDef.Name()] = finalDef
			}

			if len(errs) > 0 {
				err := kerrors.NewAggregate(errs)
				return nil, err
			}

			types.AddTypes(storageTypes)

			return types, nil
		},
	}
}

// makeStorageTypesVisitor returns a TypeVisitor to do the creation of dedicated storage types
func makeStorageTypesVisitor(types astmodel.Types) astmodel.TypeVisitor {
	factory := &StorageTypeFactory{
		types: types,
	}

	result := astmodel.MakeTypeVisitor()
	result.VisitTypeName = factory.visitTypeName
	result.VisitObjectType = factory.visitObjectType
	result.VisitArmType = factory.visitArmType

	return result
}

type StorageTypeFactory struct {
	types astmodel.Types
}

func (factory *StorageTypeFactory) visitTypeName(_ *astmodel.TypeVisitor, name astmodel.TypeName, ctx interface{}) (astmodel.Type, error) {
	vc := ctx.(StorageTypesVisitorContext)

	// Resolve the type name to the actual referenced type
	actualDefinition, actualDefinitionFound := factory.types[name]

	// Check for property specific handling
	if visitorContext.property != nil && actualDefinitionFound {
		if et, ok := actualDefinition.Type().(*astmodel.EnumType); ok {
			// Property type refers to an enum, so we use the base type instead
			return et.BaseType(), nil
		}
	}

	// Map the type name into our storage namespace
	visitedName, err := factory.mapTypeNameIntoStoragePackage(name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to map name into storage namespace")
	}

	return visitedName, nil
}

func (factory *StorageTypeFactory) visitObjectType(
	visitor *astmodel.TypeVisitor,
	object *astmodel.ObjectType,
	ctx interface{}) (astmodel.Type, error) {
	vc := ctx.(StorageTypesVisitorContext)
	oc := vc.forObject(object)

	var errs []error
	properties := object.Properties()
	for i, prop := range properties {
		propertyType, err := visitor.Visit(prop.PropertyType(), oc.forProperty(prop))
		if err != nil {
			errs = append(errs, err)
		} else {
			p := factory.makeStorageProperty(prop, propertyType)
			properties[i] = p
		}
	}

	if len(errs) > 0 {
		err := kerrors.NewAggregate(errs)
		return nil, err
	}

	ot := astmodel.NewObjectType().WithProperties(properties...)
	return astmodel.NewStorageType(*ot), nil
}

func (factory *StorageTypeFactory) makeStorageProperty(
	prop *astmodel.PropertyDefinition,
	propertyType astmodel.Type) *astmodel.PropertyDefinition {
	p := prop.WithType(propertyType).
		MakeOptional().
		WithoutValidation().
		WithDescription("")

	// If the property is *string we can simplify it to string
	// This makes other generated code simpler
	if ot, ok := p.PropertyType().(*astmodel.OptionalType); ok {
		if ot.Element().Equals(astmodel.StringType) {
			p = p.WithType(astmodel.StringType)
		}
	}

	return p
}

// mapTypeNameIntoStoragePackage maps an existing type name into the right package for the matching storage type
// Returns the original instance for any type that does not need to be mapped to storage
func (factory *StorageTypeFactory) mapTypeNameIntoStoragePackage(name astmodel.TypeName) (astmodel.TypeName, error) {
	localRef, ok := name.PackageReference.AsLocalPackage()
	if !ok {
		// Don't need to map non-local packages
		return name, nil
	}

	storagePackage := astmodel.MakeStoragePackageReference(localRef)
	newName := astmodel.MakeTypeName(storagePackage, name.Name())
	return newName, nil
}

func (factory *StorageTypeFactory) visitArmType(
	_ *astmodel.TypeVisitor,
	it *astmodel.ArmType,
	_ interface{}) (astmodel.Type, error) {
	// We don't want to do anything with ARM types
	return it, nil
}

func descriptionForStorageVariant(definition astmodel.TypeDefinition) []string {
	pkg := definition.Name().PackageReference.PackageName()

	result := []string{
		fmt.Sprintf("Storage version of %v.%v", pkg, definition.Name().Name()),
	}
	result = append(result, definition.Description()...)

	return result
}

type StorageTypesVisitorContext struct {
	object   *astmodel.ObjectType
	property *astmodel.PropertyDefinition
}

func makeStorageTypesVisitorContext() StorageTypesVisitorContext {
	return StorageTypesVisitorContext{}
}

func (context StorageTypesVisitorContext) forObject(object *astmodel.ObjectType) StorageTypesVisitorContext {
	context.object = object
	context.property = nil
	return context
}

func (context StorageTypesVisitorContext) forProperty(property *astmodel.PropertyDefinition) StorageTypesVisitorContext {
	context.property = property
	return context
}
