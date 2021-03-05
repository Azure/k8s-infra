/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"

	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
)

// removeEmptyObjectTypes removes empty ObjectTypes and references to them
func removeEmptyObjectTypes() PipelineStage {
	return MakePipelineStage(
		"removeEmptyObjectTypes",
		"Removes object types that contain no properties, as well as any references to that type",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {
			result := types
			for {
				toRemove := findEmptyObjectTypes(result)
				if len(toRemove) == 0 {
					break
				}

				var err error
				result, err = removeReferencesToTypes(result, toRemove)
				if err != nil {
					return nil, err
				}
			}

			return result, nil
		})
}

func findEmptyObjectTypes(types astmodel.Types) astmodel.TypeNameSet {
	result := make(astmodel.TypeNameSet)

	for _, def := range types {
		ot, ok := astmodel.AsObjectType(def.Type())
		if !ok {
			continue
		}

		// If there's still something "in" the object then we don't want to remove it
		if len(ot.Properties()) != 0 || len(ot.EmbeddedProperties()) != 0 {
			continue
		}

		klog.V(0).Infof("Removing %q as it has no properties", def.Name())
		result.Add(def.Name())
	}

	return result
}

func removeReferencesToTypes(types astmodel.Types, toRemove astmodel.TypeNameSet) (astmodel.Types, error) {
	result := make(astmodel.Types)
	visitor := makeRemovedTypeVisitor(toRemove)

	for _, def := range types {
		if toRemove.Contains(def.Name()) {
			continue
		}

		updatedDef, err := visitor.VisitDefinition(def, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "visiting definition %q", def.Name())
		}
		result.Add(updatedDef)
	}

	return result, nil
}

type visitorCtx struct {
	typeName *astmodel.TypeName
}

func makeRemovedTypeVisitor(toRemove astmodel.TypeNameSet) astmodel.TypeVisitor {
	visitor := astmodel.MakeTypeVisitor()
	//visitor.VisitObjectType = func(this *astmodel.TypeVisitor, it *astmodel.ObjectType, ctx interface{}) (astmodel.Type, error) {
	//	properties := it.Properties()
	//	embedded := it.EmbeddedProperties()
	//
	//	for _, property := range properties {
	//		typeName, isTypeName := astmodel.AsTypeName(property.PropertyType())
	//		if !isTypeName {
	//			continue
	//		}
	//
	//
	//		if toRemove.Contains(typeName) {
	//			// TODO: Do we need to worry about required-ness here (i.e. is removing an empty ObjectType bad if it's required?)
	//			it = it.WithoutProperty(property.PropertyName())
	//		}
	//	}
	//
	//	for _, property := range embedded {
	//		typeName, ok := astmodel.AsTypeName(property.PropertyType())
	//		if !ok {
	//			continue
	//		}
	//
	//		if toRemove.Contains(typeName) {
	//			it = it.WithoutEmbeddedProperty(typeName)
	//		}
	//	}
	//
	//	return astmodel.IdentityVisitOfObjectType(this, it, ctx)
	//}
	// TODO: This is basically copied from IDentityVisitOfObjectType, but since it has/needs a per-property context we can't use that
	visitor.VisitObjectType = func(this *astmodel.TypeVisitor, it *astmodel.ObjectType, ctx interface{}) (astmodel.Type, error) {
		// just map the property types
		var errs []error
		var newProps []*astmodel.PropertyDefinition
		for _, prop := range it.Properties() {
			ctx := &visitorCtx{}
			p, err := this.Visit(prop.PropertyType(), ctx)
			if err != nil {
				errs = append(errs, err)
			} else if ctx.typeName == nil || !toRemove.Contains(*ctx.typeName) {
				newProps = append(newProps, prop.WithType(p))
			} else if toRemove.Contains(*ctx.typeName) {
				klog.V(4).Infof("Removing property %q (referencing %q) as the type has no properties", prop.PropertyName(), *ctx.typeName)
			}
		}

		if len(errs) > 0 {
			return nil, kerrors.NewAggregate(errs)
		}

		// map the embedded types too
		var newEmbeddedProps []*astmodel.PropertyDefinition
		for _, prop := range it.EmbeddedProperties() {
			ctx := &visitorCtx{}
			p, err := this.Visit(prop.PropertyType(), ctx)
			if err != nil {
				errs = append(errs, err)
			} else if ctx.typeName != nil && !toRemove.Contains(*ctx.typeName) {
				newEmbeddedProps = append(newEmbeddedProps, prop.WithType(p))
			}
		}

		if len(errs) > 0 {
			return nil, kerrors.NewAggregate(errs)
		}

		result := it.WithoutProperties()
		result = result.WithProperties(newProps...)
		result = result.WithoutEmbeddedProperties()
		result, err := result.WithEmbeddedProperties(newEmbeddedProps...)
		if err != nil {
			return nil, err
		}

		return result, nil
	}
	visitor.VisitTypeName = func(this *astmodel.TypeVisitor, it astmodel.TypeName, ctx interface{}) (astmodel.Type, error) {
		typedCtx, ok := ctx.(*visitorCtx)
		if ok {
			// Safety check that we're not overwriting typeName
			if typedCtx.typeName != nil {
				return nil, errors.Errorf("would've overwritten ctx.typeName %q", typedCtx.typeName)
			}
			typedCtx.typeName = &it
		}
		return astmodel.IdentityVisitOfTypeName(this, it, ctx)
	}

	return visitor
}
