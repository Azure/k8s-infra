/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"fmt"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
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
	localRef, ok := name.PackageReference.AsLocalPackage()
	if !ok {
		return name, nil
	}

	storageRef := astmodel.MakeStoragePackageReference(localRef)
	visitedName := astmodel.MakeTypeName(storageRef, name.Name())
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

	return p
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
