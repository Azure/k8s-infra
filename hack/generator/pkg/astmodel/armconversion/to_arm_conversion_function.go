/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package armconversion

import (
	"fmt"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astbuilder"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"go/ast"
	"go/token"
)

func createArmResourceNameForDeployment(owningNameParameter ast.Expr, input ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(astmodel.GenRuntimePackageName),
			Sel: ast.NewIdent("CreateArmResourceNameForDeployment"),
		},
		Args: []ast.Expr{
			owningNameParameter,
			input,
		},
	}
}

func (c *ArmConversionFunction) asConvertToArmFunc(
	codeGenerationContext *astmodel.CodeGenerationContext,
	receiver astmodel.TypeName,
	methodName string) *ast.FuncDecl {

	// Determine the type we're operating on
	kubeType := getReceiverObjectType(codeGenerationContext, receiver)

	var toArmStatements []ast.Stmt

	// Starting from the kubernetes type, generate
	resultIdent := ast.NewIdent("result")
	receiverIdent := ast.NewIdent(c.idFactory.CreateIdentifier(receiver.Name(), astmodel.NotExported))

	// A bit weird but if we are passed a nil receiver just return nil
	toArmStatements = append(
		toArmStatements,
		astbuilder.NilGuardAndReturn(receiverIdent, ast.NewIdent("nil"), ast.NewIdent("nil")))
	toArmStatements = append(
		toArmStatements,
		astbuilder.NewStruct(resultIdent, ast.NewIdent(c.armTypeName.Name())))

	toArmFieldConversionWrapper := func(toProp *astmodel.PropertyDefinition, fromType *astmodel.ObjectType) []ast.Stmt {
		return c.toArmFieldConversionHandler(codeGenerationContext, methodName, resultIdent, receiverIdent, toProp, fromType)
	}

	// Each ARM object field needs to be filled out
	toArmStatements = append(
		toArmStatements,
		generateTypeConversionAssignments(
			kubeType,
			c.armType,
			toArmFieldConversionWrapper)...)

	returnStatement := &ast.ReturnStmt{
		Results: []ast.Expr{
			resultIdent,
			ast.NewIdent("nil"),
		},
	}
	toArmStatements = append(toArmStatements, returnStatement)

	toArmFunc := astbuilder.DefineFunc(
		astbuilder.FuncDetails{
			Name:          ast.NewIdent(methodName),
			ReceiverIdent: receiverIdent,
			ReceiverType: &ast.StarExpr{
				X: ast.NewIdent(receiver.Name()),
			},
			Comment: "converts from a Kubernetes CRD object to an ARM object",
			Params: []*ast.Field{
				{
					Type: ast.NewIdent("string"),
					Names: []*ast.Ident{
						ast.NewIdent("owningName"),
					},
				},
			},
			Returns: []*ast.Field{
				{
					Type: ast.NewIdent("interface{}"),
				},
				{
					Type: ast.NewIdent("error"),
				},
			},
			Body: toArmStatements,
		})

	return toArmFunc
}

func (c *ArmConversionFunction) toArmFieldConversionHandler(
	codeGenerationContext *astmodel.CodeGenerationContext,
	methodName string,
	resultIdent *ast.Ident,
	receiverIdent *ast.Ident,
	toProp *astmodel.PropertyDefinition,
	fromType *astmodel.ObjectType) []ast.Stmt {

	var results []ast.Stmt

	// There are a few special fields which need specific handling on resource types-- all the rest are just
	// a simple assignment from the field of the same name
	if _, ok := fromType.Property(GetAzureNameField(c.idFactory).PropertyName()); ok && toProp.PropertyName() == astmodel.PropertyName("Name") && c.isResource {
		fromField, _ := fromType.Property(GetAzureNameField(c.idFactory).PropertyName())

		result := astbuilder.SimpleAssignment(
			&ast.SelectorExpr{
				X:   resultIdent,
				Sel: ast.NewIdent(string(toProp.PropertyName())),
			},
			token.ASSIGN,
			createArmResourceNameForDeployment(
				ast.NewIdent("owningName"),
				&ast.SelectorExpr{
					X:   receiverIdent,
					Sel: ast.NewIdent(string(fromField.PropertyName())),
				}))

		results = append(results, result)
	} else if toProp.PropertyName() == astmodel.PropertyName("Type") && c.isResource {
		fieldType := toProp.PropertyType()
		if optionalType, ok := toProp.PropertyType().(*astmodel.OptionalType); ok {
			fieldType = optionalType.Element()
		}

		enumTypeName, ok := fieldType.(astmodel.TypeName)
		if !ok {
			panic(fmt.Sprintf("Type field was not an enum, was %T", toProp.PropertyName()))
		}

		def, err := codeGenerationContext.GetImportedDefinition(enumTypeName)
		if err != nil {
			panic(err) // TODO: Bad
		}

		enumType, ok := def.Type().(*astmodel.EnumType)
		if !ok {
			panic(fmt.Sprintf("Enum %v definition was not of type EnumDefinition", enumTypeName))
		}

		optionId := astmodel.GetEnumValueId(def.Name(), enumType.Options()[0])

		result := astbuilder.SimpleAssignment(
			&ast.SelectorExpr{
				X:   resultIdent,
				Sel: ast.NewIdent(string(toProp.PropertyName())),
			},
			token.ASSIGN,
			ast.NewIdent(optionId))

		results = append(results, result)
	} else if fromProp, ok := fromType.Property(toProp.PropertyName()); ok {
		// There are two cases here... the types are the same (primitive types, enums) or
		// the types are different (complex types).
		if toProp.PropertyType().Equals(fromProp.PropertyType()) {
			result := astbuilder.SimpleAssignment(
				&ast.SelectorExpr{
					X:   resultIdent,
					Sel: ast.NewIdent(string(toProp.PropertyName())),
				},
				token.ASSIGN,
				&ast.SelectorExpr{
					X:   receiverIdent,
					Sel: ast.NewIdent(string(fromProp.PropertyName())),
				})
			results = append(results, result)
		} else {
			destination := &ast.SelectorExpr{
				X:   resultIdent,
				Sel: ast.NewIdent(string(toProp.PropertyName())),
			}
			source := &ast.SelectorExpr{
				X:   receiverIdent,
				Sel: ast.NewIdent(string(fromProp.PropertyName())),
			}

			results = c.toArmComplexPropertyConversion(
				source,
				destination,
				toProp.PropertyType(),
				methodName,
				string(toProp.PropertyName()),
				token.ASSIGN,
				codeGenerationContext,
				nil)
		}
	} else {
		panic(fmt.Sprintf("No property found for %s", toProp.PropertyName()))
	}

	return results
}

func callToArmFunction(source ast.Expr, destination ast.Expr, methodName string) []ast.Stmt {
	var results []ast.Stmt

	// Call ToArm on the field
	fieldToArmInvocation := &ast.AssignStmt{
		Lhs: []ast.Expr{
			destination,
			ast.NewIdent("err"),
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   source,
					Sel: ast.NewIdent(methodName),
				},
				Args: []ast.Expr{
					ast.NewIdent("owningName"),
				},
			},
		},
	}
	results = append(results, fieldToArmInvocation)
	results = append(results, astbuilder.CheckErrorAndReturn(ast.NewIdent("nil")))

	return results
}

func (c *ArmConversionFunction) toArmComplexPropertyConversion(
	source ast.Expr,
	destination ast.Expr,
	destinationType astmodel.Type,
	methodName string,
	nameHint string,
	assignmentToken token.Token, // TODO: Should this be lambda like other one?
	codeGenerationContext *astmodel.CodeGenerationContext,
	conversionContext []astmodel.Type) []ast.Stmt {

	var results []ast.Stmt

	switch concreteType := destinationType.(type) {
	case *astmodel.OptionalType:
		tempVarIdent := ast.NewIdent(c.idFactory.CreateIdentifier(nameHint+"Typed", astmodel.NotExported))
		tempVarType := concreteType.Element()
		innerStatements := c.toArmComplexPropertyConversion(
			source,
			tempVarIdent,
			tempVarType,
			methodName,
			nameHint,
			token.DEFINE,
			codeGenerationContext,
			append(conversionContext, destinationType))

		// Tack on the final assignment
		innerStatements = append(
			innerStatements,
			astbuilder.SimpleAssignment(
				destination,
				token.ASSIGN,
				&ast.UnaryExpr{
					Op: token.AND,
					X:  tempVarIdent,
				}))

		result := &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  source,
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: innerStatements,
			},
		}
		results = append(results, result)

	case *astmodel.ArrayType:
		depth := countArraysAndMapsInConversionContext(conversionContext)
		typedVarIdent := ast.NewIdent("elemTyped")
		tempVarType := concreteType.Element()

		if depth > 0 {
			results = append(results, astbuilder.SimpleVariableDeclaration(
				typedVarIdent,
				concreteType.AsType(codeGenerationContext)))
			typedVarIdent = ast.NewIdent(fmt.Sprintf("elemTyped%d", depth))
		}

		itemIdent := ast.NewIdent("item")
		elemIdent := ast.NewIdent("elem")
		innerStatements := c.toArmComplexPropertyConversion(
			itemIdent,
			typedVarIdent,
			tempVarType,
			methodName,
			elemIdent.Name,
			token.DEFINE,
			codeGenerationContext,
			append(conversionContext, destinationType))

		// Append the final statement
		innerStatements = append(innerStatements, astbuilder.AppendList(destination, typedVarIdent))

		result := &ast.RangeStmt{
			Key:   ast.NewIdent("_"),
			Value: itemIdent,
			X:     source,
			Tok:   token.DEFINE,
			Body: &ast.BlockStmt{
				List: innerStatements,
			},
		}

		results = append(results, result)

	case *astmodel.MapType:
		// Assumption here that the key type is a primitive type...
		if _, ok := concreteType.KeyType().(*astmodel.PrimitiveType); !ok {
			panic(fmt.Sprintf("map had non-primitive key type: %v", concreteType.KeyType()))
		}

		keyIdent := ast.NewIdent("key")
		typedVarIdent := ast.NewIdent("elemTyped")
		valueIdent := ast.NewIdent("value")
		elemIdent := ast.NewIdent("elem")

		depth := countArraysAndMapsInConversionContext(conversionContext)
		makeMapToken := token.ASSIGN
		if depth > 0 {
			typedVarIdent = ast.NewIdent(fmt.Sprintf("elemTyped%d", depth))
			makeMapToken = token.DEFINE
		}

		innerStatements := c.toArmComplexPropertyConversion(
			valueIdent,
			typedVarIdent,
			concreteType.ValueType(),
			methodName,
			elemIdent.Name,
			token.DEFINE,
			codeGenerationContext,
			append(conversionContext, destinationType))

		// Append the final statement
		innerStatements = append(innerStatements, astbuilder.InsertMap(destination, keyIdent, typedVarIdent))

		keyTypeAst := concreteType.KeyType().AsType(codeGenerationContext)
		valueTypeAst := concreteType.ValueType().AsType(codeGenerationContext)

		makeMapStatement := astbuilder.SimpleAssignment(
			destination,
			makeMapToken,
			astbuilder.MakeMap(keyTypeAst, valueTypeAst))
		rangeStatement := &ast.RangeStmt{
			Key:   keyIdent,
			Value: valueIdent,
			X:     source,
			Tok:   token.DEFINE,
			Body: &ast.BlockStmt{
				List: innerStatements,
			},
		}

		result := &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  source,
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					rangeStatement,
				},
			},
		}

		results = append(results, makeMapStatement)
		results = append(results, result)

	case astmodel.TypeName:
		propertyLocalVarName := ast.NewIdent(c.idFactory.CreateIdentifier(nameHint, astmodel.NotExported))

		// Call ToArm on the field
		results = append(results, callToArmFunction(source, propertyLocalVarName, methodName)...)

		typeAssertExpr := &ast.TypeAssertExpr{
			X:    propertyLocalVarName,
			Type: ast.NewIdent(concreteType.Name()),
		}

		results = append(results, astbuilder.SimpleAssignment(
			destination,
			assignmentToken,
			typeAssertExpr))
	default:
		panic(fmt.Sprintf("don't know how to perform toArm conversion for type: %T", destinationType))
	}

	return results
}
