/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
)

// nameTypesForCRD - for CRDs all inner enums and objects and validated types must be named, so we do it here
func nameTypesForCRD(idFactory astmodel.IdentifierFactory) PipelineStage {

	return MakePipelineStage(
		"nameTypes",
		"Name inner types for CRD",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {

			result := make(astmodel.Types)

			// this is a little bit of a hack, better way to do it?
			getDescription := func(typeName astmodel.TypeName) []string {
				if typeDef, ok := types[typeName]; ok {
					return typeDef.Description()
				}

				return []string{}
			}

			for typeName, typeDef := range types {

				newDefs, err := nameInnerTypes(typeDef, idFactory, getDescription)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to name inner types")
				}

				for _, newDef := range newDefs {
					result.Add(newDef)
				}

				if _, ok := result[typeName]; !ok {
					// if we didn't regenerate the “input” type in nameInnerTypes then it won’t
					// have been added to the output; do it here
					result.Add(typeDef)
				}
			}

			return result, nil
		})
}

func nameInnerTypes(
	def astmodel.TypeDefinition,
	idFactory astmodel.IdentifierFactory,
	getDescription func(typeName astmodel.TypeName) []string) ([]astmodel.TypeDefinition, error) {

	var resultTypes []astmodel.TypeDefinition

	visitor := astmodel.MakeTypeVisitor()
	visitor.VisitEnumType = func(this *astmodel.TypeVisitor, it *astmodel.EnumType, ctx interface{}) (astmodel.Type, error) {
		nameHint := ctx.(string)

		enumName := astmodel.MakeTypeName(def.Name().PackageReference, idFactory.CreateEnumIdentifier(nameHint))

		namedEnum := astmodel.MakeTypeDefinition(enumName, it)
		namedEnum = namedEnum.WithDescription(getDescription(enumName))

		resultTypes = append(resultTypes, namedEnum)

		return namedEnum.Name(), nil
	}

	visitor.VisitValidatedType = func(this *astmodel.TypeVisitor, v astmodel.ValidatedType, ctx interface{}) (astmodel.Type, error) {
		// a validated type anywhere except directly under a property
		// must be named so that we can put the validations on it
		nameHint := ctx.(string)
		newElementType, err := this.Visit(v.ElementType(), nameHint+"_Validated")
		if err != nil {
			return nil, err
		}

		name := astmodel.MakeTypeName(def.Name().PackageReference, nameHint)
		validations := v.Validations().ToKubeBuilderValidations()
		namedType := astmodel.MakeTypeDefinition(name, newElementType).WithValidations(validations)
		resultTypes = append(resultTypes, namedType)
		return namedType.Name(), nil
	}

	visitor.VisitObjectType = func(this *astmodel.TypeVisitor, it *astmodel.ObjectType, ctx interface{}) (astmodel.Type, error) {
		nameHint := ctx.(string)

		var errs []error
		var props []*astmodel.PropertyDefinition
		// first map the inner types:
		for _, prop := range it.Properties() {

			propType := prop.PropertyType()

			// lift any validations out of the original type
			// and put them on the property
			if vt, ok := prop.PropertyType().(astmodel.ValidatedType); ok {
				propType = vt.ElementType()
				for _, validation := range vt.Validations().ToKubeBuilderValidations() {
					prop = prop.WithValidation(validation)
				}
			}

			newPropType, err := this.Visit(propType, nameHint+"_"+string(prop.PropertyName()))
			if err != nil {
				errs = append(errs, err)
			} else {
				props = append(props, prop.WithType(newPropType))
			}
		}

		if len(errs) > 0 {
			return nil, kerrors.NewAggregate(errs)
		}

		objectName := astmodel.MakeTypeName(def.Name().PackageReference, nameHint)

		namedObjectType := astmodel.MakeTypeDefinition(objectName, it.WithProperties(props...))
		namedObjectType = namedObjectType.WithDescription(getDescription(objectName))

		resultTypes = append(resultTypes, namedObjectType)

		return namedObjectType.Name(), nil
	}

	visitor.VisitResourceType = func(this *astmodel.TypeVisitor, it *astmodel.ResourceType, ctx interface{}) (astmodel.Type, error) {
		nameHint := ctx.(string)

		spec, err := this.Visit(it.SpecType(), nameHint+"_Spec")
		if err != nil {
			return nil, errors.Wrapf(err, "failed to name spec type %v", it.SpecType())
		}

		var status astmodel.Type
		if it.StatusType() != nil {
			status, err = this.Visit(it.StatusType(), nameHint+"_Status")
			if err != nil {
				return nil, errors.Wrapf(err, "failed to name status type %v", it.StatusType())
			}
		}

		resourceName := astmodel.MakeTypeName(def.Name().PackageReference, nameHint)

		// TODO: Should we have some better "clone" sort of thing in resource?
		newResource := astmodel.NewResourceType(spec, status).WithOwner(it.Owner())
		resource := astmodel.MakeTypeDefinition(resourceName, newResource)
		resource = resource.WithDescription(getDescription(resourceName))

		resultTypes = append(resultTypes, resource)

		return resource.Name(), nil
	}

	_, err := visitor.Visit(def.Type(), def.Name().Name())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to name inner types of %v", def.Name())
	}

	return resultTypes, nil
}
