/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"testing"

	. "github.com/onsi/gomega"
)

const group = "grp"
const version = "v1"

var rootTypeName = MakeTypeName(makeTestLocalPackageReference(group, version), "Root")
var leftTypeName = MakeTypeName(makeTestLocalPackageReference(group, version), "Left")
var rightTypeName = MakeTypeName(makeTestLocalPackageReference(group, version), "Right")

func makeSimpleTestTypeGraph() Types {
	result := make(Types)

	leftChildType := NewObjectType().WithProperty(
		NewPropertyDefinition("SimpleString", "simpleString", StringType))
	leftChildDef := MakeTypeDefinition(
		leftTypeName,
		leftChildType)
	rightChildType := NewObjectType().WithProperty(
		NewPropertyDefinition("SimpleInt", "simpleInt", IntType))
	rightChildDef := MakeTypeDefinition(
		rightTypeName,
		rightChildType)

	rootType := NewObjectType().WithProperty(
		NewPropertyDefinition("Left", "left", leftChildDef.Name())).WithProperty(
		NewPropertyDefinition("Right", "right", rightChildDef.Name()))
	rootDef := MakeTypeDefinition(
		rootTypeName,
		rootType)

	result.Add(leftChildDef)
	result.Add(rightChildDef)
	result.Add(rootDef)

	return result
}

func makeDuplicateReferencesTypeGraph() Types {
	result := make(Types)

	childType := NewObjectType().WithProperty(
		NewPropertyDefinition("SimpleString", "simpleString", StringType))
	childDef := MakeTypeDefinition(
		leftTypeName,
		childType)

	rootType := NewObjectType().WithProperty(
		NewPropertyDefinition("Left", "left", childDef.Name())).WithProperty(
		NewPropertyDefinition("Right", "right", childDef.Name()))
	rootDef := MakeTypeDefinition(
		rootTypeName,
		rootType)

	result.Add(childDef)
	result.Add(rootDef)

	return result
}

func makeCycleTypeGraph() Types {
	result := make(Types)

	leftChildType := NewObjectType().WithProperty(
		NewPropertyDefinition("Root", "root", rootTypeName))
	leftChildDef := MakeTypeDefinition(
		leftTypeName,
		leftChildType)
	rightChildType := NewObjectType().WithProperty(
		NewPropertyDefinition("SimpleInt", "simpleInt", IntType))
	rightChildDef := MakeTypeDefinition(
		rightTypeName,
		rightChildType)

	rootType := NewObjectType().WithProperty(
		NewPropertyDefinition("Left", "left", leftChildDef.Name())).WithProperty(
		NewPropertyDefinition("Right", "right", rightChildDef.Name()))
	rootDef := MakeTypeDefinition(
		rootTypeName,
		rootType)

	result.Add(leftChildDef)
	result.Add(rightChildDef)
	result.Add(rootDef)

	return result
}

func TestTypeWalker_IdentityWalkReturnsIdenticalTypes(t *testing.T) {
	g := NewGomegaWithT(t)

	types := makeSimpleTestTypeGraph()
	visitor := MakeTypeVisitor()
	walker := NewTypeWalker(types, visitor)

	var walked []TypeDefinition

	walker.WalkFunc = func(this *TypeWalker, def TypeDefinition, ctx interface{}) (TypeDefinition, error) {
		walked = append(walked, def)
		return IdentityWalk(this, def, ctx)
	}

	rootDef := types[rootTypeName]
	updatedTypes, err := walker.Walk(rootDef, nil)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(len(walked)).To(Equal(3))
	g.Expect(len(updatedTypes)).To(Equal(3))

	g.Expect(walked[0].Name().Name()).To(Equal("Root"))
	g.Expect(walked[1].Name().Name()).To(Equal("Left"))
	g.Expect(walked[2].Name().Name()).To(Equal("Right"))

	for _, updated := range updatedTypes {
		g.Expect(updated.Type().Equals(types[updated.Name()].Type())).To(BeTrue())
	}
}

func TestTypeWalker_DuplicateTypesAreWalkedOnceEach_ReturnedOnce(t *testing.T) {
	g := NewGomegaWithT(t)

	types := makeDuplicateReferencesTypeGraph()
	visitor := MakeTypeVisitor()
	walker := NewTypeWalker(types, visitor)

	var walked []TypeDefinition

	walker.WalkFunc = func(this *TypeWalker, def TypeDefinition, ctx interface{}) (TypeDefinition, error) {
		walked = append(walked, def)
		return IdentityWalk(this, def, ctx)
	}

	rootDef := types[rootTypeName]
	updatedTypes, err := walker.Walk(rootDef, nil)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(len(walked)).To(Equal(3))
	g.Expect(len(updatedTypes)).To(Equal(2))

	g.Expect(walked[0].Name().Name()).To(Equal("Root"))
	g.Expect(walked[1].Name().Name()).To(Equal("Left"))

	for _, updated := range updatedTypes {
		g.Expect(updated.Type().Equals(types[updated.Name()].Type())).To(BeTrue())
	}
}

func TestTypeWalker_CyclesAllowed(t *testing.T) {
	g := NewGomegaWithT(t)

	types := makeCycleTypeGraph()
	visitor := MakeTypeVisitor()
	walker := NewTypeWalker(types, visitor)

	var walked []TypeDefinition

	walker.WalkFunc = func(this *TypeWalker, def TypeDefinition, ctx interface{}) (TypeDefinition, error) {
		walked = append(walked, def)
		return IdentityWalk(this, def, ctx)
	}

	rootDef := types[rootTypeName]
	updatedTypes, err := walker.Walk(rootDef, nil)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(len(walked)).To(Equal(3))
	g.Expect(len(updatedTypes)).To(Equal(3))

	g.Expect(walked[0].Name().Name()).To(Equal("Root"))
	g.Expect(walked[1].Name().Name()).To(Equal("Left"))
	g.Expect(walked[2].Name().Name()).To(Equal("Right"))

	for _, updated := range updatedTypes {
		g.Expect(updated.Type().Equals(types[updated.Name()].Type())).To(BeTrue())
	}
}

func TestTypeWalker_ContextPropagated(t *testing.T) {
	g := NewGomegaWithT(t)

	types := makeSimpleTestTypeGraph()
	visitor := MakeTypeVisitor()
	walker := NewTypeWalker(types, visitor)

	walked := make(map[TypeName]int)

	walker.WalkFunc = func(this *TypeWalker, def TypeDefinition, ctx interface{}) (TypeDefinition, error) {
		typedCtx := ctx.(int)
		walked[def.Name()] = typedCtx
		return IdentityWalk(this, def, ctx)
	}
	walker.EnqueueContextFunc = func(_ TypeName, ctx interface{}) (interface{}, error) {
		typedCtx := ctx.(int)
		typedCtx += 1

		return typedCtx, nil
	}

	rootDef := types[rootTypeName]
	updatedTypes, err := walker.Walk(rootDef, 0)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(len(walked)).To(Equal(3))
	g.Expect(len(updatedTypes)).To(Equal(3))

	g.Expect(walked[rootTypeName]).To(Equal(0))
	g.Expect(walked[leftTypeName]).To(Equal(1))
	g.Expect(walked[rightTypeName]).To(Equal(1))

	for _, updated := range updatedTypes {
		g.Expect(updated.Type().Equals(types[updated.Name()].Type())).To(BeTrue())
	}
}

func TestTypeWalker_VisitorApplied(t *testing.T) {
	g := NewGomegaWithT(t)

	types := makeSimpleTestTypeGraph()
	visitor := MakeTypeVisitor()
	visitor.VisitObjectType = func(this *TypeVisitor, it *ObjectType, ctx interface{}) (Type, error) {
		// Find any properties of type string and remove them
		for _, prop := range it.Properties() {
			if prop.PropertyType() == StringType {
				it = it.WithoutProperty(prop.PropertyName())
			}
		}

		return IdentityVisitOfObjectType(this, it, ctx)
	}
	walker := NewTypeWalker(types, visitor)

	rootDef := types[rootTypeName]
	updatedTypes, err := walker.Walk(rootDef, nil)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(len(updatedTypes)).To(Equal(3))

	for _, updated := range updatedTypes {
		// Expect only left to be updated - that is only type with string property
		if updated.Name().Equals(leftTypeName) {
			g.Expect(updated.Type().Equals(types[updated.Name()].Type())).To(BeFalse())
			obj, ok := updated.Type().(*ObjectType)
			g.Expect(ok).To(BeTrue())

			g.Expect(len(obj.Properties())).To(Equal(0))
		} else {
			g.Expect(updated.Type().Equals(types[updated.Name()].Type())).To(BeTrue())
		}
	}
}
