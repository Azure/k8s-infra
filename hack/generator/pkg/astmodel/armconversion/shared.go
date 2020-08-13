/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package armconversion

import (
	"fmt"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"go/ast"
	"sync"
)

type conversionBuilder struct {
	receiverIdent              *ast.Ident
	receiverTypeExpr           ast.Expr
	armTypeIdent               *ast.Ident
	codeGenerationContext      *astmodel.CodeGenerationContext
	idFactory                  astmodel.IdentifierFactory
	isResource                 bool
	methodName                 string
	kubeType                   *astmodel.ObjectType
	armType                    *astmodel.ObjectType
	propertyConversionHandlers []propertyConversionHandler
}

func (builder conversionBuilder) propertyConversionHandler(
	toProp *astmodel.PropertyDefinition,
	fromType *astmodel.ObjectType) []ast.Stmt {

	for _, conversionHandler := range builder.propertyConversionHandlers {
		matched, matchedProperty := conversionHandler.finder(toProp, fromType)
		if matched {
			return conversionHandler.matchHandler(toProp, matchedProperty)
		}
	}

	panic(fmt.Sprintf("No property found for %s", toProp.PropertyName()))
}

type propertyMatchHandlerFunc = func(toProp *astmodel.PropertyDefinition, fromProp *astmodel.PropertyDefinition) []ast.Stmt

type propertyConversionHandler struct {
	finder       func(toProp *astmodel.PropertyDefinition, fromType *astmodel.ObjectType) (bool, *astmodel.PropertyDefinition)
	matchHandler propertyMatchHandlerFunc
}

var once sync.Once
var azureNameProperty *astmodel.PropertyDefinition

func initializeAzureName(idFactory astmodel.IdentifierFactory) {
	azureName := "azureName"
	azureNameFieldDescription := "The name of the resource in Azure. This is often the same as" +
		" the name of the resource in Kubernetes but it doesn't have to be."
	azureNameProperty = astmodel.NewPropertyDefinition(
		idFactory.CreatePropertyName(azureName, astmodel.Exported),
		azureName,
		astmodel.StringType).WithDescription(azureNameFieldDescription)
}

// GetAzureNameProperty returns the special "AzureName" field
func GetAzureNameProperty(idFactory astmodel.IdentifierFactory) *astmodel.PropertyDefinition {
	once.Do(func() { initializeAzureName(idFactory) })

	return azureNameProperty
}

func getReceiverObjectType(codeGenerationContext *astmodel.CodeGenerationContext, receiver astmodel.TypeName) *astmodel.ObjectType {
	// Determine the type we're operating on
	receiverType, err := codeGenerationContext.GetImportedDefinition(receiver)
	if err != nil {
		panic(err)
	}

	kubeType, ok := receiverType.Type().(*astmodel.ObjectType)
	if !ok {
		panic(fmt.Sprintf("receiver for ArmConversionFunction is not of type ObjectType. TypeName: %v, Type %T", receiver, receiverType.Type()))
	}

	return kubeType
}

func generateTypeConversionAssignments(
	fromType *astmodel.ObjectType,
	toType *astmodel.ObjectType,
	propertyHandler func(toProp *astmodel.PropertyDefinition, fromType *astmodel.ObjectType) []ast.Stmt) []ast.Stmt {

	var result []ast.Stmt
	for _, toField := range toType.Properties() {
		fieldConversionStmts := propertyHandler(toField, fromType)
		result = append(result, fieldConversionStmts...)
	}

	return result
}

// NewArmTransformerImpl creates a new interface with the specified ARM conversion functions
func NewArmTransformerImpl(
	armTypeName astmodel.TypeName,
	armType *astmodel.ObjectType,
	idFactory astmodel.IdentifierFactory,
	isResource bool) *astmodel.Interface {

	funcs := map[string]astmodel.Function{
		"ToArm": &ArmConversionFunction{
			armTypeName: armTypeName,
			armType:     armType,
			idFactory:   idFactory,
			direction:   ConversionDirectionToArm,
			isResource:  isResource,
		},
		"FromArm": &ArmConversionFunction{
			armTypeName: armTypeName,
			armType:     armType,
			idFactory:   idFactory,
			direction:   ConversionDirectionFromArm,
			isResource:  isResource,
		},
	}

	result := astmodel.NewInterface(
		astmodel.MakeTypeName(astmodel.MakeGenRuntimePackageReference(), "ArmTransformer"),
		funcs)

	return result
}

type complexPropertyConversionParameters struct {
	source            ast.Expr
	destination       ast.Expr
	destinationType   astmodel.Type
	nameHint          string
	conversionContext []astmodel.Type
	assignmentHandler func(result ast.Expr, destination ast.Expr) ast.Stmt
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

func (params complexPropertyConversionParameters) withSource(source ast.Expr) complexPropertyConversionParameters {
	result := params.copy()
	result.source = source

	return result
}

func (params complexPropertyConversionParameters) withNameHint(nameHint string) complexPropertyConversionParameters {
	result := params.copy()
	result.nameHint = nameHint

	return result
}

func (params complexPropertyConversionParameters) withDestination(destination ast.Expr) complexPropertyConversionParameters {
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
	assignmentHandler func(result ast.Expr, destination ast.Expr) ast.Stmt) complexPropertyConversionParameters {
	result := params.copy()
	result.assignmentHandler = assignmentHandler

	return result
}

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
