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

func (c *ArmConversionFunction) asConvertFromArmFunc(
	codeGenerationContext *astmodel.CodeGenerationContext,
	receiver astmodel.TypeName,
	methodName string) *ast.FuncDecl {

	// Determine the type we're operating on
	kubeType := getReceiverObjectType(codeGenerationContext, receiver)

	var fromArmStatements []ast.Stmt

	receiverIdent := ast.NewIdent(c.idFactory.CreateIdentifier(receiver.Name(), astmodel.NotExported))

	// Note: If we have a property with these names we will have a compilation issue in the generated
	// code. Right now that doesn't seem to be the case anywhere but if it does happen we may need
	// to harden this logic some to choose an unused name.
	inputIdent := ast.NewIdent("armInput")
	typedInputIdent := ast.NewIdent("typedInput")

	// perform a type assert
	fromArmStatements = append(fromArmStatements, &ast.AssignStmt{
		Lhs: []ast.Expr{
			typedInputIdent,
			ast.NewIdent("ok"),
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.TypeAssertExpr{
				X:    inputIdent,
				Type: ast.NewIdent(c.armTypeName.Name()),
			},
		},
	})

	// Check the result of the type assert
	fromArmStatements = append(fromArmStatements, &ast.IfStmt{
		Cond: &ast.UnaryExpr{
			Op: token.NOT,
			X:  ast.NewIdent("ok"),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("fmt"),
								Sel: ast.NewIdent("Errorf"),
							},
							Args: []ast.Expr{
								&ast.BasicLit{
									Kind:  token.STRING,
									Value: fmt.Sprintf("\"unexpected type supplied for FromArm function. Expected %s, got %%T\"", c.armTypeName.Name()),
								},
								inputIdent,
							},
						},
					},
				},
			},
		},
	})

	definedErr := false

	fromArmFieldConversionWrapper := func(toProp *astmodel.PropertyDefinition, fromType *astmodel.ObjectType) []ast.Stmt {
		return c.fromArmFieldConversionHandler(codeGenerationContext, methodName, typedInputIdent, receiverIdent, toProp, fromType, &definedErr)
	}

	// Do all of the assignments
	fromArmStatements = append(
		fromArmStatements,
		generateTypeConversionAssignments(
			c.armType,
			kubeType,
			fromArmFieldConversionWrapper)...)
	fromArmStatements = append(fromArmStatements, &ast.ReturnStmt{
		Results: []ast.Expr{
			ast.NewIdent("nil"),
		},
	})

	fromArmFunc := astbuilder.DefineFunc(
		astbuilder.FuncDetails{
			Name:          ast.NewIdent(methodName),
			ReceiverIdent: receiverIdent,
			ReceiverType: &ast.StarExpr{
				X: ast.NewIdent(receiver.Name()),
			},
			Comment: "converts from an Azure ARM object to a Kubernetes CRD object",
			Params: []*ast.Field{
				{
					Type: &ast.SelectorExpr{
						X:   ast.NewIdent(astmodel.GenRuntimePackageName),
						Sel: ast.NewIdent("KnownResourceReference"),
					},
					Names: []*ast.Ident{
						ast.NewIdent("owner"),
					},
				},
				{
					Type: ast.NewIdent("interface{}"),
					Names: []*ast.Ident{
						inputIdent,
					},
				},
			},
			Returns: []*ast.Field{
				{
					Type: ast.NewIdent("error"),
				},
			},
			Body: fromArmStatements,
		})

	return fromArmFunc
}

func (c *ArmConversionFunction) fromArmComplexPropertyConversion(
	source ast.Expr,
	destination ast.Expr,
	destinationType astmodel.Type,
	methodName string,
	nameHint string, // TODO: remove this or clarify it?
	codeGenerationContext *astmodel.CodeGenerationContext,
	conversionContext []astmodel.Type,
	assignmentHandler func(result ast.Expr, destination ast.Expr) ast.Stmt) []ast.Stmt {

	var results []ast.Stmt

	switch concreteType := destinationType.(type) {
	case *astmodel.OptionalType:
		handler := func(result ast.Expr, destination ast.Expr) ast.Stmt {
			return astbuilder.SimpleAssignment(
				destination,
				token.ASSIGN,
				&ast.UnaryExpr{
					Op: token.AND,
					X:  result,
				})
		}

		innerStatements := c.fromArmComplexPropertyConversion(
			source,
			destination,
			concreteType.Element(),
			methodName,
			nameHint,
			codeGenerationContext,
			append(conversionContext, destinationType),
			handler)

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
		itemIdent := ast.NewIdent("item")
		depth := countArraysAndMapsInConversionContext(conversionContext)

		elemIdent := ast.NewIdent("elem")
		elemType := concreteType.Element()
		actualDestination := destination
		if depth > 0 {
			actualDestination = elemIdent
			results = append(
				results,
				astbuilder.SimpleVariableDeclaration(
					elemIdent,
					concreteType.AsType(codeGenerationContext)))
			elemIdent = ast.NewIdent(fmt.Sprintf("elem%d", depth))
		}

		handler := func(result ast.Expr, destination ast.Expr) ast.Stmt {
			return astbuilder.AppendList(destination, result)
		}

		result := &ast.RangeStmt{
			Key:   ast.NewIdent("_"),
			Value: itemIdent,
			X:     source,
			Tok:   token.DEFINE,
			Body: &ast.BlockStmt{
				List: c.fromArmComplexPropertyConversion(
					itemIdent,
					actualDestination,
					elemType,
					methodName,
					elemIdent.Name,
					codeGenerationContext,
					append(conversionContext, destinationType),
					handler),
			},
		}
		results = append(results, result)
		if depth > 0 {
			results = append(results, assignmentHandler(actualDestination, destination))
		}
	case *astmodel.MapType:
		// Assumption here that the key type is a primitive type...
		if _, ok := concreteType.KeyType().(*astmodel.PrimitiveType); !ok {
			panic(fmt.Sprintf("map had non-primitive key type: %v", concreteType.KeyType()))
		}

		depth := countArraysAndMapsInConversionContext(conversionContext)

		keyIdent := ast.NewIdent("key")
		valueIdent := ast.NewIdent("value")
		elemIdent := ast.NewIdent("elem")

		actualDestination := destination
		makeMapToken := token.ASSIGN
		if depth > 0 {
			actualDestination = elemIdent
			elemIdent = ast.NewIdent(fmt.Sprintf("elem%d", depth))
			makeMapToken = token.DEFINE
		}

		handler := func(result ast.Expr, destination ast.Expr) ast.Stmt {
			return astbuilder.InsertMap(destination, keyIdent, result)
		}

		keyTypeAst := concreteType.KeyType().AsType(codeGenerationContext)
		valueTypeAst := concreteType.ValueType().AsType(codeGenerationContext)

		makeMapStatement := astbuilder.SimpleAssignment(
			actualDestination,
			makeMapToken,
			astbuilder.MakeMap(keyTypeAst, valueTypeAst))
		rangeStatement := &ast.RangeStmt{
			Key:   keyIdent,
			Value: valueIdent,
			X:     source,
			Tok:   token.DEFINE,
			Body: &ast.BlockStmt{
				List: c.fromArmComplexPropertyConversion(
					valueIdent,
					actualDestination,
					concreteType.ValueType(),
					methodName,
					elemIdent.Name,
					codeGenerationContext,
					append(conversionContext, destinationType),
					handler),
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
					makeMapStatement,
					rangeStatement,
				},
			},
		}

		if depth > 0 {
			result.Body.List = append(result.Body.List, assignmentHandler(actualDestination, destination))
		}

		results = append(results, result)

	case astmodel.TypeName:
		propertyLocalVarName := ast.NewIdent(c.idFactory.CreateIdentifier(nameHint, astmodel.NotExported))

		results = append(results, astbuilder.NewStruct(propertyLocalVarName, ast.NewIdent(concreteType.Name())))
		results = append(
			results,
			astbuilder.SimpleAssignment(
				ast.NewIdent("err"),
				token.ASSIGN,
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   propertyLocalVarName,
						Sel: ast.NewIdent(methodName),
					},
					Args: []ast.Expr{
						ast.NewIdent("owner"),
						source,
					},
				}))
		results = append(results, astbuilder.CheckErrorAndReturn())
		if assignmentHandler == nil {
			results = append(
				results,
				astbuilder.SimpleAssignment(
					destination,
					token.ASSIGN,
					propertyLocalVarName))
		} else {
			results = append(
				results,
				assignmentHandler(propertyLocalVarName, destination))
		}

	default:
		panic(fmt.Sprintf("don't know how to perform fromArm conversion for type: %T", destinationType))
	}

	return results
}

func (c *ArmConversionFunction) fromArmFieldConversionHandler(
	codeGenerationContext *astmodel.CodeGenerationContext,
	methodName string,
	typedInputIdent *ast.Ident,
	receiverIdent *ast.Ident,
	toProp *astmodel.PropertyDefinition,
	fromType *astmodel.ObjectType,
	definedErr *bool) []ast.Stmt {

	var results []ast.Stmt

	// There are a few special fields which need specific handling -- all the rest are just
	// a simple assignment from the field of the same name
	if toProp.Equals(GetAzureNameField(c.idFactory)) && c.isResource {
		fromField, ok := fromType.Property(astmodel.PropertyName("Name")) // On ARM object this is just "name"
		if !ok {
			panic("Resource missing property AzureName")
		}

		result := astbuilder.SimpleAssignment(
			&ast.SelectorExpr{
				X:   receiverIdent,
				Sel: ast.NewIdent(string(toProp.PropertyName())),
			},
			token.ASSIGN,
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent(astmodel.GenRuntimePackageName),
					Sel: ast.NewIdent("ExtractKubernetesResourceNameFromArmName"),
				},
				Args: []ast.Expr{
					&ast.SelectorExpr{
						X:   typedInputIdent,
						Sel: ast.NewIdent(string(fromField.PropertyName())),
					},
				},
			})

		results = append(results, result)
	} else if toProp.PropertyName() == astmodel.PropertyName("Owner") && c.isResource {
		result := astbuilder.SimpleAssignment(
			&ast.SelectorExpr{
				X:   receiverIdent,
				Sel: ast.NewIdent(string(toProp.PropertyName())),
			},
			token.ASSIGN,
			ast.NewIdent("owner"))

		results = append(results, result)
	} else if fromProp, ok := fromType.Property(toProp.PropertyName()); ok {
		// There are two cases here... the types are the same (primitive types, enums) or
		// the types are different (complex types).
		if toProp.PropertyType().Equals(fromProp.PropertyType()) {
			result := astbuilder.SimpleAssignment(
				&ast.SelectorExpr{
					X:   receiverIdent,
					Sel: ast.NewIdent(string(fromProp.PropertyName())),
				},
				token.ASSIGN,
				&ast.SelectorExpr{
					X:   typedInputIdent,
					Sel: ast.NewIdent(string(toProp.PropertyName())),
				})
			results = append(results, result)
		} else {
			// predefine the err field used later in these conversion statements,
			// so that we don't have to figure out if we need to use := or =

			if !*definedErr {
				results = append(
					results,
					astbuilder.SimpleVariableDeclaration(
						ast.NewIdent("err"),
						ast.NewIdent("error")))
				*definedErr = true
			}

			conversion := c.fromArmComplexPropertyConversion(
				&ast.SelectorExpr{
					X:   typedInputIdent,
					Sel: ast.NewIdent(string(fromProp.PropertyName())),
				},
				&ast.SelectorExpr{
					X:   receiverIdent,
					Sel: ast.NewIdent(string(toProp.PropertyName())),
				},
				toProp.PropertyType(),
				methodName,
				string(toProp.PropertyName()),
				codeGenerationContext,
				nil,
				nil)

			results = append(
				results,
				conversion...)
		}
	} else {
		panic(fmt.Sprintf("No property found for %s", toProp.PropertyName()))
	}

	return results
}
