/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astbuilder"
	"go/ast"
)

// NewArmSpecInterfaceImpl creates a new interface implementation with the functions required to implement the
// genruntime.ArmResourceSpec interface
func NewArmSpecInterfaceImpl(
	idFactory IdentifierFactory,
	spec *ObjectType) *InterfaceImplementation {

	funcs := map[string]Function{
		"GetName": &objectFunction{
			o:         spec,
			idFactory: idFactory,
			asFunc:    getNameFunction,
		},
		"GetType": &objectFunction{
			o:         spec,
			idFactory: idFactory,
			asFunc:    getTypeFunction,
		},
		"GetApiVersion": &objectFunction{
			o:         spec,
			idFactory: idFactory,
			asFunc:    getApiVersionFunction,
		},
	}

	result := NewInterfaceImplementation(
		MakeTypeName(MakeGenRuntimePackageReference(), "ArmResourceSpec"),
		funcs)

	return result
}

func getNameFunction(k *objectFunction, codeGenerationContext *CodeGenerationContext, receiver TypeName, methodName string) *ast.FuncDecl {
	return armSpecInterfaceSimpleGetFunction(
		k,
		codeGenerationContext,
		receiver,
		methodName,
		"Name",
		false)
}

func getTypeFunction(k *objectFunction, codeGenerationContext *CodeGenerationContext, receiver TypeName, methodName string) *ast.FuncDecl {
	return armSpecInterfaceSimpleGetFunction(
		k,
		codeGenerationContext,
		receiver,
		methodName,
		"Type",
		true)
}

func getApiVersionFunction(k *objectFunction, codeGenerationContext *CodeGenerationContext, receiver TypeName, methodName string) *ast.FuncDecl {
	return armSpecInterfaceSimpleGetFunction(
		k,
		codeGenerationContext,
		receiver,
		methodName,
		"ApiVersion",
		true)
}

func armSpecInterfaceSimpleGetFunction(
	k *objectFunction,
	codeGenerationContext *CodeGenerationContext,
	receiver TypeName,
	methodName string,
	propertyName string,
	castToString bool) *ast.FuncDecl {

	receiverIdent := ast.NewIdent(k.idFactory.CreateIdentifier(receiver.Name(), NotExported))
	receiverType := receiver.AsType(codeGenerationContext)

	var result ast.Expr
	result = &ast.SelectorExpr{
		X:   receiverIdent,
		Sel: ast.NewIdent(propertyName),
	}

	// This is not the most beautiful thing but it saves some code.
	if castToString {
		result = &ast.CallExpr{
			Fun: ast.NewIdent("string"),
			Args: []ast.Expr{
				result,
			},
		}
	}

	return astbuilder.DefineFunc(
		astbuilder.FuncDetails{
			Name:          ast.NewIdent(methodName),
			ReceiverIdent: receiverIdent,
			// TODO: We're too loosey-goosey here with ptr vs value receiver.
			// TODO: We basically need to use a value receiver right now because
			// TODO: ConvertToArm always returns a value, but for other interface impls
			// TODO: for example on resource we use ptr receiver... the inconsistency is
			// TODO: awkward...
			ReceiverType: receiverType,
			Comment:      fmt.Sprintf("returns the %s of the resource", propertyName),
			Params:       nil,
			Returns: []*ast.Field{
				{
					Type: ast.NewIdent("string"),
				},
			},
			Body: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						result,
					},
				},
			},
		})
}
