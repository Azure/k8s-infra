package codegen

import (
	"context"
	"fmt"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
)

// createStorageTypes returns a pipeline stage that creates dedicated storage types for each resource and nested object.
// Storage versions are created for *all* API versions to allow users of older versions of the operator to easily
// upgrade. This is of course a bit odd for the first release, but defining the approach from day one is useful.
func createStorageTypes() PipelineStage {
	return PipelineStage{
		id: "createStorage",
		description: "Create storage versions of CRD types",
		Action: func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {
			t := make(astmodel.Types)
			for n, d := range types {
				t[n] = d.WithDescription(descriptionForStorageVariant(d))
			}

			visitor := makeStorageTypesVisitor(t)
			transformedTypes := visitor.VisitAll(t, nil)
			storageTypes := transformedTypes.Where(
				func(def astmodel.TypeDefinition) bool {
					return !types.Contains(def)
				})
			types.AddAll(storageTypes)
			return types, nil
		},
	}
}

// makeStorageTypesVisitor returns a TypeVisitor to do the creation of dedicated storage types
func makeStorageTypesVisitor(types astmodel.Types) astmodel.TypeVisitor {
	propertyVisitor := makeStoragePropertiesVisitor(types)

	result := astmodel.MakeTypeVisitor()

	result.VisitTypeName = func(_ *astmodel.TypeVisitor, name astmodel.TypeName, ctx interface{}) astmodel.Type {
		return mapTypeName(name, types)
	}

	result.VisitObjectType = func(visitor *astmodel.TypeVisitor, object *astmodel.ObjectType, _ interface{}) astmodel.Type {
		return mapObjectForStorage(object, &propertyVisitor)
	}

	return result
}

// makeStoragePropertiesVisitor returns a visitor used to modify the properties within each resource or object
func makeStoragePropertiesVisitor(types astmodel.Types) astmodel.TypeVisitor {
	result := astmodel.MakeTypeVisitor()

	result.VisitTypeName = func(_ *astmodel.TypeVisitor, name astmodel.TypeName, ctx interface{}) astmodel.Type {
		return mapTypeNameForProperty(name, types)
	}

	result.VisitPrimitive = func(_ *astmodel.TypeVisitor, t *astmodel.PrimitiveType, _ interface{}) astmodel.Type {
		return astmodel.NewOptionalType(t)
	}

	return result
}

// mapTypeName maps an existing type name into the right package for the matching storage type
// Returns the original instance for any type that does not need to be mapped to storage
func mapTypeName(name astmodel.TypeName, types astmodel.Types) astmodel.TypeName {
	if def, ok := types[name]; ok {
		// If the type name refers to an enum, we don't need it
		if _, ok = def.Type().(*astmodel.EnumType); ok {
			return name
		}
	}

	group, pkg, err := name.PackageReference.GroupAndPackage()
	if err != nil {
		panic(err)
	}

	return astmodel.MakeTypeName(
		astmodel.MakeLocalPackageReference(group, pkg+"s"),
		name.Name())
}

// mapTypeNameForProperty transforms the type of a property into a simpler form
// Enum types are converted into the underlying base type
func mapTypeNameForProperty(name astmodel.TypeName, types astmodel.Types) astmodel.Type {
	if d, ok := types[name]; ok {
		if e, ok := d.Type().(*astmodel.EnumType); ok {
			return e.BaseType()
		}
	}

	return mapTypeName(name, types)
}

// mapObjectForStorage converts an existing object type into a more permissive form for serialization
func mapObjectForStorage(object *astmodel.ObjectType, propertyVisitor *astmodel.TypeVisitor) astmodel.Type {
	properties := object.Properties()
	for i, prop := range properties {
		propertyType := propertyVisitor.Visit(prop.PropertyType(), nil)
		p := prop.WithType(propertyType).
			MakeOptional().
			WithoutValidation().
			WithDescription("")
		properties[i] = p
	}

	return object.WithProperties(properties...)
}

func descriptionForStorageVariant(definition astmodel.TypeDefinition) []string {
	_, pkg, err := definition.Name().PackageReference.GroupAndPackage()
	if err != nil {
		panic(err)
	}

	result := []string{
		fmt.Sprintf("Storage version of %v.%v", pkg, definition.Name().Name()),
	}
	result = append(result, definition.Description()...)

	return result
}
