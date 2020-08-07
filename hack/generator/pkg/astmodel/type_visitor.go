package astmodel

import (
	"fmt"
	"k8s.io/klog"
)

// TypeVisitor represents a visitor for a tree of types.
// The `ctx` argument can be used to “smuggle” additional data down the call-chain.
type TypeVisitor struct {
	VisitTypeName     func(this *TypeVisitor, it TypeName, ctx interface{}) Type
	VisitArrayType    func(this *TypeVisitor, it *ArrayType, ctx interface{}) Type
	VisitPrimitive    func(this *TypeVisitor, it *PrimitiveType, ctx interface{}) Type
	VisitObjectType   func(this *TypeVisitor, it *ObjectType, ctx interface{}) Type
	VisitMapType      func(this *TypeVisitor, it *MapType, ctx interface{}) Type
	VisitOptionalType func(this *TypeVisitor, it *OptionalType, ctx interface{}) Type
	VisitEnumType     func(this *TypeVisitor, it *EnumType, ctx interface{}) Type
	VisitResourceType func(this *TypeVisitor, it *ResourceType, ctx interface{}) Type
}

// Visit invokes the appropriate VisitX on TypeVisitor
func (tv *TypeVisitor) Visit(t Type, ctx interface{}) Type {
	if t == nil {
		return nil
	}

	switch it := t.(type) {
	case TypeName:
		return tv.VisitTypeName(tv, it, ctx)
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
	}

	panic(fmt.Sprintf("unhandled type: (%T) %v", t, t))
}

// VisitAll applies the visitor to each of the types provided, returning a new set of types
func (tv *TypeVisitor) VisitAll(types Types, ctx interface{}) Types {
	result := make(Types)
	for _, d := range types {
		// Allow for rename of type
		n, ok := tv.Visit(d.Name(), ctx).(TypeName)
		if !ok {
			// If we get anything other than a TypeName back, we can't handle it -
			// falling back to the original name of the type is more useful than a panic
			klog.V(4).Infof("Ignoring transformation of TypeName %q to %v", d.Name(), n)
			n = d.Name()
		}
		t := tv.Visit(d.Type(), ctx)
		def := MakeTypeDefinition(n, t).WithDescription(d.description)
		result[def.Name()] = def
	}

	return result
}

// MakeTypeVisitor returns a default (identity transform) visitor, which
// visits every type in the tree. If you want to actually do something you will
// need to override the properties on the returned TypeVisitor.
func MakeTypeVisitor() TypeVisitor {
	// TODO [performance]: we can do reference comparisons on the results of
	// recursive invocations of Visit to avoid having to rebuild the tree if the
	// leaf nodes do not actually change.
	return TypeVisitor{
		VisitTypeName: func(_ *TypeVisitor, it TypeName, _ interface{}) Type {
			return it
		},
		VisitArrayType: func(this *TypeVisitor, it *ArrayType, ctx interface{}) Type {
			newElement := this.Visit(it.element, ctx)
			return NewArrayType(newElement)
		},
		VisitPrimitive: func(_ *TypeVisitor, it *PrimitiveType, _ interface{}) Type {
			return it
		},
		VisitObjectType: func(this *TypeVisitor, it *ObjectType, ctx interface{}) Type {
			// just map the property types
			var newProps []*PropertyDefinition
			for _, prop := range it.properties {
				newProps = append(newProps, prop.WithType(this.Visit(prop.propertyType, ctx)))
			}
			return it.WithProperties(newProps...)
		},
		VisitMapType: func(this *TypeVisitor, it *MapType, ctx interface{}) Type {
			newKey := this.Visit(it.key, ctx)
			newValue := this.Visit(it.value, ctx)
			return NewMapType(newKey, newValue)
		},
		VisitEnumType: func(_ *TypeVisitor, it *EnumType, _ interface{}) Type {
			// if we visit the enum base type then we will also have to do something
			// about the values. so by default don't do anything with the enum base
			return it
		},
		VisitOptionalType: func(this *TypeVisitor, it *OptionalType, ctx interface{}) Type {
			return NewOptionalType(this.Visit(it.element, ctx))
		},
		VisitResourceType: func(this *TypeVisitor, it *ResourceType, ctx interface{}) Type {
			spec := this.Visit(it.spec, ctx)
			status := this.Visit(it.status, ctx)
			return NewResourceType(spec, status)
		},
	}
}
