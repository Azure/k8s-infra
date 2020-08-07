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

type convertToArmBuilder struct {
	conversionBuilder
	resultIdent *ast.Ident
}

func newConvertToArmFunctionBuilder(
	c *ArmConversionFunction,
	codeGenerationContext *astmodel.CodeGenerationContext,
	receiver astmodel.TypeName,
	methodName string) *convertToArmBuilder {

	result := &convertToArmBuilder{
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
		resultIdent: ast.NewIdent("result"),
	}

	result.propertyConversionHandlers = []propertyConversionHandler{
		result.namePropertyHandler(),
		result.typePropertyHandler(),
		result.propertiesWithSameNameAndTypeHandler(),
		result.propertiesWithSameNameButDifferentTypeHandler(),
	}

	return result
}

func (builder *convertToArmBuilder) functionDeclaration() *ast.FuncDecl {
	return astbuilder.DefineFunc(
		astbuilder.FuncDetails{
			Name:          ast.NewIdent(builder.methodName),
			ReceiverIdent: builder.receiverIdent,
			ReceiverType: &ast.StarExpr{
				X: builder.receiverTypeExpr,
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
			Body: builder.functionBodyStatements(),
		})
}

func (builder *convertToArmBuilder) functionBodyStatements() []ast.Stmt {
	var result []ast.Stmt

	// If we are passed a nil receiver just return nil - this is a bit weird
	// but saves us some nil-checks
	result = append(
		result,
		astbuilder.ReturnIfNil(builder.receiverIdent, ast.NewIdent("nil"), ast.NewIdent("nil")))
	result = append(result, astbuilder.NewStruct(builder.resultIdent, builder.armTypeIdent))

	// Each ARM object property needs to be filled out
	result = append(
		result,
		generateTypeConversionAssignments(
			builder.kubeType,
			builder.armType,
			builder.propertyConversionHandler)...)

	returnStatement := &ast.ReturnStmt{
		Results: []ast.Expr{
			builder.resultIdent,
			ast.NewIdent("nil"),
		},
	}
	result = append(result, returnStatement)

	return result
}

//////////////////////
// Conversion handlers
//////////////////////

func (builder *convertToArmBuilder) namePropertyHandler() propertyConversionHandler {

	return propertyConversionHandler{
		finder: func(toProp *astmodel.PropertyDefinition, fromType *astmodel.ObjectType) (bool, *astmodel.PropertyDefinition) {
			_, ok := fromType.Property(GetAzureNameProperty(builder.idFactory).PropertyName())
			if ok && toProp.PropertyName() == "Name" && builder.isResource {
				fromProp, _ := fromType.Property(GetAzureNameProperty(builder.idFactory).PropertyName())
				return true, fromProp
			}

			return false, nil
		},
		matchHandler: func(toProp *astmodel.PropertyDefinition, fromProp *astmodel.PropertyDefinition) []ast.Stmt {
			result := astbuilder.SimpleAssignment(
				&ast.SelectorExpr{
					X:   builder.resultIdent,
					Sel: ast.NewIdent(string(toProp.PropertyName())),
				},
				token.ASSIGN,
				createArmResourceNameForDeployment(
					ast.NewIdent("owningName"),
					&ast.SelectorExpr{
						X:   builder.receiverIdent,
						Sel: ast.NewIdent(string(fromProp.PropertyName())),
					}))
			return []ast.Stmt{result}
		},
	}
}

func (builder *convertToArmBuilder) typePropertyHandler() propertyConversionHandler {

	return propertyConversionHandler{
		finder: func(toProp *astmodel.PropertyDefinition, fromType *astmodel.ObjectType) (bool, *astmodel.PropertyDefinition) {
			if toProp.PropertyName() == "Type" && builder.isResource {
				return true, nil
			}

			return false, nil
		},
		matchHandler: func(toProp *astmodel.PropertyDefinition, fromProp *astmodel.PropertyDefinition) []ast.Stmt {
			propertyType := toProp.PropertyType()
			if optionalType, ok := toProp.PropertyType().(*astmodel.OptionalType); ok {
				propertyType = optionalType.Element()
			}

			enumTypeName, ok := propertyType.(astmodel.TypeName)
			if !ok {
				panic(fmt.Sprintf("Type property was not an enum, was %T", toProp.PropertyName()))
			}

			def, err := builder.codeGenerationContext.GetImportedDefinition(enumTypeName)
			if err != nil {
				panic(err)
			}

			enumType, ok := def.Type().(*astmodel.EnumType)
			if !ok {
				panic(fmt.Sprintf("Enum %v definition was not of type EnumDefinition", enumTypeName))
			}

			optionId := astmodel.GetEnumValueId(def.Name(), enumType.Options()[0])

			result := astbuilder.SimpleAssignment(
				&ast.SelectorExpr{
					X:   builder.resultIdent,
					Sel: ast.NewIdent(string(toProp.PropertyName())),
				},
				token.ASSIGN,
				ast.NewIdent(optionId))

			return []ast.Stmt{result}
		},
	}
}

func (builder *convertToArmBuilder) propertiesWithSameNameAndTypeHandler() propertyConversionHandler {

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
					X:   builder.resultIdent,
					Sel: ast.NewIdent(string(toProp.PropertyName())),
				},
				token.ASSIGN,
				&ast.SelectorExpr{
					X:   builder.receiverIdent,
					Sel: ast.NewIdent(string(fromProp.PropertyName())),
				})

			return []ast.Stmt{result}
		},
	}
}

func (builder *convertToArmBuilder) propertiesWithSameNameButDifferentTypeHandler() propertyConversionHandler {

	return propertyConversionHandler{
		finder: func(toProp *astmodel.PropertyDefinition, fromType *astmodel.ObjectType) (bool, *astmodel.PropertyDefinition) {
			fromProp, ok := fromType.Property(toProp.PropertyName())
			if ok && !toProp.PropertyType().Equals(fromProp.PropertyType()) {
				return true, fromProp
			}
			return false, nil
		},
		matchHandler: func(toProp *astmodel.PropertyDefinition, fromProp *astmodel.PropertyDefinition) []ast.Stmt {
			destination := &ast.SelectorExpr{
				X:   builder.resultIdent,
				Sel: ast.NewIdent(string(toProp.PropertyName())),
			}
			source := &ast.SelectorExpr{
				X:   builder.receiverIdent,
				Sel: ast.NewIdent(string(fromProp.PropertyName())),
			}

			return builder.toArmComplexPropertyConversion(
				source,
				destination,
				toProp.PropertyType(),
				string(toProp.PropertyName()),
				nil,
				assignmentHandlerAssign)
		},
	}
}

//////////////////////////////////////////////////////////////////////////////////
// Complex property conversion (for when properties aren't simple primitive types)
//////////////////////////////////////////////////////////////////////////////////

func (builder *convertToArmBuilder) toArmComplexPropertyConversion(
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
		panic(fmt.Sprintf("don't know how to perform toArm conversion for type: %T", destinationType))
	}
}

func assignmentHandlerDefine(lhs ast.Expr, rhs ast.Expr) ast.Stmt {
	return astbuilder.SimpleAssignment(lhs, token.DEFINE, rhs)
}

func assignmentHandlerAssign(lhs ast.Expr, rhs ast.Expr) ast.Stmt {
	return astbuilder.SimpleAssignment(lhs, token.ASSIGN, rhs)
}

// convertComplexOptionalProperty handles conversion for optional properties with complex elements
// This function generates code that looks like this:
// 	if <source> != nil {
//		<code for producing result from destinationType.Element()>
//		<destination> = &<result>
//	}
func (builder *convertToArmBuilder) convertComplexOptionalProperty(
	source ast.Expr,
	destination ast.Expr,
	destinationType *astmodel.OptionalType,
	nameHint string,
	conversionContext []astmodel.Type,
	_ func(result ast.Expr, destination ast.Expr) ast.Stmt) []ast.Stmt {

	tempVarIdent := ast.NewIdent(builder.idFactory.CreateIdentifier(nameHint+"Typed", astmodel.NotExported))
	tempVarType := destinationType.Element()

	innerStatements := builder.toArmComplexPropertyConversion(
		source,
		tempVarIdent,
		tempVarType,
		nameHint,
		append(conversionContext, destinationType),
		assignmentHandlerDefine)

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
	return []ast.Stmt{result}
}

// convertComplexArrayProperty handles conversion for array properties with complex elements
// This function generates code that looks like this:
// 	for _, item := range <source> {
//		<code for producing result from destinationType.Element()>
//		<destination> = append(<destination>, <result>)
//	}
func (builder *convertToArmBuilder) convertComplexArrayProperty(
	source ast.Expr,
	destination ast.Expr,
	destinationType *astmodel.ArrayType,
	_ string,
	conversionContext []astmodel.Type,
	_ func(result ast.Expr, destination ast.Expr) ast.Stmt) []ast.Stmt {

	var results []ast.Stmt

	depth := countArraysAndMapsInConversionContext(conversionContext)
	typedVarIdent := ast.NewIdent("elemTyped")
	tempVarType := destinationType.Element()
	itemIdent := ast.NewIdent("item")
	elemIdent := ast.NewIdent("elem")

	if depth > 0 {
		results = append(results, astbuilder.SimpleVariableDeclaration(
			typedVarIdent,
			destinationType.AsType(builder.codeGenerationContext)))
		typedVarIdent = ast.NewIdent(fmt.Sprintf("elemTyped%d", depth))
	}

	innerStatements := builder.toArmComplexPropertyConversion(
		itemIdent,
		typedVarIdent,
		tempVarType,
		elemIdent.Name,
		append(conversionContext, destinationType),
		assignmentHandlerDefine)

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
	return results
}

// convertComplexMapProperty handles conversion for map properties with complex values.
// This function panics if the map keys are not primitive types.
// This function generates code that looks like this:
//	<destination> = make(map[<destinationType.KeyType()]<destinationType.ValueType()>)
// 	if <source> != nil {
//		for key, value := range <source> {
// 			<code for producing result from destinationType.ValueType()>
//			<destination>[key] = <result>
//		}
//	}
func (builder *convertToArmBuilder) convertComplexMapProperty(
	source ast.Expr,
	destination ast.Expr,
	destinationType *astmodel.MapType,
	_ string,
	conversionContext []astmodel.Type,
	_ func(result ast.Expr, destination ast.Expr) ast.Stmt) []ast.Stmt {

	if _, ok := destinationType.KeyType().(*astmodel.PrimitiveType); !ok {
		panic(fmt.Sprintf("map had non-primitive key type: %v", destinationType.KeyType()))
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

	innerStatements := builder.toArmComplexPropertyConversion(
		valueIdent,
		typedVarIdent,
		destinationType.ValueType(),
		elemIdent.Name,
		append(conversionContext, destinationType),
		assignmentHandlerDefine)

	// Append the final statement
	innerStatements = append(innerStatements, astbuilder.InsertMap(destination, keyIdent, typedVarIdent))

	keyTypeAst := destinationType.KeyType().AsType(builder.codeGenerationContext)
	valueTypeAst := destinationType.ValueType().AsType(builder.codeGenerationContext)

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
	return []ast.Stmt{makeMapStatement, result}
}

// convertComplexTypeNameProperty handles conversion of complex TypeName properties.
// This function generates code that looks like this:
// 	<nameHint>, err := <source>.ToArm(owningName)
//	if err != nil {
//		return nil, err
//	}
//	<destination> = <nameHint>.(FooArm)
func (builder *convertToArmBuilder) convertComplexTypeNameProperty(
	source ast.Expr,
	destination ast.Expr,
	destinationType astmodel.TypeName,
	nameHint string,
	_ []astmodel.Type,
	assignmentHandler func(result ast.Expr, destination ast.Expr) ast.Stmt) []ast.Stmt {

	var results []ast.Stmt
	propertyLocalVarName := ast.NewIdent(builder.idFactory.CreateIdentifier(nameHint, astmodel.NotExported))

	// Call ToArm on the property
	results = append(results, callToArmFunction(source, propertyLocalVarName, builder.methodName)...)

	typeAssertExpr := &ast.TypeAssertExpr{
		X:    propertyLocalVarName,
		Type: ast.NewIdent(destinationType.Name()),
	}

	results = append(results, assignmentHandler(destination, typeAssertExpr))

	return results
}

func callToArmFunction(source ast.Expr, destination ast.Expr, methodName string) []ast.Stmt {
	var results []ast.Stmt

	// Call ToArm on the property
	propertyToArmInvocation := &ast.AssignStmt{
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
	results = append(results, propertyToArmInvocation)
	results = append(results, astbuilder.CheckErrorAndReturn(ast.NewIdent("nil")))

	return results
}

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
