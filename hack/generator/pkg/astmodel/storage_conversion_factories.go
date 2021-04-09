/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"
	"go/token"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astbuilder"
	"github.com/dave/dst"
	"github.com/pkg/errors"
)

// StorageTypeConversion generates the AST for a given conversion.
// reader is an expression to read the original value.
// writer is a function that accepts an expression for reading a value and creates one or more
// statements to write that value.
// Both of these might be complex expressions, possibly involving indexing into arrays or maps.
type StorageTypeConversion func(reader dst.Expr, writer func(dst.Expr) []dst.Stmt, generationContext *CodeGenerationContext) []dst.Stmt

// StorageTypeConversionFactory represents factory methods that can be used to create StorageTypeConversions
// for a specific pair of types
// source is the endpoint that will be read
// destination is the endpoint that will be written
// ctx contains additional information that may be needed when creating a conversion
type StorageTypeConversionFactory func(
	source *StorageConversionEndpoint,
	destination *StorageConversionEndpoint,
	conversionContext *StorageConversionContext) StorageTypeConversion

// A list of all known type conversion factory methods
var typeConversionFactories []StorageTypeConversionFactory

func init() {
	typeConversionFactories = []StorageTypeConversionFactory{
		// Meta-conversions
		assignToOptionalType,
		// Primitive types
		assignPrimitiveTypeFromPrimitiveType,
		assignPrimitiveTypeFromOptionalPrimitiveType,
		// Collection Types
		assignArrayFromArray,
		assignMapFromMap,
		// Enumerations
		assignEnumTypeFromEnumType,
		assignEnumTypeFromOptionalEnumType,
		// Complex object types
		assignObjectTypeFromObjectType,
		assignObjectTypeFromOptionalObjectType,
	}
}

// createTypeConversion tries to create a type conversion between the two provided types, using
// all of the available type conversion functions in priority order to do so.
func createTypeConversion(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint,
	conversionContext *StorageConversionContext) (StorageTypeConversion, error) {
	for _, f := range typeConversionFactories {
		result := f(sourceEndpoint, destinationEndpoint, conversionContext)
		if result != nil {
			return result, nil
		}
	}

	err := errors.Errorf(
		"no conversion found to assign %q from %q",
		destinationEndpoint.name,
		sourceEndpoint.name)

	return nil, err
}

// assignToOptionalType will generate a conversion where the destination is optional, if the
// underlying type of the destination is compatible with the source.
//
// <destination> = &<source>
//
func assignToOptionalType(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint,
	conversionContext *StorageConversionContext) StorageTypeConversion {

	// Require destination to be optional
	dstOpt, dstIsOpt := AsOptionalType(destinationEndpoint.Type())
	if !dstIsOpt {
		// Destination is not optional
		return nil
	}

	// Require a conversion between the wrapped type and our source
	dstEp := destinationEndpoint.WithType(dstOpt.element)
	conversion, _ := createTypeConversion(sourceEndpoint, dstEp, conversionContext)
	if conversion == nil {
		return nil
	}

	//copyVar := destinationEndpoint.CreateLocal("", "Temp")

	return func(reader dst.Expr, writer func(dst.Expr) []dst.Stmt, generationContext *CodeGenerationContext) []dst.Stmt {
		// Create a writer that takes the address of the passed expression
		// Note that we are dependent on any wrapping conversion to ensure no aliasing occurs
		addrOfWriter := func(expr dst.Expr) []dst.Stmt {
			return writer(astbuilder.AddrOf(expr))
		}

		return conversion(reader, addrOfWriter, generationContext)
	}
}

// assignPrimitiveTypeFromPrimitiveType will generate a direct assignment if both types have the
// same underlying primitive type and are not optional
//
// <destination> = <source>
//
func assignPrimitiveTypeFromPrimitiveType(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint,
	_ *StorageConversionContext) StorageTypeConversion {

	// Require source to be non-optional
	if _, srcOpt := AsOptionalType(sourceEndpoint.Type()); srcOpt {
		return nil
	}

	// Require destination to be non-optional
	if _, dstOpt := AsOptionalType(destinationEndpoint.Type()); dstOpt {
		return nil
	}

	// Require source to be a primitive type
	srcPrim, srcOk := AsPrimitiveType(sourceEndpoint.Type())
	if !srcOk {
		return nil
	}

	// Require destination to be a primitive type
	dstPrim, dstOk := AsPrimitiveType(destinationEndpoint.Type())
	if !dstOk {
		return nil
	}

	// Require both properties to have the same primitive type
	if !srcPrim.Equals(dstPrim) {
		return nil
	}

	local := destinationEndpoint.CreateLocal("", "Value")

	return func(reader dst.Expr, writer func(dst.Expr) []dst.Stmt, generationContext *CodeGenerationContext) []dst.Stmt {
		assign := astbuilder.SimpleAssignment(dst.NewIdent(local), token.DEFINE, reader)
		return astbuilder.Statements(assign, writer(dst.NewIdent(local)))
	}
}

// assignPrimitiveTypeFromOptionalPrimitiveType will generate a direct assignment if both types
// have the same underlying primitive type and only the source is optional
//
// if <source> != nil {
//    <destination> = *<source>
// } else {
//    <destination> = <zero>
// }
func assignPrimitiveTypeFromOptionalPrimitiveType(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint,
	_ *StorageConversionContext) StorageTypeConversion {

	// Require source to be optional
	if _, srcOpt := AsOptionalType(sourceEndpoint.Type()); !srcOpt {
		return nil
	}

	// Require source to be a primitive type
	srcPrim, srcOk := AsPrimitiveType(sourceEndpoint.Type())
	if !srcOk {
		return nil
	}

	// Require destination to be non-optional
	if _, dstOpt := AsOptionalType(destinationEndpoint.Type()); dstOpt {
		return nil
	}

	// Require destination to be a primitive type
	dstPrim, dstOk := AsPrimitiveType(destinationEndpoint.Type())
	if !dstOk {
		return nil
	}

	// Require both properties to have the same primitive type
	if !srcPrim.Equals(dstPrim) {
		return nil
	}

	local := destinationEndpoint.CreateLocal("", "Value")

	return func(reader dst.Expr, writer func(dst.Expr) []dst.Stmt, generationContext *CodeGenerationContext) []dst.Stmt {

		declaration := astbuilder.LocalVariableDeclaration(local, dstPrim.AsType(generationContext), "")

		// Need to check for null and only assign if we have a value
		cond := astbuilder.NotEqual(reader, dst.NewIdent("nil"))

		writeLocal := func(expr dst.Expr) []dst.Stmt {
			return []dst.Stmt{
				astbuilder.SimpleAssignment(
					dst.NewIdent(local),
					token.ASSIGN,
					expr),
			}
		}

		updateLocalFromReader := writeLocal(astbuilder.Dereference(reader))

		updateLocalWithZero := writeLocal(&dst.BasicLit{
			Value: zeroValue(srcPrim),
		})

		convert := astbuilder.SimpleIfElse(
			cond,
			astbuilder.StatementBlock(updateLocalFromReader...),
			astbuilder.StatementBlock(updateLocalWithZero...))

		assignValue := writer(dst.NewIdent(local))
		return astbuilder.Statements(declaration, convert, assignValue)
	}
}

// assignArrayFromArray will generate a code fragment to populate an array, assuming the
// underlying types of the two arrays are compatible
//
// <arr> := make([]<type>, len(<reader>))
// for <index>, <value> := range <reader> {
//     // Shadow the loop variable to avoid aliasing
//     <value> := <value>
//     <arr>[<index>] := <value> // Or other conversion as required
// }
// <writer> = <arr>
//
func assignArrayFromArray(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint,
	conversionContext *StorageConversionContext) StorageTypeConversion {

	// Require source to be an array type
	srcArray, srcOk := AsArrayType(sourceEndpoint.Type())
	if !srcOk {
		return nil
	}

	// Require destination to be an array type
	dstArray, dstOk := AsArrayType(destinationEndpoint.Type())
	if !dstOk {
		return nil
	}

	// Require a conversion between the array types
	srcEp := sourceEndpoint.WithType(srcArray.element)
	dstEp := destinationEndpoint.WithType(dstArray.element)
	conversion, _ := createTypeConversion(srcEp, dstEp, conversionContext)
	if conversion == nil {
		return nil
	}

	return func(reader dst.Expr, writer func(dst.Expr) []dst.Stmt, generationContext *CodeGenerationContext) []dst.Stmt {
		// We create three obviously related identifiers to use for the conversion
		itemId := sourceEndpoint.CreateLocal("Item")
		indexId := sourceEndpoint.CreateLocal("Index")
		tempId := sourceEndpoint.CreateLocal("List")

		declaration := astbuilder.SimpleAssignment(
			dst.NewIdent(tempId),
			token.DEFINE,
			astbuilder.MakeList(dstArray.AsType(generationContext), astbuilder.CallFunc("len", reader)))

		writeToElement := func(expr dst.Expr) []dst.Stmt {
			return []dst.Stmt{
				astbuilder.SimpleAssignment(
					&dst.IndexExpr{
						X:     dst.NewIdent(tempId),
						Index: dst.NewIdent(indexId),
					},
					token.ASSIGN,
					expr),
			}
		}

		avoidAliasing := astbuilder.SimpleAssignment(dst.NewIdent(itemId), token.DEFINE, dst.NewIdent(itemId))
		avoidAliasing.Decs.Start.Append("// Shadow the loop variable to avoid aliasing")
		avoidAliasing.Decs.Before = dst.NewLine

		loopBody := astbuilder.Statements(
			avoidAliasing,
			conversion(dst.NewIdent(itemId), writeToElement, generationContext))

		assign := writer(dst.NewIdent(tempId))
		loop := astbuilder.IterateOverListWithIndex(indexId, itemId, reader, loopBody...)
		return astbuilder.Statements(declaration, loop, assign)
	}
}

// assignMapFromMap will generate a code fragment to populate an array, assuming the
// underlying types of the two arrays are compatible
//
// <map> := make(map[<key>]<type>)
// for key, <item> := range <reader> {
//     <map>[<key>] := <item>
// }
// <writer> = <map>
//
func assignMapFromMap(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint,
	conversionContext *StorageConversionContext) StorageTypeConversion {

	// Require source to be a map
	srcMap, isMap := AsMapType(sourceEndpoint.Type())
	if !isMap {
		// Source is not a map
		return nil
	}

	// Require destination to be a map
	dstMap, isMap := AsMapType(destinationEndpoint.Type())
	if !isMap {
		// Destination is not a map
		return nil
	}

	// Require map keys to be identical
	if !srcMap.key.Equals(dstMap.key) {
		// Keys are different types
		return nil
	}

	// Require a conversion between the map items
	srcEp := sourceEndpoint.WithType(srcMap.value)
	dstEp := destinationEndpoint.WithType(dstMap.value)
	conversion, _ := createTypeConversion(srcEp, dstEp, conversionContext)
	if conversion == nil {
		return nil
	}

	return func(reader dst.Expr, writer func(dst.Expr) []dst.Stmt, generationContext *CodeGenerationContext) []dst.Stmt {
		// We create three obviously related identifiers to use for the conversion
		itemId := sourceEndpoint.CreateLocal("Value")
		keyId := sourceEndpoint.CreateLocal("Key")
		tempId := sourceEndpoint.CreateLocal("Map")

		declaration := astbuilder.SimpleAssignment(
			dst.NewIdent(tempId),
			token.DEFINE,
			astbuilder.MakeMap(dstMap.key.AsType(generationContext), dstMap.value.AsType(generationContext)))

		assignToItem := func(expr dst.Expr) []dst.Stmt {
			return []dst.Stmt{
				astbuilder.SimpleAssignment(
					&dst.IndexExpr{
						X:     dst.NewIdent(tempId),
						Index: dst.NewIdent(keyId),
					},
					token.ASSIGN,
					expr),
			}
		}

		avoidAliasing := astbuilder.SimpleAssignment(dst.NewIdent(itemId), token.DEFINE, dst.NewIdent(itemId))
		avoidAliasing.Decs.Start.Append("// Shadow the loop variable to avoid aliasing")
		avoidAliasing.Decs.Before = dst.NewLine

		loopBody := astbuilder.Statements(
			avoidAliasing,
			conversion(dst.NewIdent(itemId), assignToItem, generationContext))

		assign := writer(dst.NewIdent(tempId))
		loop := astbuilder.IterateOverMapWithValue(keyId, itemId, reader, loopBody...)
		return astbuilder.Statements(declaration, loop, assign)
	}
}

// assignEnumTypeFromEnumType will generate a conversion if both types have the same underlying
// primitive type and neither source nor destination is optional
//
// <local> = <baseType>(<source>)
// <destination> = <enum>(<local>)
//
func assignEnumTypeFromEnumType(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint,
	conversionContext *StorageConversionContext) StorageTypeConversion {

	// Require source to be non-optional
	if _, srcOpt := AsOptionalType(sourceEndpoint.Type()); srcOpt {
		return nil
	}

	// Require destination to be non-optional
	if _, dstOpt := AsOptionalType(destinationEndpoint.Type()); dstOpt {
		return nil
	}

	// Require source to be an enumeration
	_, srcType, ok := conversionContext.ResolveType(sourceEndpoint.Type())
	if !ok {
		return nil
	}
	srcEnum, srcIsEnum := AsEnumType(srcType)
	if !srcIsEnum {
		return nil
	}

	// Require destination to be an enumeration
	dstName, dstType, ok := conversionContext.ResolveType(destinationEndpoint.Type())
	if !ok {
		return nil
	}
	dstEnum, dstIsEnum := AsEnumType(dstType)
	if !dstIsEnum {
		return nil
	}

	// Require enumerations to have the same base types
	if !srcEnum.baseType.Equals(dstEnum.baseType) {
		return nil
	}

	local := destinationEndpoint.CreateLocal("", "Value")
	return func(reader dst.Expr, writer func(dst.Expr) []dst.Stmt, ctx *CodeGenerationContext) []dst.Stmt {
		result := []dst.Stmt{
			astbuilder.SimpleAssignment(
				dst.NewIdent(local),
				token.DEFINE,
				astbuilder.CallFunc(dstName.Name(), reader)),
		}

		result = append(result, writer(dst.NewIdent(local))...)
		return result
	}
}

// assignEnumTypeFromOptionalEnumType will generate a direct assignment if both types have the same
// underlying primitive type and only the source is optional
//
// if <source> != nil {
//    <destination> = <enum>(*<source>)
// } else {
//    <destination> = <zero>
// }
//
func assignEnumTypeFromOptionalEnumType(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint,
	conversionContext *StorageConversionContext) StorageTypeConversion {

	// Require source to be optional
	if _, srcOpt := AsOptionalType(sourceEndpoint.Type()); !srcOpt {
		return nil
	}

	// Require destination to be non-optional
	if _, dstOpt := AsOptionalType(destinationEndpoint.Type()); dstOpt {
		return nil
	}

	// Require source to be an enumeration
	_, srcType, ok := conversionContext.ResolveType(sourceEndpoint.Type())
	if !ok {
		return nil
	}
	srcEnum, srcIsEnum := AsEnumType(srcType)
	if !srcIsEnum {
		return nil
	}

	// Require destination to be an enumeration
	dstName, dstType, ok := conversionContext.ResolveType(destinationEndpoint.Type())
	if !ok {
		return nil
	}
	dstEnum, dstIsEnum := AsEnumType(dstType)
	if !dstIsEnum {
		return nil
	}

	// Require enumerations to have the same base types
	if !srcEnum.baseType.Equals(dstEnum.baseType) {
		return nil
	}

	local := destinationEndpoint.CreateLocal("", "Enum")

	return func(reader dst.Expr, writer func(dst.Expr) []dst.Stmt, ctx *CodeGenerationContext) []dst.Stmt {

		declaration := astbuilder.LocalVariableDeclaration(local, dst.NewIdent(dstName.Name()), "")

		// Need to check for null and only assign if we have a value
		cond := astbuilder.NotEqual(reader, dst.NewIdent("nil"))

		writeLocal := func(expr dst.Expr) []dst.Stmt {
			return []dst.Stmt{
				astbuilder.SimpleAssignment(
					dst.NewIdent(local),
					token.ASSIGN,
					expr),
			}
		}

		updateLocalFromReader := writeLocal(
			astbuilder.CallFunc(dstName.name, astbuilder.Dereference(reader)))

		updateLocalWithZero := writeLocal(
			astbuilder.CallFunc(dstName.name,
				&dst.BasicLit{
					Value: zeroValue(srcEnum.baseType),
				}))

		stmt := &dst.IfStmt{
			Cond: cond,
			Body: astbuilder.StatementBlock(updateLocalFromReader...),
			Else: astbuilder.StatementBlock(updateLocalWithZero...),
		}

		assignValue := writer(dst.NewIdent(local))
		return astbuilder.Statements(declaration, stmt, assignValue)
	}
}

// assignObjectTypeFromObjectType will generate a conversion if both properties are TypeNames
// referencing ObjectType definitions and neither property is optional
//
// For ConvertFrom:
//
// var <local> <destinationType>
// err := <local>.ConvertFrom(<source>)
// if err != nil {
//     return errors.Wrap(err, "while calling <local>.ConvertFrom(<source>)")
// }
// <destination> = <local>
//
// For ConvertTo:
//
// var <local> <destinationType>
// err := <source>.ConvertTo(&<local>)
// if err != nil {
//     return errors.Wrap(err, "while calling <local>.ConvertTo(<source>)")
// }
// <destination> = <local>
//
func assignObjectTypeFromObjectType(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint,
	conversionContext *StorageConversionContext) StorageTypeConversion {

	// Require source to be non-optional
	if _, srcOpt := AsOptionalType(sourceEndpoint.Type()); srcOpt {
		return nil
	}

	// Require destination to be non-optional
	if _, dstOpt := AsOptionalType(destinationEndpoint.Type()); dstOpt {
		return nil
	}

	// Require source to be an object
	_, srcType, ok := conversionContext.ResolveType(sourceEndpoint.Type())
	if !ok {
		return nil
	}
	if _, srcIsObject := AsObjectType(srcType); !srcIsObject {
		return nil
	}

	// Require destination to be an object
	dstName, dstType, ok := conversionContext.ResolveType(destinationEndpoint.Type())
	if !ok {
		return nil
	}
	_, dstIsObject := AsObjectType(dstType)
	if !dstIsObject {
		return nil
	}

	copyVar := destinationEndpoint.CreateLocal()

	return func(reader dst.Expr, writer func(dst.Expr) []dst.Stmt, generationContext *CodeGenerationContext) []dst.Stmt {

		localId := dst.NewIdent(copyVar)
		errLocal := dst.NewIdent("err")

		declaration := astbuilder.LocalVariableDeclaration(copyVar, createTypeDeclaration(dstName, generationContext), "")

		var conversion dst.Stmt
		if dstName.PackageReference.Equals(generationContext.CurrentPackage()) {
			conversion = astbuilder.SimpleAssignment(
				errLocal,
				token.DEFINE,
				astbuilder.CallExpr(localId, conversionContext.functionName, astbuilder.AddrOf(reader)))
		} else {
			conversion = astbuilder.SimpleAssignment(
				errLocal,
				token.DEFINE,
				astbuilder.CallExpr(reader, conversionContext.functionName, localId))
		}

		checkForError := astbuilder.ReturnIfNotNil(
			errLocal,
			astbuilder.WrappedErrorf(
				"populating %s from %s, calling %s()",
				destinationEndpoint.name, sourceEndpoint.name, conversionContext.functionName))

		assignment := writer(dst.NewIdent(copyVar))
		return astbuilder.Statements(declaration, conversion, checkForError, assignment)
	}
}

// assignObjectTypeFromOptionalObjectType will generate a conversion if both properties are
// TypeNames referencing ObjectType definitions and only the source is optional
//
// For ConvertFrom:
//
// var <local> <destinationType>
// if <source> != nil {
//     err := <local>.ConvertFrom(<source>)
//     if err != nil {
//         return errors.Wrap(err, "while calling <local>.ConvertTo(<source>)")
//     }
// }
// <destination> = <local>
//
// For ConvertTo:
//
// var <local> <destinationType>
// if <source> != nil {
//     err := <source>.ConvertTo(&<local>)
//     if err != nil {
//         return errors.Wrap(err, "while calling <local>.ConvertTo(<source>)")
//     }
// }
// <destination> = <local>
//
func assignObjectTypeFromOptionalObjectType(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint,
	conversionContext *StorageConversionContext) StorageTypeConversion {

	// Require source to be optional
	if _, srcOpt := AsOptionalType(sourceEndpoint.Type()); !srcOpt {
		return nil
	}

	// Require destination to be non-optional
	if _, dstOpt := AsOptionalType(destinationEndpoint.Type()); dstOpt {
		return nil
	}

	// Require source to be an object
	_, srcType, ok := conversionContext.ResolveType(sourceEndpoint.Type())
	if !ok {
		return nil
	}
	if _, srcIsObject := AsObjectType(srcType); !srcIsObject {
		return nil
	}

	// Require destination to be an object
	dstName, dstType, ok := conversionContext.ResolveType(destinationEndpoint.Type())
	if !ok {
		return nil
	}
	_, dstIsObject := AsObjectType(dstType)
	if !dstIsObject {
		return nil
	}

	copyVar := destinationEndpoint.CreateLocal()

	return func(reader dst.Expr, writer func(dst.Expr) []dst.Stmt, generationContext *CodeGenerationContext) []dst.Stmt {

		localId := dst.NewIdent(copyVar)
		errLocal := dst.NewIdent("err")

		declaration := astbuilder.LocalVariableDeclaration(copyVar, createTypeDeclaration(dstName, generationContext), "")

		var conversion dst.Stmt
		if dstName.PackageReference.Equals(generationContext.CurrentPackage()) {
			conversion = astbuilder.SimpleAssignment(
				errLocal,
				token.DEFINE,
				astbuilder.CallExpr(localId, conversionContext.functionName, astbuilder.AddrOf(reader)))
		} else {
			conversion = astbuilder.SimpleAssignment(
				errLocal,
				token.DEFINE,
				astbuilder.CallExpr(reader, conversionContext.functionName, localId))
		}

		checkForError := astbuilder.ReturnIfNotNil(
			errLocal,
			astbuilder.WrappedErrorf(
				"populating %s from %s, calling %s()",
				destinationEndpoint.name, sourceEndpoint.name, conversionContext.functionName))

		safeConversion := astbuilder.IfNotNil(reader, conversion, checkForError)

		assignment := writer(dst.NewIdent(copyVar))
		return astbuilder.Statements(declaration, safeConversion, assignment)
	}
}

func createTypeDeclaration(name TypeName, generationContext *CodeGenerationContext) dst.Expr {
	if name.PackageReference.Equals(generationContext.CurrentPackage()) {
		return dst.NewIdent(name.Name())
	}

	packageName := generationContext.MustGetImportedPackageName(name.PackageReference)
	return astbuilder.Selector(dst.NewIdent(packageName), name.Name())
}

func zeroValue(p *PrimitiveType) string {
	switch p {
	case StringType:
		return "\"\""
	case IntType:
		return "0"
	case FloatType:
		return "0"
	case UInt32Type:
		return "0"
	case UInt64Type:
		return "0"
	case BoolType:
		return "false"
	default:
		panic(fmt.Sprintf("unexpected primitive type %q", p.String()))
	}
}
