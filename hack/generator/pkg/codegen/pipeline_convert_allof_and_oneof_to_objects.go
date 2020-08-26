/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

// convertAllOfAndOneOfToObjects reduces the AllOfType and OneOfType to ObjectType
func convertAllOfAndOneOfToObjects(idFactory astmodel.IdentifierFactory) PipelineStage {
	return MakePipelineStage(
		"allof-anyof-objects",
		"Convert allOf and oneOf to object types",
		func(ctx context.Context, defs astmodel.Types) (result astmodel.Types, err error) {
			visitor := astmodel.MakeTypeVisitor()

			// the context here is whether we are selecting spec or status fields

			//originalVisitAllOf := visitor.VisitAllOfType
			visitor.VisitAllOfType = func(this *astmodel.TypeVisitor, it astmodel.AllOfType, ctx interface{}) astmodel.Type {
				synth := synthesizer{
					resourceSelector: ctx.(resourceSelector),
					defs:             defs,
					idFactory:        idFactory,
				}

				object, err := synth.allOfObject(it)
				if err != nil {
					panic(err)
				}

				// we might end up with something that requires re-visiting
				return this.Visit(object, ctx)
			}

			originalVisitOneOf := visitor.VisitOneOfType
			visitor.VisitOneOfType = func(this *astmodel.TypeVisitor, it astmodel.OneOfType, ctx interface{}) astmodel.Type {
				// process children first
				result := originalVisitOneOf(this, it, ctx)

				if resultOneOf, ok := result.(astmodel.OneOfType); ok {
					synth := synthesizer{
						resourceSelector: ctx.(resourceSelector),
						defs:             defs,
						idFactory:        idFactory,
					}

					object, err := synth.oneOfObject(resultOneOf)
					if err != nil {
						panic(err)
					}

					result = object
				}

				return this.Visit(result, ctx)
			}

			var specSelector resourceSelector = func(r *astmodel.ResourceType) astmodel.Type {
				return r.SpecType()
			}

			var statusSelector resourceSelector = func(r *astmodel.ResourceType) astmodel.Type {
				return r.StatusType()
			}

			visitor.VisitResourceType = func(this *astmodel.TypeVisitor, it *astmodel.ResourceType, ctx interface{}) astmodel.Type {
				spec := this.Visit(it.SpecType(), specSelector)
				status := this.Visit(it.StatusType(), statusSelector)
				return it.WithSpec(spec).WithStatus(status)
			}

			var processing astmodel.TypeName

			// convert any panics we threw above back to errs
			// this sets the named error type in the result
			defer func() {
				if caught := recover(); caught != nil {
					if e, ok := caught.(error); ok {
						err = errors.Wrapf(e, "error while visiting %v", processing)
					} else {
						panic(caught)
					}
				}
			}()

			result = make(astmodel.Types)
			for _, def := range defs {
				resourceUpdater := specSelector
				// TODO: we need flags
				if strings.HasSuffix(def.Name().Name(), "_Status") {
					resourceUpdater = statusSelector
				}

				processing = def.Name()

				transformed := visitor.VisitDefinition(def, resourceUpdater)
				result.Add(transformed)
			}

			return // result, err named return types
		})
}

type resourceSelector func(*astmodel.ResourceType) astmodel.Type
type synthesizer struct {
	resourceSelector resourceSelector
	idFactory        astmodel.IdentifierFactory
	defs             astmodel.Types
}

func (s synthesizer) oneOfObject(oneOf astmodel.OneOfType) (astmodel.Type, error) {
	// If there's more than one option, synthesize a type.
	// Note that this is required because Kubernetes CRDs do not support OneOf the same way
	// OpenAPI does, see https://github.com/Azure/k8s-infra/issues/71
	var properties []*astmodel.PropertyDefinition

	propertyDescription := "mutually exclusive with all other properties"
	for i, t := range oneOf.Types() {
		prop, err := s.extractOneOfProperties(i, t)
		if err != nil {
			return nil, err
		}

		prop = prop.MakeOptional()
		prop = prop.WithDescription(propertyDescription)

		properties = append(properties, prop)
	}

	objectType := astmodel.NewObjectType().WithProperties(properties...)
	objectType = objectType.WithFunction(astmodel.JSONMarshalFunctionName, astmodel.NewOneOfJSONMarshalFunction(objectType, s.idFactory))

	return objectType, nil
}

func (s synthesizer) extractOneOfProperties(i int, from astmodel.Type) (*astmodel.PropertyDefinition, error) {
	switch concreteType := from.(type) {
	case astmodel.TypeName:
		propertyName := s.idFactory.CreatePropertyName(concreteType.Name(), astmodel.Exported)

		// JSON name is unimportant here because we will implement the JSON marshaller anyway,
		// but we still need it for controller-gen
		jsonName := s.idFactory.CreateIdentifier(concreteType.Name(), astmodel.NotExported)
		return astmodel.NewPropertyDefinition(propertyName, jsonName, concreteType), nil
	case *astmodel.EnumType:
		// TODO: This name sucks but what alternative do we have?
		name := fmt.Sprintf("enum%v", i)
		propertyName := s.idFactory.CreatePropertyName(name, astmodel.Exported)

		// JSON name is unimportant here because we will implement the JSON marshaller anyway,
		// but we still need it for controller-gen
		jsonName := s.idFactory.CreateIdentifier(name, astmodel.NotExported)
		return astmodel.NewPropertyDefinition(propertyName, jsonName, concreteType), nil
	case *astmodel.ObjectType:
		// TODO: This name sucks but what alternative do we have?
		name := fmt.Sprintf("object%v", i)
		propertyName := s.idFactory.CreatePropertyName(name, astmodel.Exported)

		// JSON name is unimportant here because we will implement the JSON marshaller anyway,
		// but we still need it for controller-gen
		jsonName := s.idFactory.CreateIdentifier(name, astmodel.NotExported)
		return astmodel.NewPropertyDefinition(propertyName, jsonName, concreteType), nil
	case *astmodel.PrimitiveType:
		var primitiveTypeName string
		if concreteType == astmodel.AnyType {
			primitiveTypeName = "anything"
		} else {
			primitiveTypeName = concreteType.Name()
		}

		// TODO: This name sucks but what alternative do we have?
		name := fmt.Sprintf("%v%v", primitiveTypeName, i)
		propertyName := s.idFactory.CreatePropertyName(name, astmodel.Exported)

		// JSON name is unimportant here because we will implement the JSON marshaller anyway,
		// but we still need it for controller-gen
		jsonName := s.idFactory.CreateIdentifier(name, astmodel.NotExported)
		return astmodel.NewPropertyDefinition(propertyName, jsonName, concreteType).MakeOptional(), nil

	case *astmodel.ResourceType:
		// a resource nested inside another type
		// need to see if we are processing status or spec types
		selectedType := s.resourceSelector(concreteType)
		return s.extractOneOfProperties(i, selectedType) // no resource prefix?

	default:
		return nil, errors.Errorf("unexpected oneOf member, type: %T", from)
	}

}

func (s synthesizer) intersectTypes(left astmodel.Type, right astmodel.Type) (astmodel.Type, error) {
	for _, handler := range intersectHandlers {
		result, err := handler(s, left, right)
		if err != nil {
			return nil, err
		}

		if result != nil {
			return result, nil
		}
	}

	return nil, errors.Errorf("don't know how to intersect types: %s and %s", left, right)
}

// intersectHandler knows how to do intersection for one case only
// it is biased to examine the LHS type (but some are symmetric and handle both sides at once)
type intersectHandler = func(synthesizer, astmodel.Type, astmodel.Type) (astmodel.Type, error)

// handles the same case for the right hand side
func flip(f intersectHandler) intersectHandler {
	return func(s synthesizer, left astmodel.Type, right astmodel.Type) (astmodel.Type, error) {
		return f(s, right, left)
	}
}

var intersectHandlers []intersectHandler

func init() {
	intersectHandlers = []intersectHandler{
		synthesizer.handleEqualTypes, // equality is symmetric
		synthesizer.handleAnyType, flip(synthesizer.handleAnyType),
		synthesizer.handleAllOfType, flip(synthesizer.handleAllOfType),
		synthesizer.handleTypeName, flip(synthesizer.handleTypeName),
		synthesizer.handleOneOf, flip(synthesizer.handleOneOf),
		synthesizer.handleOptionalOptional,                           // symmetric
		synthesizer.handleOptional, flip(synthesizer.handleOptional), // needs to be before Enum
		synthesizer.handleEnumEnum, // symmetric
		synthesizer.handleEnum, flip(synthesizer.handleEnum),
		synthesizer.handleObjectObject, // symmetric
		synthesizer.handleMapMap,       // symmetric
		synthesizer.handleMapObject, flip(synthesizer.handleMapObject),
		synthesizer.handleResourceType, flip(synthesizer.handleResourceType),
	}
}

func (s synthesizer) handleOptional(left astmodel.Type, right astmodel.Type) (astmodel.Type, error) {
	if leftOptional, ok := left.(*astmodel.OptionalType); ok {
		// is this wrong? it feels wrong, but needed for {optional{enum}, string}
		return s.intersectTypes(leftOptional.Element(), right)
	}

	return nil, nil
}

func (s synthesizer) handleResourceType(left astmodel.Type, right astmodel.Type) (astmodel.Type, error) {
	if leftResource, ok := left.(*astmodel.ResourceType); ok {
		if _, ok := right.(*astmodel.ResourceType); ok {
			return nil, errors.Errorf("cannot combine two resource types") // safety check
		}

		// resourceSelector picks the right spec/status for us
		return s.intersectTypes(s.resourceSelector(leftResource), right)
	}

	return nil, nil
}

func (s synthesizer) handleOptionalOptional(left astmodel.Type, right astmodel.Type) (astmodel.Type, error) {
	// if both optional merge their contents and put back in an optional
	if leftOptional, ok := left.(*astmodel.OptionalType); ok {
		if rightOptional, ok := right.(*astmodel.OptionalType); ok {
			result, err := s.intersectTypes(leftOptional.Element(), rightOptional.Element())
			if err != nil {
				return nil, err
			}

			return astmodel.NewOptionalType(result), nil
		}
	}

	return nil, nil
}

func (s synthesizer) handleMapMap(left astmodel.Type, right astmodel.Type) (astmodel.Type, error) {
	if leftMap, ok := left.(*astmodel.MapType); ok {
		if rightMap, ok := right.(*astmodel.MapType); ok {
			keyType, err := s.intersectTypes(leftMap.KeyType(), rightMap.KeyType())
			if err != nil {
				return nil, err
			}

			valueType, err := s.intersectTypes(leftMap.ValueType(), rightMap.ValueType())
			if err != nil {
				return nil, err
			}

			return astmodel.NewMapType(keyType, valueType), nil
		}
	}

	return nil, nil
}

func (s synthesizer) handleObjectObject(left astmodel.Type, right astmodel.Type) (astmodel.Type, error) {
	if leftObj, ok := left.(*astmodel.ObjectType); ok {
		if rightObj, ok := right.(*astmodel.ObjectType); ok {
			mergedProps := make(map[astmodel.PropertyName]*astmodel.PropertyDefinition)

			for _, p := range leftObj.Properties() {
				mergedProps[p.PropertyName()] = p
			}

			for _, p := range rightObj.Properties() {
				if existingProp, ok := mergedProps[p.PropertyName()]; ok {
					newType, err := s.intersectTypes(existingProp.PropertyType(), p.PropertyType())
					if err != nil {
						klog.Errorf("unable to combine properties: %s (%v)", p.PropertyName(), err)
						continue
						//return nil, err
					}

					// TODO: need to handle merging requiredness and tags and...
					mergedProps[p.PropertyName()] = existingProp.WithType(newType)
				} else {
					mergedProps[p.PropertyName()] = p
				}
			}

			// flatten
			var properties []*astmodel.PropertyDefinition
			for _, p := range mergedProps {
				properties = append(properties, p)
			}

			// TODO: need to handle merging other bits of objects
			return leftObj.WithProperties(properties...), nil
		}
	}

	return nil, nil
}

func (s synthesizer) handleEnumEnum(left astmodel.Type, right astmodel.Type) (astmodel.Type, error) {
	if leftEnum, ok := left.(*astmodel.EnumType); ok {
		// if both are enums then the elements must be common
		if rightEnum, ok := right.(*astmodel.EnumType); ok {
			if !leftEnum.BaseType().Equals(rightEnum.BaseType()) {
				return nil, errors.Errorf("cannot merge enums with differing base types")
			}

			var inBoth []astmodel.EnumValue

			for _, option := range leftEnum.Options() {
				for _, otherOption := range rightEnum.Options() {
					if option == otherOption {
						inBoth = append(inBoth, option)
						break
					}
				}
			}

			return astmodel.NewEnumType(leftEnum.BaseType(), inBoth), nil
		}
	}

	return nil, nil
}

func (s synthesizer) handleEnum(left astmodel.Type, right astmodel.Type) (astmodel.Type, error) {
	if leftEnum, ok := left.(*astmodel.EnumType); ok {
		// we can restrict from a (maybe optional) base type to an enum type
		if leftEnum.BaseType().Equals(right) ||
			astmodel.NewOptionalType(leftEnum.BaseType()).Equals(right) {
			return leftEnum, nil
		}

		var strs []string
		for _, enumValue := range leftEnum.Options() {
			strs = append(strs, enumValue.String())
		}

		return nil, errors.Errorf("don't know how to merge enum type (%s) with %s", strings.Join(strs, ", "), right)
	}

	return nil, nil
}

func (s synthesizer) handleAllOfType(left astmodel.Type, right astmodel.Type) (astmodel.Type, error) {
	if leftAllOf, ok := left.(astmodel.AllOfType); ok {
		result, err := s.allOfObject(leftAllOf)
		if err != nil {
			return nil, err
		}

		return s.intersectTypes(result, right)
	}

	return nil, nil
}

// if combining a type with a oneOf that contains that type, the result is that type
func (s synthesizer) handleOneOf(left astmodel.Type, right astmodel.Type) (astmodel.Type, error) {
	if leftOneOf, ok := left.(astmodel.OneOfType); ok {
		// if there is an equal case, use that:
		for _, lType := range leftOneOf.Types() {
			if lType.Equals(right) {
				return lType, nil
			}
		}

		// otherwise intersect with each type:
		var newTypes []astmodel.Type
		for _, lType := range leftOneOf.Types() {
			newType, err := s.intersectTypes(lType, right)
			if err != nil {
				return nil, err
			}

			newTypes = append(newTypes, newType)
		}

		return astmodel.MakeOneOfType(newTypes), nil
	}

	return nil, nil
}

func (s synthesizer) handleTypeName(left astmodel.Type, right astmodel.Type) (astmodel.Type, error) {
	if leftName, ok := left.(astmodel.TypeName); ok {
		if found, ok := s.defs[leftName]; !ok {
			panic(fmt.Sprintf("couldn't find type %s", leftName))
		} else {
			result, err := s.intersectTypes(found.Type(), right)
			if err != nil {
				return nil, err
			}

			// TODO: can we somehow process these pointed-to types first,
			// so that their innards are resolved already and we can retain
			// the names?

			if result.Equals(found.Type()) {
				// if we got back the same thing we are referencing, preserve the reference
				return leftName, nil
			}

			return result, nil
		}
	}

	return nil, nil
}

// any type always disappears when intersected with another type
func (synthesizer) handleAnyType(left astmodel.Type, right astmodel.Type) (astmodel.Type, error) {
	if left.Equals(astmodel.AnyType) {
		return right, nil
	}

	return nil, nil
}

// two identical types can become the same type
func (synthesizer) handleEqualTypes(left astmodel.Type, right astmodel.Type) (astmodel.Type, error) {
	if left.Equals(right) {
		return left, nil
	}

	return nil, nil
}

// a string map and object can be combined with the map type becoming additionalProperties
func (synthesizer) handleMapObject(left astmodel.Type, right astmodel.Type) (astmodel.Type, error) {
	if leftMap, ok := left.(*astmodel.MapType); ok {
		if leftMap.KeyType().Equals(astmodel.StringType) {
			if rightObj, ok := right.(*astmodel.ObjectType); ok {
				if len(rightObj.Properties()) == 0 {
					// no properties, treat as map
					// TODO: there could be other things in the object to check?
					return leftMap, nil
				}

				additionalProps := astmodel.NewPropertyDefinition("additionalProperties", "additionalProperties", leftMap)
				return rightObj.WithProperties(additionalProps), nil
			}
		}
	}

	return nil, nil
}

// makes an ObjectType for an AllOf type
func (s synthesizer) allOfObject(allOf astmodel.AllOfType) (astmodel.Type, error) {

	var intersection astmodel.Type = astmodel.AnyType
	for _, t := range allOf.Types() {
		var err error
		intersection, err = s.intersectTypes(intersection, t)
		if err != nil {
			return nil, err
		}
	}

	return intersection, nil
}
