/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"

	"github.com/pkg/errors"

	kerrors "k8s.io/apimachinery/pkg/util/errors"
)

// TypeVisitor represents a visitor for a tree of types.
// The `ctx` argument can be used to “smuggle” additional data down the call-chain.
type TypeVisitor struct {
	VisitTypeName      func(this *TypeVisitor, it TypeName, ctx interface{}) (Type, error)
	VisitOneOfType     func(this *TypeVisitor, it *OneOfType, ctx interface{}) (Type, error)
	VisitAllOfType     func(this *TypeVisitor, it *AllOfType, ctx interface{}) (Type, error)
	VisitArrayType     func(this *TypeVisitor, it *ArrayType, ctx interface{}) (Type, error)
	VisitPrimitive     func(this *TypeVisitor, it *PrimitiveType, ctx interface{}) (Type, error)
	VisitObjectType    func(this *TypeVisitor, it *ObjectType, ctx interface{}) (Type, error)
	VisitMapType       func(this *TypeVisitor, it *MapType, ctx interface{}) (Type, error)
	VisitOptionalType  func(this *TypeVisitor, it *OptionalType, ctx interface{}) (Type, error)
	VisitEnumType      func(this *TypeVisitor, it *EnumType, ctx interface{}) (Type, error)
	VisitResourceType  func(this *TypeVisitor, it *ResourceType, ctx interface{}) (Type, error)
	VisitFlaggedType   func(this *TypeVisitor, it *FlaggedType, ctx interface{}) (Type, error)
	VisitValidatedType func(this *TypeVisitor, it *ValidatedType, ctx interface{}) (Type, error)
	VisitErroredType   func(this *TypeVisitor, it *ErroredType, ctx interface{}) (Type, error)
}

// Visit invokes the appropriate VisitX on TypeVisitor
func (tv *TypeVisitor) Visit(t Type, ctx interface{}) (Type, error) {
	if t == nil {
		return nil, nil
	}

	switch it := t.(type) {
	case TypeName:
		return tv.VisitTypeName(tv, it, ctx)
	case *OneOfType:
		return tv.VisitOneOfType(tv, it, ctx)
	case *AllOfType:
		return tv.VisitAllOfType(tv, it, ctx)
	case *ArrayType:
		return tv.VisitArrayType(tv, it, ctx)
	case *PrimitiveType:
		return tv.VisitPrimitive(tv, it, ctx)
	case *ObjectType:
		return tv.VisitObjectType(tv, it, ctx)
	case *MapType:
		return tv.VisitMapType(tv, it, ctx)
	case *OptionalType:
		return tv.VisitOptionalType(tv, it, ctx)
	case *EnumType:
		return tv.VisitEnumType(tv, it, ctx)
	case *ResourceType:
		return tv.VisitResourceType(tv, it, ctx)
	case *FlaggedType:
		return tv.VisitFlaggedType(tv, it, ctx)
	case *ValidatedType:
		return tv.VisitValidatedType(tv, it, ctx)
	case *ErroredType:
		return tv.VisitErroredType(tv, it, ctx)
	}

	panic(fmt.Sprintf("unhandled type: (%T) %v", t, t))
}

// VisitDefinition invokes the TypeVisitor on both the name and type of the definition
// NB: this is only valid if VisitTypeName returns a TypeName and not generally a Type
func (tv *TypeVisitor) VisitDefinition(td TypeDefinition, ctx interface{}) (TypeDefinition, error) {
	visitedName, err := tv.VisitTypeName(tv, td.Name(), ctx)
	if err != nil {
		return TypeDefinition{}, errors.Wrapf(err, "visit of %q failed", td.Name())
	}

	name, ok := visitedName.(TypeName)
	if !ok {
		return TypeDefinition{}, errors.Errorf("expected visit of %q to return TypeName, not %T", td.Name(), visitedName)
	}

	visitedType, err := tv.Visit(td.Type(), ctx)
	if err != nil {
		return TypeDefinition{}, errors.Wrapf(err, "visit of type of %q failed", td.Name())
	}

	def := td.WithName(name).WithType(visitedType)
	return def, nil
}

func (tv *TypeVisitor) VisitDefinitions(definitions Types, ctx interface{}) (Types, error) {
	result := make(Types)
	var errs []error
	for _, d := range definitions {
		def, err := tv.VisitDefinition(d, ctx)
		if err != nil {
			errs = append(errs, err)
		} else {
			result[def.Name()] = def
		}
	}

	if len(errs) > 0 {
		err := kerrors.NewAggregate(errs)
		return nil, err
	}

	return result, nil
}

// MakeTypeVisitor returns visitor to visit every type in a tree, configured with the functions
// passed. Further customization can be done by overriding the fields in the returned instance.
//
// visitations is a sequence of functions to be wired into the resulting type visitor, each
// providing a specific desired transformation.
//
// The treatment of each function depends on its signature, as follows:
//
// func(*TypeVisitor, <specificType>, interface{}) (Type, error)
//
//   o  Will be wired in as a handler for <specificType>, replacing any existing visitor
//   o  If you want to use the *TypeVisitor to visit a nested type, you must do this yourself
//
// func (<specificType>) (Type, error)
//
//   o  Will be wired in as a handler for <specificType>, replacing any existing visitor
//   o  The result of the function will be returned directly, it won't be visited automatically
//
// All assignments overwrite any existing field values, so if you have multiple functions matching
// the same field, only the last one will be retained.
//
func MakeTypeVisitor(visitations ...interface{}) TypeVisitor {
	// TODO [performance]: we can do reference comparisons on the results of
	// recursive invocations of Visit to avoid having to rebuild the tree if the
	// leaf nodes do not actually change.
	result := TypeVisitor{
		VisitTypeName:      IdentityVisitOfTypeName,
		VisitArrayType:     IdentityVisitOfArrayType,
		VisitPrimitive:     IdentityVisitOfPrimitiveType,
		VisitObjectType:    IdentityVisitOfObjectType,
		VisitMapType:       IdentityVisitOfMapType,
		VisitEnumType:      IdentityVisitOfEnumType,
		VisitOptionalType:  IdentityVisitOfOptionalType,
		VisitResourceType:  IdentityVisitOfResourceType,
		VisitOneOfType:     IdentityVisitOfOneOfType,
		VisitAllOfType:     IdentityVisitOfAllOfType,
		VisitFlaggedType:   IdentityVisitOfFlaggedType,
		VisitValidatedType: IdentityVisitOfValidatedType,
		VisitErroredType:   IdentityVisitOfErroredType,
	}

	for _, visitation := range visitations {
		switch v := visitation.(type) {
		// TypeName
		case func(*TypeVisitor, TypeName, interface{}) (Type, error):
			result.VisitTypeName = v
		case func(TypeName) (Type, error):
			result.VisitTypeName = func(_ *TypeVisitor, it TypeName, _ interface{}) (Type, error) {
				return v(it)
			}
		// OneOfType
		case func(*TypeVisitor, *OneOfType, interface{}) (Type, error):
			result.VisitOneOfType = v
		case func(*OneOfType) (Type, error):
			result.VisitOneOfType = func(_ *TypeVisitor, it *OneOfType, _ interface{}) (Type, error) {
				return v(it)
			}
		// AllOfType
		case func(*TypeVisitor, *AllOfType, interface{}) (Type, error):
			result.VisitAllOfType = v
		case func(*AllOfType) (Type, error):
			result.VisitAllOfType = func(_ *TypeVisitor, it *AllOfType, _ interface{}) (Type, error) {
				return v(it)
			}
		// ArrayType
		case func(*TypeVisitor, *ArrayType, interface{}) (Type, error):
			result.VisitArrayType = v
		case func(*ArrayType) (Type, error):
			result.VisitArrayType = func(_ *TypeVisitor, it *ArrayType, _ interface{}) (Type, error) {
				return v(it)
			}
		// PrimitiveType
		case func(*TypeVisitor, *PrimitiveType, interface{}) (Type, error):
			result.VisitPrimitive = v
		case func(*PrimitiveType) (Type, error):
			result.VisitPrimitive = func(_ *TypeVisitor, it *PrimitiveType, _ interface{}) (Type, error) {
				return v(it)
			}
		// ObjectType
		case func(*TypeVisitor, *ObjectType, interface{}) (Type, error):
			result.VisitObjectType = v
		case func(*ObjectType) (Type, error):
			result.VisitObjectType = func(_ *TypeVisitor, it *ObjectType, _ interface{}) (Type, error) {
				return v(it)
			}
		// MapType
		case func(*TypeVisitor, *MapType, interface{}) (Type, error):
			result.VisitMapType = v
		case func(*MapType) (Type, error):
			result.VisitMapType = func(_ *TypeVisitor, it *MapType, _ interface{}) (Type, error) {
				return v(it)
			}
		// OptionalType
		case func(*TypeVisitor, *OptionalType, interface{}) (Type, error):
			result.VisitOptionalType = v
		case func(*OptionalType) (Type, error):
			result.VisitOptionalType = func(_ *TypeVisitor, it *OptionalType, _ interface{}) (Type, error) {
				return v(it)
			}
		// EnumType
		case func(*TypeVisitor, *EnumType, interface{}) (Type, error):
			result.VisitEnumType = v
		case func(*EnumType) (Type, error):
			result.VisitEnumType = func(_ *TypeVisitor, it *EnumType, _ interface{}) (Type, error) {
				return v(it)
			}
		// ResourceType
		case func(*TypeVisitor, *ResourceType, interface{}) (Type, error):
			result.VisitResourceType = v
		case func(*ResourceType) (Type, error):
			result.VisitResourceType = func(_ *TypeVisitor, it *ResourceType, _ interface{}) (Type, error) {
				return v(it)
			}
		// FlaggedType
		case func(*TypeVisitor, *FlaggedType, interface{}) (Type, error):
			result.VisitFlaggedType = v
		case func(*FlaggedType) (Type, error):
			result.VisitFlaggedType = func(_ *TypeVisitor, it *FlaggedType, _ interface{}) (Type, error) {
				return v(it)
			}
		// ErroredType
		case func(*TypeVisitor, *ErroredType, interface{}) (Type, error):
			result.VisitErroredType = v
		case func(*ErroredType) (Type, error):
			result.VisitErroredType = func(_ *TypeVisitor, it *ErroredType, _ interface{}) (Type, error) {
				return v(it)
			}
		// ValidatedType
		case func(*TypeVisitor, *ValidatedType, interface{}) (Type, error):
			result.VisitValidatedType = v
		case func(*ValidatedType) (Type, error):
			result.VisitValidatedType = func(_ *TypeVisitor, it *ValidatedType, _ interface{}) (Type, error) {
				return v(it)
			}
		default:
			panic(fmt.Sprintf("unexpected visitation %#v", v))
		}
	}

	return result
}

func IdentityVisitOfTypeName(_ *TypeVisitor, it TypeName, _ interface{}) (Type, error) {
	return it, nil
}

func IdentityVisitOfArrayType(this *TypeVisitor, it *ArrayType, ctx interface{}) (Type, error) {
	newElement, err := this.Visit(it.element, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to visit type of array")
	}

	if newElement == it.element {
		return it, nil // short-circuit
	}

	return NewArrayType(newElement), nil
}

func IdentityVisitOfPrimitiveType(_ *TypeVisitor, it *PrimitiveType, _ interface{}) (Type, error) {
	return it, nil
}

func IdentityVisitOfObjectType(this *TypeVisitor, it *ObjectType, ctx interface{}) (Type, error) {
	// just map the property types
	var errs []error
	var newProps []*PropertyDefinition
	for _, prop := range it.Properties() {
		p, err := this.Visit(prop.propertyType, ctx)
		if err != nil {
			errs = append(errs, err)
		} else {
			newProps = append(newProps, prop.WithType(p))
		}
	}

	if len(errs) > 0 {
		return nil, kerrors.NewAggregate(errs)
	}

	// map the embedded types too
	var newEmbeddedProps []*PropertyDefinition
	for _, prop := range it.EmbeddedProperties() {
		p, err := this.Visit(prop.propertyType, ctx)
		if err != nil {
			errs = append(errs, err)
		} else {
			newEmbeddedProps = append(newEmbeddedProps, prop.WithType(p))
		}
	}

	if len(errs) > 0 {
		return nil, kerrors.NewAggregate(errs)
	}

	result := it.WithProperties(newProps...)
	// Since it's possible that the type was renamed we need to clear the old embedded properties
	result = result.WithoutEmbeddedProperties()
	result, err := result.WithEmbeddedProperties(newEmbeddedProps...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func IdentityVisitOfMapType(this *TypeVisitor, it *MapType, ctx interface{}) (Type, error) {
	visitedKey, err := this.Visit(it.key, ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to visit map key type %T", it.key)
	}

	visitedValue, err := this.Visit(it.value, ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to visit map value type %T", it.value)
	}

	if visitedKey == it.key && visitedValue == it.value {
		return it, nil // short-circuit
	}

	return NewMapType(visitedKey, visitedValue), nil
}

func IdentityVisitOfEnumType(_ *TypeVisitor, it *EnumType, _ interface{}) (Type, error) {
	// if we visit the enum base type then we will also have to do something
	// about the values. so by default don't do anything with the enum base
	return it, nil
}

func IdentityVisitOfOptionalType(this *TypeVisitor, it *OptionalType, ctx interface{}) (Type, error) {
	visitedElement, err := this.Visit(it.element, ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to visit optional element type %T", it.element)
	}

	if visitedElement == it.element {
		return it, nil // short-circuit
	}

	return NewOptionalType(visitedElement), nil
}

func IdentityVisitOfResourceType(this *TypeVisitor, it *ResourceType, ctx interface{}) (Type, error) {
	visitedSpec, err := this.Visit(it.spec, ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to visit resource spec type %T", it.spec)
	}

	visitedStatus, err := this.Visit(it.status, ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to visit resource status type %T", it.status)
	}

	if visitedSpec == it.spec && visitedStatus == it.status {
		return it, nil // short-circuit
	}

	return it.WithSpec(visitedSpec).WithStatus(visitedStatus), nil
}

func IdentityVisitOfOneOfType(this *TypeVisitor, it *OneOfType, ctx interface{}) (Type, error) {
	var newTypes []Type
	err := it.Types().ForEachError(func(oneOf Type, _ int) error {
		newType, err := this.Visit(oneOf, ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to visit oneOf")
		}

		newTypes = append(newTypes, newType)
		return nil
	})

	if err != nil {
		return nil, err
	}

	if typeSlicesFastEqual(newTypes, it.types.types) {
		return it, nil // short-circuit
	}

	return BuildOneOfType(newTypes...), nil
}

func IdentityVisitOfAllOfType(this *TypeVisitor, it *AllOfType, ctx interface{}) (Type, error) {
	var newTypes []Type
	err := it.Types().ForEachError(func(allOf Type, _ int) error {
		newType, err := this.Visit(allOf, ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to visit allOf")
		}

		newTypes = append(newTypes, newType)
		return nil
	})

	if err != nil {
		return nil, err
	}

	if typeSlicesFastEqual(newTypes, it.types.types) {
		return it, nil // short-circuit
	}

	return BuildAllOfType(newTypes...), nil
}

// just checks reference equality of Types
// this is used to short-circuit when we don't need to make new types,
// in a fast manner than invoking Equals and walking the whole tree
func typeSlicesFastEqual(t1 []Type, t2 []Type) bool {
	if len(t1) != len(t2) {
		return false
	}

	for ix := range t1 {
		if t1[ix] != t2[ix] {
			return false
		}
	}

	return true
}

func IdentityVisitOfFlaggedType(this *TypeVisitor, ft *FlaggedType, ctx interface{}) (Type, error) {
	nt, err := this.Visit(ft.element, ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to visit flagged type %T", ft.element)
	}

	if nt == ft.element {
		return ft, nil // short-circuit
	}

	return ft.WithElement(nt), nil
}

func IdentityVisitOfValidatedType(this *TypeVisitor, v *ValidatedType, ctx interface{}) (Type, error) {
	nt, err := this.Visit(v.element, ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to visit validated type %T", v.element)
	}

	if nt == v.element {
		return v, nil // short-circuit
	}

	return v.WithType(nt), nil
}

func IdentityVisitOfErroredType(this *TypeVisitor, e *ErroredType, ctx interface{}) (Type, error) {
	nt, err := this.Visit(e.inner, ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to visit errored type %T", e.inner)
	}

	if nt == e.inner {
		return e, nil // short-circuit
	}

	return e.WithType(nt), nil
}
