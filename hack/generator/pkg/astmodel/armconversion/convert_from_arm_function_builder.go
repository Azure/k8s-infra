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

type convertFromArmBuilder struct {
	conversionBuilder
	typedInputIdent *ast.Ident
	inputIdent      *ast.Ident
}

func newConvertFromArmFunctionBuilder(
	c *ArmConversionFunction,
	codeGenerationContext *astmodel.CodeGenerationContext,
	receiver astmodel.TypeName,
	methodName string) *convertFromArmBuilder {

	result := &convertFromArmBuilder{
		// Note: If we have a property with these names we will have a compilation issue in the generated
		// code. Right now that doesn't seem to be the case anywhere but if it does happen we may need
		// to harden this logic some to choose an unused name.
		typedInputIdent: ast.NewIdent("typedInput"),
		inputIdent:      ast.NewIdent("armInput"),

		conversionBuilder: conversionBuilder{
			methodName:            methodName,
			armType:               c.armType,
			kubeType:              getReceiverObjectType(codeGenerationContext, receiver),
			receiverIdent:         ast.NewIdent(c.idFactory.CreateIdentifier(receiver.Name(), astmodel.NotExported)),
			receiverTypeExpr:      receiver.AsType(codeGenerationContext),
			armTypeIdent:          ast.NewIdent(c.armTypeName.Name()),
			idFactory:             c.idFactory,
			isResource:            c.isResource,
			codeGenerationContext: codeGenerationContext,
		},
	}

	result.propertyConversionHandlers = []propertyConversionHandler{
		result.namePropertyHandler(),
		result.ownerPropertyHandler(),
		result.propertiesWithSameNameAndTypeHandler(),
		result.propertiesWithSameNameButDifferentTypeHandler(),
	}

	return result
}

func (builder *convertFromArmBuilder) functionDeclaration() *ast.FuncDecl {

	return astbuilder.DefineFunc(
		astbuilder.FuncDetails{
			Name:          ast.NewIdent(builder.methodName),
			ReceiverIdent: builder.receiverIdent,
			ReceiverType: &ast.StarExpr{
				X: builder.receiverTypeExpr,
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
						builder.inputIdent,
					},
				},
			},
			Returns: []*ast.Field{
				{
					Type: ast.NewIdent("error"),
				},
			},
			Body: builder.functionBodyStatements(),
		})
}

func (builder *convertFromArmBuilder) functionBodyStatements() []ast.Stmt {
	var result []ast.Stmt

	// perform a type assert and check its results
	result = append(result, builder.assertInputTypeIsArm()...)

	// Do all of the assignments for each property
	result = append(
		result,
		generateTypeConversionAssignments(
			builder.armType,
			builder.kubeType,
			builder.propertyConversionHandler)...)

	// Return nil error if we make it to the end
	result = append(
		result,
		&ast.ReturnStmt{
			Results: []ast.Expr{
				ast.NewIdent("nil"),
			},
		})

	return result
}

func (builder *convertFromArmBuilder) assertInputTypeIsArm() []ast.Stmt {
	var result []ast.Stmt

	// perform a type assert
	result = append(
		result,
		astbuilder.TypeAssert(builder.typedInputIdent, builder.inputIdent, builder.armTypeIdent))

	// Check the result of the type assert
	result = append(
		result,
		astbuilder.ReturnIfNotOk(
			astbuilder.FormatError(
				fmt.Sprintf("\"unexpected type supplied for FromArm function. Expected %s, got %%T\"", builder.armTypeIdent.Name),
				builder.inputIdent)))

	return result
}

//////////////////////
// Conversion handlers
//////////////////////

func (builder *convertFromArmBuilder) namePropertyHandler() propertyConversionHandler {

	return propertyConversionHandler{
		finder: func(toProp *astmodel.PropertyDefinition, fromType *astmodel.ObjectType) (bool, *astmodel.PropertyDefinition) {
			if toProp.Equals(GetAzureNameProperty(builder.idFactory)) && builder.isResource {
				// Check to make sure that the ARM object has a "Name" property (which matches our "AzureName")
				fromProperty, ok := fromType.Property(astmodel.PropertyName("Name"))
				if !ok {
					panic("Arm resource missing property Name")
				}
				return true, fromProperty
			}

			return false, nil
		},
		matchHandler: func(toProp *astmodel.PropertyDefinition, fromProp *astmodel.PropertyDefinition) []ast.Stmt {
			result := astbuilder.SimpleAssignment(
				&ast.SelectorExpr{
					X:   builder.receiverIdent,
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
							X:   builder.typedInputIdent,
							Sel: ast.NewIdent(string(fromProp.PropertyName())),
						},
					},
				})

			return []ast.Stmt{result}
		},
	}
}

func (builder *convertFromArmBuilder) ownerPropertyHandler() propertyConversionHandler {

	return propertyConversionHandler{
		finder: func(toProp *astmodel.PropertyDefinition, _ *astmodel.ObjectType) (bool, *astmodel.PropertyDefinition) {

			if toProp.PropertyName() == astmodel.PropertyName("Owner") && builder.isResource {
				// This property doesn't exist on the fromType, so we fabricate one (it's not actually used but we need to match)
				return true, nil
			}
			return false, nil
		},
		matchHandler: func(toProp *astmodel.PropertyDefinition, _ *astmodel.PropertyDefinition) []ast.Stmt {

			result := astbuilder.SimpleAssignment(
				&ast.SelectorExpr{
					X:   builder.receiverIdent,
					Sel: ast.NewIdent(string(toProp.PropertyName())),
				},
				token.ASSIGN,
				ast.NewIdent("owner"))
			return []ast.Stmt{result}
		},
	}
}

func (builder *convertFromArmBuilder) propertiesWithSameNameAndTypeHandler() propertyConversionHandler {

	return propertyConversionHandler{
		finder: func(toProp *astmodel.PropertyDefinition, fromType *astmodel.ObjectType) (bool, *astmodel.PropertyDefinition) {

			fromProp, ok := fromType.Property(toProp.PropertyName())
			if ok && toProp.PropertyType().Equals(fromProp.PropertyType()) {
				return true, fromProp
			}
			return false, nil
		},
		matchHandler: func(toProp *astmodel.PropertyDefinition, fromProp *astmodel.PropertyDefinition) []ast.Stmt {

			result := astbuilder.SimpleAssignment(
				&ast.SelectorExpr{
					X:   builder.receiverIdent,
					Sel: ast.NewIdent(string(fromProp.PropertyName())),
				},
				token.ASSIGN,
				&ast.SelectorExpr{
					X:   builder.typedInputIdent,
					Sel: ast.NewIdent(string(toProp.PropertyName())),
				})
			return []ast.Stmt{result}
		},
	}
}

func (builder *convertFromArmBuilder) propertiesWithSameNameButDifferentTypeHandler() propertyConversionHandler {
	temp := false
	definedErrVar := &temp

	return propertyConversionHandler{
		finder: func(toProp *astmodel.PropertyDefinition, fromType *astmodel.ObjectType) (bool, *astmodel.PropertyDefinition) {
			fromProp, ok := fromType.Property(toProp.PropertyName())
			if ok && !toProp.PropertyType().Equals(fromProp.PropertyType()) {
				return true, fromProp
			}
			return false, nil
		},
		matchHandler: func(toProp *astmodel.PropertyDefinition, fromProp *astmodel.PropertyDefinition) []ast.Stmt {
			var result []ast.Stmt

			if !*definedErrVar {
				result = append(
					result,
					astbuilder.SimpleVariableDeclaration(ast.NewIdent("err"), ast.NewIdent("error")))
				*definedErrVar = true
			}

			complexConversion := builder.fromArmComplexPropertyConversion(
				&ast.SelectorExpr{
					X:   builder.typedInputIdent,
					Sel: ast.NewIdent(string(fromProp.PropertyName())),
				},
				&ast.SelectorExpr{
					X:   builder.receiverIdent,
					Sel: ast.NewIdent(string(toProp.PropertyName())),
				},
				toProp.PropertyType(),
				string(toProp.PropertyName()),
				nil,
				nil)

			result = append(result, complexConversion...)
			return result
		},
	}
}

//////////////////////////////////////////////////////////////////////////////////
// Complex property conversion (for when properties aren't simple primitive types)
//////////////////////////////////////////////////////////////////////////////////

func (builder *convertFromArmBuilder) fromArmComplexPropertyConversion(
	source ast.Expr,
	destination ast.Expr,
	destinationType astmodel.Type,
	nameHint string,
	conversionContext []astmodel.Type,
	assignmentHandler func(result ast.Expr, destination ast.Expr) ast.Stmt) []ast.Stmt {

	switch concreteType := destinationType.(type) {
	case *astmodel.OptionalType:
		return builder.convertComplexOptionalProperty(
			source,
			destination,
			concreteType,
			nameHint,
			conversionContext,
			assignmentHandler)
	case *astmodel.ArrayType:
		return builder.convertComplexArrayProperty(
			source,
			destination,
			concreteType,
			nameHint,
			conversionContext,
			assignmentHandler)
	case *astmodel.MapType:
		return builder.convertComplexMapProperty(
			source,
			destination,
			concreteType,
			nameHint,
			conversionContext,
			assignmentHandler)
	case astmodel.TypeName:
		return builder.convertComplexTypeNameProperty(
			source,
			destination,
			concreteType,
			nameHint,
			conversionContext,
			assignmentHandler)
	default:
		panic(fmt.Sprintf("don't know how to perform fromArm conversion for type: %T", destinationType))
	}
}

// convertComplexOptionalProperty handles conversion for optional properties with complex elements
// This function generates code that looks like this:
// 	if <source> != nil {
//		<code for producing result from destinationType.Element()>
//		<destination> = &<result>
//	}
func (builder *convertFromArmBuilder) convertComplexOptionalProperty(
	source ast.Expr,
	destination ast.Expr,
	destinationType *astmodel.OptionalType,
	nameHint string,
	conversionContext []astmodel.Type,
	_ func(lhs ast.Expr, rhs ast.Expr) ast.Stmt) []ast.Stmt {

	handler := func(lhs ast.Expr, rhs ast.Expr) ast.Stmt {
		return astbuilder.SimpleAssignment(
			lhs,
			token.ASSIGN,
			&ast.UnaryExpr{
				Op: token.AND,
				X:  rhs,
			})
	}

	innerStatements := builder.fromArmComplexPropertyConversion(
		source,
		destination,
		destinationType.Element(),
		nameHint,
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

	return []ast.Stmt{result}
}

// convertComplexArrayProperty handles conversion for array properties with complex elements
// This function generates code that looks like this:
// 	for _, item := range <source> {
//		<code for producing result from destinationType.Element()>
//		<destination> = append(<destination>, <result>)
//	}
func (builder *convertFromArmBuilder) convertComplexArrayProperty(
	source ast.Expr,
	destination ast.Expr,
	destinationType *astmodel.ArrayType,
	_ string,
	conversionContext []astmodel.Type,
	assignmentHandler func(lhs ast.Expr, rhs ast.Expr) ast.Stmt) []ast.Stmt {

	var results []ast.Stmt

	itemIdent := ast.NewIdent("item")
	elemIdent := ast.NewIdent("elem")

	depth := countArraysAndMapsInConversionContext(conversionContext)

	elemType := destinationType.Element()
	actualDestination := destination // TODO: improve name
	if depth > 0 {
		actualDestination = elemIdent
		results = append(
			results,
			astbuilder.SimpleVariableDeclaration(
				elemIdent,
				destinationType.AsType(builder.codeGenerationContext)))
		elemIdent = ast.NewIdent(fmt.Sprintf("elem%d", depth))
	}

	result := &ast.RangeStmt{
		Key:   ast.NewIdent("_"),
		Value: itemIdent,
		X:     source,
		Tok:   token.DEFINE,
		Body: &ast.BlockStmt{
			List: builder.fromArmComplexPropertyConversion(
				itemIdent,
				actualDestination,
				elemType,
				elemIdent.Name,
				append(conversionContext, destinationType),
				astbuilder.AppendList),
		},
	}
	results = append(results, result)
	if depth > 0 {
		results = append(results, assignmentHandler(destination, actualDestination))
	}

	return results
}

// convertComplexMapProperty handles conversion for map properties with complex values.
// This function panics if the map keys are not primitive types.
// This function generates code that looks like this:
// 	if <source> != nil {
//		<destination> = make(map[<destinationType.KeyType()]<destinationType.ValueType()>)
//		for key, value := range <source> {
// 			<code for producing result from destinationType.ValueType()>
//			<destination>[key] = <result>
//		}
//	}
func (builder *convertFromArmBuilder) convertComplexMapProperty(
	source ast.Expr,
	destination ast.Expr,
	destinationType *astmodel.MapType,
	_ string,
	conversionContext []astmodel.Type,
	assignmentHandler func(lhs ast.Expr, rhs ast.Expr) ast.Stmt) []ast.Stmt {

	if _, ok := destinationType.KeyType().(*astmodel.PrimitiveType); !ok {
		panic(fmt.Sprintf("map had non-primitive key type: %v", destinationType.KeyType()))
	}

	depth := countArraysAndMapsInConversionContext(conversionContext)

	keyIdent := ast.NewIdent("key")
	valueIdent := ast.NewIdent("value")
	elemIdent := ast.NewIdent("elem")

	actualDestination := destination // TODO: improve name
	makeMapToken := token.ASSIGN
	if depth > 0 {
		actualDestination = elemIdent
		elemIdent = ast.NewIdent(fmt.Sprintf("elem%d", depth))
		makeMapToken = token.DEFINE
	}

	handler := func(lhs ast.Expr, rhs ast.Expr) ast.Stmt {
		return astbuilder.InsertMap(lhs, keyIdent, rhs)
	}

	keyTypeAst := destinationType.KeyType().AsType(builder.codeGenerationContext)
	valueTypeAst := destinationType.ValueType().AsType(builder.codeGenerationContext)

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
			List: builder.fromArmComplexPropertyConversion(
				valueIdent,
				actualDestination,
				destinationType.ValueType(),
				elemIdent.Name,
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
		result.Body.List = append(result.Body.List, assignmentHandler(destination, actualDestination))
	}

	return []ast.Stmt{result}
}

// convertComplexTypeNameProperty handles conversion of complex TypeName properties.
// This function generates code that looks like this:
//	<nameHint> := <destinationType>{}
//	err = <nameHint>.FromArm(owner, <source>)
//	if err != nil {
//		return err
//	}
//	<destination> = <nameHint>
func (builder *convertFromArmBuilder) convertComplexTypeNameProperty(
	source ast.Expr,
	destination ast.Expr,
	destinationType astmodel.TypeName,
	nameHint string,
	_ []astmodel.Type,
	assignmentHandler func(lhs ast.Expr, rhs ast.Expr) ast.Stmt) []ast.Stmt {

	var results []ast.Stmt

	propertyLocalVarName := ast.NewIdent(builder.idFactory.CreateIdentifier(nameHint, astmodel.NotExported))

	results = append(results, astbuilder.NewStruct(propertyLocalVarName, ast.NewIdent(destinationType.Name())))
	results = append(
		results,
		astbuilder.SimpleAssignment(
			ast.NewIdent("err"),
			token.ASSIGN,
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   propertyLocalVarName,
					Sel: ast.NewIdent(builder.methodName),
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
			assignmentHandler(destination, propertyLocalVarName))
	}

	return results
}
