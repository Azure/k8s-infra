/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package armconversion

import (
	"fmt"
	"go/token"
	"sync"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/dave/dst"
)

type conversionBuilder struct {
	receiverIdent              string
	receiverTypeExpr           dst.Expr
	armTypeIdent               string
	codeGenerationContext      *astmodel.CodeGenerationContext
	idFactory                  astmodel.IdentifierFactory
	isSpecType                 bool
	methodName                 string
	kubeType                   *astmodel.ObjectType
	armType                    *astmodel.ObjectType
	propertyConversionHandlers []propertyConversionHandler
}

type TypeKind int

const (
	OrdinaryType TypeKind = iota
	SpecType
	StatusType
)

func (builder conversionBuilder) propertyConversionHandler(
	toProp *astmodel.PropertyDefinition,
	fromType *astmodel.ObjectType) []dst.Stmt {

	for _, conversionHandler := range builder.propertyConversionHandlers {
		stmts := conversionHandler(toProp, fromType)
		if len(stmts) > 0 {
			return stmts
		}
	}

	panic(fmt.Sprintf("No property found for %s in method %s\nFrom: %+v\nTo: %+v", toProp.PropertyName(), builder.methodName, *builder.kubeType, *builder.armType))
}

// deepCopyJSON special cases copying JSON-type fields to call the DeepCopy method.
// It generates code that looks like:
//     <destination> = *<source>.DeepCopy()
func (builder *conversionBuilder) deepCopyJSON(
	params complexPropertyConversionParameters) []dst.Stmt {
	newSource := &dst.UnaryExpr{
		X: &dst.CallExpr{
			Fun: &dst.SelectorExpr{
				X:   params.Source(),
				Sel: dst.NewIdent("DeepCopy"),
			},
			Args: []dst.Expr{},
		},
		Op: token.MUL,
	}
	assignmentHandler := params.assignmentHandler
	if assignmentHandler == nil {
		assignmentHandler = assignmentHandlerAssign
	}
	return []dst.Stmt{
		assignmentHandler(params.Destination(), newSource),
	}
}

type propertyConversionHandler = func(toProp *astmodel.PropertyDefinition, fromType *astmodel.ObjectType) []dst.Stmt

var once sync.Once
var azureNameProperty *astmodel.PropertyDefinition

func initializeAzureName(idFactory astmodel.IdentifierFactory) {
	azureNameFieldDescription := "The name of the resource in Azure. This is often the same as" +
		" the name of the resource in Kubernetes but it doesn't have to be."
	azureNameProperty = astmodel.NewPropertyDefinition(
		idFactory.CreatePropertyName(astmodel.AzureNameProperty, astmodel.Exported),
		idFactory.CreateIdentifier(astmodel.AzureNameProperty, astmodel.NotExported),
		astmodel.StringType).WithDescription(azureNameFieldDescription)
}

// GetAzureNameProperty returns the special "AzureName" field
func GetAzureNameProperty(idFactory astmodel.IdentifierFactory) *astmodel.PropertyDefinition {
	once.Do(func() { initializeAzureName(idFactory) })

	return azureNameProperty
}

func getReceiverObjectType(codeGenerationContext *astmodel.CodeGenerationContext, receiver astmodel.TypeName) *astmodel.ObjectType {
	// Determine the type we're operating on
	rt, err := codeGenerationContext.GetImportedDefinition(receiver)
	if err != nil {
		panic(err)
	}

	receiverType, ok := rt.Type().(*astmodel.ObjectType)
	if !ok {
		// Don't expect to have any wrapper types left at this point
		panic(fmt.Sprintf("receiver for ARMConversionFunction is not of expected type. TypeName: %v, Type %T", receiver, rt.Type()))
	}

	return receiverType
}

func generateTypeConversionAssignments(
	fromType *astmodel.ObjectType,
	toType *astmodel.ObjectType,
	propertyHandler func(toProp *astmodel.PropertyDefinition, fromType *astmodel.ObjectType) []dst.Stmt) []dst.Stmt {

	var result []dst.Stmt
	for _, toField := range toType.Properties() {
		fieldConversionStmts := propertyHandler(toField, fromType)
		result = append(result, fieldConversionStmts...)
	}

	return result
}

// NewARMTransformerImpl creates a new interface with the specified ARM conversion functions
func NewARMTransformerImpl(
	armTypeName astmodel.TypeName,
	armType *astmodel.ObjectType,
	idFactory astmodel.IdentifierFactory,
	typeKind TypeKind) *astmodel.InterfaceImplementation {

	var convertToARMFunc *ConvertToARMFunction
	if typeKind != StatusType {
		// status type should not have ConvertToARM
		convertToARMFunc = &ConvertToARMFunction{
			ARMConversionFunction: ARMConversionFunction{
				armTypeName: armTypeName,
				armType:     armType,
				idFactory:   idFactory,
				isSpecType:  typeKind == SpecType,
			},
		}
	}

	populateFromARMFunc := &PopulateFromARMFunction{
		ARMConversionFunction: ARMConversionFunction{
			armTypeName: armTypeName,
			armType:     armType,
			idFactory:   idFactory,
			isSpecType:  typeKind == SpecType,
		},
	}

	createEmptyARMValueFunc := CreateEmptyARMValueFunc{idFactory: idFactory, armTypeName: armTypeName}

	if convertToARMFunc != nil {
		// can convert both to and from ARM = the ARMTransformer interface
		return astmodel.NewInterfaceImplementation(
			astmodel.MakeTypeName(astmodel.GenRuntimeReference, "ARMTransformer"),
			createEmptyARMValueFunc,
			convertToARMFunc,
			populateFromARMFunc)
	} else {
		// only convert in one direction with the FromARMConverter interface
		return astmodel.NewInterfaceImplementation(
			astmodel.MakeTypeName(astmodel.GenRuntimeReference, "FromARMConverter"),
			createEmptyARMValueFunc,
			populateFromARMFunc)
	}
}

type complexPropertyConversionParameters struct {
	source            dst.Expr
	destination       dst.Expr
	destinationType   astmodel.Type
	nameHint          string
	conversionContext []astmodel.Type
	assignmentHandler func(destination, source dst.Expr) dst.Stmt

	// sameTypes indicates that the source and destination types are
	// the same, so no conversion between ARM and non-ARM types is
	// required (although structure copying is).
	sameTypes bool
}

func (params complexPropertyConversionParameters) Source() dst.Expr {
	return dst.Clone(params.source).(dst.Expr)
}

func (params complexPropertyConversionParameters) Destination() dst.Expr {
	return dst.Clone(params.destination).(dst.Expr)
}

func (params complexPropertyConversionParameters) copy() complexPropertyConversionParameters {
	result := params
	result.conversionContext = nil
	result.conversionContext = append(result.conversionContext, params.conversionContext...)

	return result
}

func (params complexPropertyConversionParameters) withAdditionalConversionContext(t astmodel.Type) complexPropertyConversionParameters {
	result := params.copy()
	result.conversionContext = append(result.conversionContext, t)

	return result
}

func (params complexPropertyConversionParameters) withSource(source dst.Expr) complexPropertyConversionParameters {
	result := params.copy()
	result.source = source

	return result
}

func (params complexPropertyConversionParameters) withDestination(destination dst.Expr) complexPropertyConversionParameters {
	result := params.copy()
	result.destination = destination

	return result
}

func (params complexPropertyConversionParameters) withDestinationType(t astmodel.Type) complexPropertyConversionParameters {
	result := params.copy()
	result.destinationType = t

	return result
}

func (params complexPropertyConversionParameters) withAssignmentHandler(
	assignmentHandler func(result dst.Expr, destination dst.Expr) dst.Stmt) complexPropertyConversionParameters {
	result := params.copy()
	result.assignmentHandler = assignmentHandler

	return result
}

// countArraysAndMapsInConversionContext returns the number of arrays/maps which are in the conversion context.
// This is to aid in situations where there are deeply nested conversions (i.e. array of map of maps). In these contexts,
// just using a simple assignment such as "elem := ..." isn't sufficient because elem my already have been defined above by
// an enclosing map/array conversion context. We use the depth to do "elem1 := ..." or "elem7 := ...".
func (params complexPropertyConversionParameters) countArraysAndMapsInConversionContext() int {
	result := 0
	for _, t := range params.conversionContext {
		switch t.(type) {
		case *astmodel.MapType:
			result += 1
		case *astmodel.ArrayType:
			result += 1
		}
	}

	return result
}
