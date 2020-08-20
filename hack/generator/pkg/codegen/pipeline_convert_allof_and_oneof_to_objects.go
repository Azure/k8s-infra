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
)

// convertAllOfAndOneOfToObjects reduces the AllOfType and OneOfType to ObjectType
func convertAllOfAndOneOfToObjects(idFactory astmodel.IdentifierFactory) PipelineStage {
	return MakePipelineStage(
		"allof-anyof-objects",
		"Convert allOf and oneOf to object types",
		func(ctx context.Context, defs astmodel.Types) (result astmodel.Types, err error) {
			visitor := astmodel.MakeTypeVisitor()
			synthesizer := synthesizer{idFactory, defs}

			// the context here is whether or not to flatten OneOf types
			// we only want this inside AllOfType, so it should only go down one level

			const flattenOneOf = "flattenOneOf"

			visitor.VisitTypeName = func(this *astmodel.TypeVisitor, it astmodel.TypeName, ctx interface{}) astmodel.Type {
				if ctx != flattenOneOf {
					return it
				}

				resolved, err := synthesizer.fullyResolve(it)
				if err != nil {
					panic(err)
				}

				if _, ok := resolved.(astmodel.OneOfType); ok {
					return resolved
				}

				return it
			}

			originalVisitAllOf := visitor.VisitAllOfType
			visitor.VisitAllOfType = func(this *astmodel.TypeVisitor, it astmodel.AllOfType, _ interface{}) astmodel.Type {
				// process children first
				result := originalVisitAllOf(this, it, flattenOneOf)

				if resultAllOf, ok := result.(astmodel.AllOfType); ok {
					object, err := synthesizer.allOfObject(resultAllOf)
					if err != nil {
						panic(err)
					}

					result = object
				}

				// we might end up with something that requires re-visiting
				return this.Visit(result, flattenOneOf)
			}

			originalVisitOneOf := visitor.VisitOneOfType
			visitor.VisitOneOfType = func(this *astmodel.TypeVisitor, it astmodel.OneOfType, _ interface{}) astmodel.Type {
				// process children first
				result := originalVisitOneOf(this, it, nil)

				if resultOneOf, ok := result.(astmodel.OneOfType); ok {
					object, err := synthesizer.oneOfObject(resultOneOf)
					if err != nil {
						panic(err)
					}

					result = object
				}

				return this.Visit(result, nil)
			}

			// these all pass nil context since we only want flattening inside AllOf
			originalVisitObject := visitor.VisitObjectType
			visitor.VisitObjectType = func(this *astmodel.TypeVisitor, it *astmodel.ObjectType, _ interface{}) astmodel.Type {
				return originalVisitObject(this, it, nil)
			}

			originalVisitArray := visitor.VisitArrayType
			visitor.VisitArrayType = func(this *astmodel.TypeVisitor, it *astmodel.ArrayType, _ interface{}) astmodel.Type {
				return originalVisitArray(this, it, nil)
			}

			originalVisitOptional := visitor.VisitOptionalType
			visitor.VisitOptionalType = func(this *astmodel.TypeVisitor, it *astmodel.OptionalType, _ interface{}) astmodel.Type {
				return originalVisitOptional(this, it, nil)
			}

			originalVisitMap := visitor.VisitMapType
			visitor.VisitMapType = func(this *astmodel.TypeVisitor, it *astmodel.MapType, _ interface{}) astmodel.Type {
				return originalVisitMap(this, it, nil)
			}

			originalVisitResource := visitor.VisitResourceType
			visitor.VisitResourceType = func(this *astmodel.TypeVisitor, it *astmodel.ResourceType, _ interface{}) astmodel.Type {
				return originalVisitResource(this, it, nil)
			}

			// convert any panics we threw above back to errs
			// this sets the named error type in the result
			defer func() {
				if caught := recover(); caught != nil {
					if e, ok := caught.(error); ok {
						err = e
					} else {
						panic(caught)
					}
				}
			}()

			result = make(astmodel.Types)
			for _, def := range defs {
				result.Add(visitor.VisitDefinition(def, nil))
			}

			return // result, err named return types
		})
}

type synthesizer struct {
	idFactory astmodel.IdentifierFactory
	defs      astmodel.Types
}

func (synthesizer synthesizer) oneOfObject(oneOf astmodel.OneOfType) (astmodel.Type, error) {
	// If there's more than one option, synthesize a type.
	// Note that this is required because Kubernetes CRDs do not support OneOf the same way
	// OpenAPI does, see https://github.com/Azure/k8s-infra/issues/71
	var properties []*astmodel.PropertyDefinition
	propertyDescription := "mutually exclusive with all other properties"

	for i, t := range oneOf.Types() {
		switch concreteType := t.(type) {
		case astmodel.TypeName:
			propertyName := synthesizer.idFactory.CreatePropertyName(concreteType.Name(), astmodel.Exported)

			// JSON name is unimportant here because we will implement the JSON marshaller anyway,
			// but we still need it for controller-gen
			jsonName := synthesizer.idFactory.CreateIdentifier(concreteType.Name(), astmodel.NotExported)
			property := astmodel.NewPropertyDefinition(propertyName, jsonName, concreteType).MakeOptional().WithDescription(propertyDescription)
			properties = append(properties, property)
		case *astmodel.EnumType:
			// TODO: This name sucks but what alternative do we have?
			name := fmt.Sprintf("enum%v", i)
			propertyName := synthesizer.idFactory.CreatePropertyName(name, astmodel.Exported)

			// JSON name is unimportant here because we will implement the JSON marshaller anyway,
			// but we still need it for controller-gen
			jsonName := synthesizer.idFactory.CreateIdentifier(name, astmodel.NotExported)
			property := astmodel.NewPropertyDefinition(propertyName, jsonName, concreteType).MakeOptional().WithDescription(propertyDescription)
			properties = append(properties, property)
		case *astmodel.ObjectType:
			// TODO: This name sucks but what alternative do we have?
			name := fmt.Sprintf("object%v", i)
			propertyName := synthesizer.idFactory.CreatePropertyName(name, astmodel.Exported)

			// JSON name is unimportant here because we will implement the JSON marshaller anyway,
			// but we still need it for controller-gen
			jsonName := synthesizer.idFactory.CreateIdentifier(name, astmodel.NotExported)
			property := astmodel.NewPropertyDefinition(propertyName, jsonName, concreteType).MakeOptional().WithDescription(propertyDescription)
			properties = append(properties, property)
		case *astmodel.PrimitiveType:
			var primitiveTypeName string
			if concreteType == astmodel.AnyType {
				primitiveTypeName = "anything"
			} else {
				primitiveTypeName = concreteType.Name()
			}

			// TODO: This name sucks but what alternative do we have?
			name := fmt.Sprintf("%v%v", primitiveTypeName, i)
			propertyName := synthesizer.idFactory.CreatePropertyName(name, astmodel.Exported)

			// JSON name is unimportant here because we will implement the JSON marshaller anyway,
			// but we still need it for controller-gen
			jsonName := synthesizer.idFactory.CreateIdentifier(name, astmodel.NotExported)
			property := astmodel.NewPropertyDefinition(propertyName, jsonName, concreteType).MakeOptional().WithDescription(propertyDescription)
			properties = append(properties, property)
		default:
			return nil, errors.Errorf("unexpected oneOf member, type: %T", t)
		}
	}

	objectType := astmodel.NewObjectType().WithProperties(properties...)
	objectType = objectType.WithFunction(astmodel.JSONMarshalFunctionName, astmodel.NewOneOfJSONMarshalFunction(objectType, synthesizer.idFactory))

	return objectType, nil
}

// converts a type that might be a typeName into something that isn't a typeName
func (synthesizer synthesizer) fullyResolve(t astmodel.Type) (astmodel.Type, error) {
	tName, ok := t.(astmodel.TypeName)
	for ok {
		tDef, found := synthesizer.defs[tName]
		if !found {
			return nil, errors.Errorf("couldn't find definition for %v", tName)
		}

		t = tDef.Type()
		tName, ok = t.(astmodel.TypeName)
	}

	return t, nil
}

// makes an ObjectType for an AllOf type
func (synthesizer synthesizer) allOfObject(allOf astmodel.AllOfType) (astmodel.Type, error) {

	props, err := synthesizer.extractAllOfProperties(allOf)
	if err != nil {
		return nil, err
	}

	return astmodel.NewObjectType().WithProperties(props...), nil

	// see if any of the inner ones are a oneOf:
	var oneOfs []astmodel.OneOfType
	var noneOneOfs []astmodel.Type
	for _, t := range allOf.Types() {
		tResolved, err := synthesizer.fullyResolve(t)
		if err != nil {
			return nil, err
		}

		if oneOf, ok := tResolved.(astmodel.OneOfType); ok {
			oneOf, err := synthesizer.flattenOneOf(oneOf)
			if err != nil {
				return nil, err
			}

			oneOfs = append(oneOfs, astmodel.MakeOneOfType(oneOf).(astmodel.OneOfType)) // guaranteed to be of correct type
		} else {
			// do _not_ use tResolved here as we want to preserve TypeNames that
			// do not point to a OneOf
			noneOneOfs = append(noneOneOfs, t)
		}
	}

	if len(oneOfs) > 1 {
		panic("cannot handle multiple oneOfs in allOf")
	}

	// If there's more than one option, synthesize a type.
	var properties []*astmodel.PropertyDefinition

	for _, t := range noneOneOfs {
		// unpack the contents of what we got from subhandlers:
		ps, err := synthesizer.extractAllOfProperties(t)
		if err != nil {
			return nil, err
		}

		properties = append(properties, ps...)
	}

	if len(oneOfs) == 0 {
		result := astmodel.NewObjectType().WithProperties(properties...)
		return result, nil
	}

	// if we have an allOf over a oneOf then we want to push
	// the allOf into the oneOf
	// so:
	// 	allOf { x, y, oneOf {a, b} }
	// becomes:
	//  oneOf { allOf {x, y, a}, allOf {x, y, b} }
	// the latter is much easier to model

	var newOneOfTypes []astmodel.Type
	for _, oneOfType := range oneOfs[0].Types() {
		// unpack the
		unpackedOneOf, err := synthesizer.extractAllOfProperties(oneOfType)
		if err != nil {
			return nil, err
		}

		newOneOfType := astmodel.NewObjectType().WithProperties(append(properties, unpackedOneOf...)...)
		newOneOfTypes = append(newOneOfTypes, newOneOfType)
	}

	return astmodel.MakeOneOfType(newOneOfTypes), nil
}

// flattenOneOf flattens a OneOf so that it does not contain any nested OneOfs,
// including via TypeName indirection
func (synthesizer synthesizer) flattenOneOf(oneOfType astmodel.OneOfType) ([]astmodel.Type, error) {
	var types []astmodel.Type

	for _, t := range oneOfType.Types() {
		tResolved, err := synthesizer.fullyResolve(t)
		if err != nil {
			return nil, err
		}

		if tOneOf, ok := tResolved.(astmodel.OneOfType); ok {
			types = append(types, tOneOf.Types()...)
		} else {
			// do _not_ use tResolved here as we want to preserve TypeNames
			// that do not point to OneOfs
			types = append(types, t)
		}
	}

	return types, nil
}

// extractAllOfProperties pulls out all the properties to be put into the destination object type
func (synthesizer synthesizer) extractAllOfProperties(st astmodel.Type) ([]*astmodel.PropertyDefinition, error) {
	switch concreteType := st.(type) {
	case *astmodel.ObjectType:
		// if it's an object type get all its properties:
		return concreteType.Properties(), nil

	case *astmodel.ResourceType:
		// it is a little strange to merge one resource into another with allOf,
		// but it is done and therefore we have to support it.
		// (an example is the Microsoft.VisualStudio Project type)
		// at the moment we will just take the spec type:
		return synthesizer.extractAllOfProperties(concreteType.SpecType())

	case astmodel.TypeName:
		if def, ok := synthesizer.defs[concreteType]; ok {
			return synthesizer.extractAllOfProperties(def.Type())
		}

		return nil, errors.Errorf("couldn't find definition for: %v", concreteType)

	case *astmodel.MapType:
		if concreteType.KeyType().Equals(astmodel.StringType) {
			// move map type into 'additionalProperties' property
			// TODO: consider privileging this as its own property on ObjectType,
			// since it has special behaviour and we need to handle it differently
			// for JSON serialization
			newProp := astmodel.NewPropertyDefinition(
				"additionalProperties",
				"additionalProperties",
				concreteType)

			return []*astmodel.PropertyDefinition{newProp}, nil
		}

	case astmodel.AllOfType:
		var properties []*astmodel.PropertyDefinition
		for _, t := range concreteType.Types() {
			props, err := synthesizer.extractAllOfProperties(t)
			if err != nil {
				return nil, err
			}

			properties = append(properties, props...)
		}

		return properties, nil
	}

	return nil, errors.Errorf("unable to handle type in allOf: %#v\n", st)
}
