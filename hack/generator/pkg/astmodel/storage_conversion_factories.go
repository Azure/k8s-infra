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
// reader is an expression to read the original value
// writer is a function that creates one or more statements to write the converted value
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
		// Primitive types
		assignPrimitiveTypeFromPrimitiveType,
		assignOptionalPrimitiveTypeFromPrimitiveType,
		assignPrimitiveTypeFromOptionalPrimitiveType,
		assignOptionalPrimitiveTypeFromOptionalPrimitiveType,
		// Collection Types
		assignArrayFromArray,
		assignMapFromMap,
		// Enumerations
		assignEnumTypeFromEnumType,
		assignEnumTypeFromOptionalEnumType,
		assignOptionalEnumTypeFromEnumType,
		assignOptionalEnumTypeFromOptionalEnumType,
		// Complex object types
		assignObjectTypeFromObjectType,
		assignObjectTypeFromOptionalObjectType,
		assignOptionalObjectTypeFromObjectType,
		assignOptionalObjectTypeFromOptionalObjectType,
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

	return func(reader dst.Expr, writer func(dst.Expr) []dst.Stmt, generationContext *CodeGenerationContext) []dst.Stmt {
		// can write a primitive value directly
		return writer(reader)
	}
}

// assignOptionalPrimitiveTypeFromPrimitiveType will generate a direct assignment if both types
// have the same underlying primitive type and only the destination is optional.
//
// <destination> = &<source>
//
func assignOptionalPrimitiveTypeFromPrimitiveType(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint,
	_ *StorageConversionContext) StorageTypeConversion {

	// Require source to be non optional
	if _, srcOpt := AsOptionalType(sourceEndpoint.Type()); srcOpt {
		return nil
	}

	// Require source to be a primitive type
	srcPrim, srcOk := AsPrimitiveType(sourceEndpoint.Type())
	if !srcOk {
		// Source is not a primitive type
		return nil
	}

	// Require destination to be optional
	if _, dstOpt := AsOptionalType(destinationEndpoint.Type()); !dstOpt {
		// Destination is not optional
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

	copyVar := destinationEndpoint.CreateLocal("", "Value")

	return func(reader dst.Expr, writer func(dst.Expr) []dst.Stmt, generationContext *CodeGenerationContext) []dst.Stmt {
		result := []dst.Stmt{
			// Stash the value in a local just in case the original gets modified later on
			astbuilder.SimpleAssignment(dst.NewIdent(copyVar), token.DEFINE, reader),
		}

		result = append(result, writer(astbuilder.AddrOf(dst.NewIdent(copyVar)))...)
		return result
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

	return func(reader dst.Expr, writer func(dst.Expr) []dst.Stmt, generationContext *CodeGenerationContext) []dst.Stmt {
		// Need to check for null and only assign if we have a value
		cond := astbuilder.NotEqual(reader, dst.NewIdent("nil"))

		assignValue := writer(astbuilder.Dereference(reader))

		assignZero := writer(&dst.BasicLit{
			Value: zeroValue(srcPrim),
		})

		stmt := astbuilder.SimpleIfElse(
			cond,
			astbuilder.StatementBlock(assignValue...),
			astbuilder.StatementBlock(assignZero...))

		return []dst.Stmt{stmt}
	}
}

// assignOptionalPrimitiveTypeFromOptionalPrimitiveType will generate a direct assignment if both types have the
// same underlying primitive type and both are optional
//
// <destination> = <source>
//
func assignOptionalPrimitiveTypeFromOptionalPrimitiveType(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint,
	_ *StorageConversionContext) StorageTypeConversion {

	// Require source to be optional
	if _, srcOpt := AsOptionalType(sourceEndpoint.Type()); !srcOpt {
		// Source is not optional
		return nil
	}

	// Require source to be a primitive type
	srcPrim, srcOk := AsPrimitiveType(sourceEndpoint.Type())
	if !srcOk {
		return nil
	}

	// Require destination to be optional
	if _, dstOpt := AsOptionalType(destinationEndpoint.Type()); !dstOpt {
		// Destination is not optional
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

	copyVar := destinationEndpoint.CreateLocal("", "Value")

	return func(reader dst.Expr, writer func(dst.Expr) []dst.Stmt, ctx *CodeGenerationContext) []dst.Stmt {

		// Need to check for null and only assign if we have a value
		cond := astbuilder.NotEqual(reader, dst.NewIdent("nil"))

		readValue := astbuilder.SimpleAssignment(
			dst.NewIdent(copyVar),
			token.DEFINE,
			astbuilder.Dereference(reader))
		readValue.Decs.Start.Append("// Copy to a local to avoid aliasing")
		readValue.Decs.Before = dst.NewLine

		writeValue := writer(astbuilder.AddrOf(dst.NewIdent(copyVar)))
		writeNil := writer(dst.NewIdent("nil"))

		body := []dst.Stmt{
			readValue,
		}

		body = append(body, writeValue...)

		stmt := &dst.IfStmt{
			Cond: cond,
			Body: astbuilder.StatementBlock(body...),
			Else: astbuilder.StatementBlock(writeNil...),
		}

		return []dst.Stmt{
			stmt,
		}
	}
}

// assignArrayFromArray will generate a code fragment to populate an array, assuming the
// underlying types of the two arrays are compatible
//
// <arr> := make([]<type>, len(<reader>))
// for <index>, <value> := range <reader> {
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

		assignToElement := func(expr dst.Expr) []dst.Stmt {
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

		loopBody := conversion(
			dst.NewIdent(itemId),
			assignToElement,
			generationContext)

		assign := writer(dst.NewIdent(tempId))

		loop := astbuilder.IterateOverListWithIndex(indexId, itemId, reader, loopBody...)

		result := []dst.Stmt{
			declaration,
			loop,
		}

		result = append(result, assign...)
		return result
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

		loopBody := conversion(
			dst.NewIdent(itemId),
			assignToItem,
			generationContext,
		)

		assign := writer(dst.NewIdent(tempId))

		loop := astbuilder.IterateOverMapWithValue(keyId, itemId, reader, loopBody...)

		result := []dst.Stmt{
			declaration,
			loop,
		}

		result = append(result, assign...)
		return result
	}
}

// assignEnumTypeFromEnumType will generate a direct assignment if both types have the same
// underlying primitive type and neither source nor destination is optional
//
// <destination> = <enum>(<source>)
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

	return func(reader dst.Expr, writer func(dst.Expr) []dst.Stmt, ctx *CodeGenerationContext) []dst.Stmt {
		return writer(astbuilder.CallFunc(dstName.name, reader))
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
		// Need to check for null and only assign if we have a value
		cond := astbuilder.NotEqual(reader, dst.NewIdent("nil"))

		readValue := astbuilder.SimpleAssignment(
			dst.NewIdent(local),
			token.DEFINE,
			astbuilder.CallFunc(dstName.name, astbuilder.Dereference(reader)))

		writeValue := writer(dst.NewIdent(local))

		assignZero := writer(astbuilder.CallFunc(dstName.name,
			&dst.BasicLit{
				Value: zeroValue(srcEnum.baseType),
			}))

		body := []dst.Stmt{
			readValue,
		}
		body = append(body, writeValue...)

		stmt := &dst.IfStmt{
			Cond: cond,
			Body: astbuilder.StatementBlock(body...),
			Else: astbuilder.StatementBlock(assignZero...),
		}

		return []dst.Stmt{stmt}
	}
}

// assignOptionalEnumTypeFromEnumType will generate a direct assignment if both types have the same
// underlying primitive type and only the destination is optional
//
// <destination> = <enum>(<source>)
//
func assignOptionalEnumTypeFromEnumType(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint,
	conversionContext *StorageConversionContext) StorageTypeConversion {

	// Require source to be non-optional
	if _, srcOpt := AsOptionalType(sourceEndpoint.Type()); srcOpt {
		return nil
	}

	// Require destination to be optional
	if _, dstOpt := AsOptionalType(destinationEndpoint.Type()); !dstOpt {
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

	// Require both enumerations to have the same underlying type
	if !srcEnum.baseType.Equals(dstEnum.baseType) {
		return nil
	}

	copyVar := destinationEndpoint.CreateLocal("", "Enum")

	return func(reader dst.Expr, writer func(dst.Expr) []dst.Stmt, ctx *CodeGenerationContext) []dst.Stmt {
		result := []dst.Stmt{
			astbuilder.SimpleAssignment(
				dst.NewIdent(copyVar),
				token.DEFINE,
				astbuilder.CallFunc(dstName.name, reader)),
		}

		result = append(result, writer(astbuilder.AddrOf(dst.NewIdent(copyVar)))...)
		return result
	}
}

// assignOptionalEnumTypeFromOptionalEnumType will generate a direct assignment if both types have
// the same underlying primitive type and are both optional
//
// <local> = <enum>(*<source>)
// <destination> = &<local>
//
func assignOptionalEnumTypeFromOptionalEnumType(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint,
	conversionContext *StorageConversionContext) StorageTypeConversion {

	// Require source to be optional
	if _, srcOpt := AsOptionalType(sourceEndpoint.Type()); !srcOpt {
		return nil
	}

	// Require destination to be optional
	if _, dstOpt := AsOptionalType(destinationEndpoint.Type()); !dstOpt {
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

	// Require both enumerations to have the same underlying type
	if !srcEnum.baseType.Equals(dstEnum.baseType) {
		return nil
	}

	copyVar := destinationEndpoint.CreateLocal("", "Enum")

	return func(reader dst.Expr, writer func(dst.Expr) []dst.Stmt, ctx *CodeGenerationContext) []dst.Stmt {

		// Need to check for null and only assign if we have a value
		cond := astbuilder.NotEqual(reader, dst.NewIdent("nil"))

		readValue := astbuilder.SimpleAssignment(
			dst.NewIdent(copyVar),
			token.DEFINE,
			astbuilder.CallFunc(dstName.name,
				astbuilder.Dereference(reader)))
		readValue.Decs.Start.Append("// Copy to a local to avoid aliasing")
		readValue.Decs.Before = dst.NewLine

		writeValue := writer(astbuilder.AddrOf(dst.NewIdent(copyVar)))

		assignZero := writer(astbuilder.CallFunc(
			dstName.name,
			&dst.BasicLit{
				Value: zeroValue(dstEnum.baseType),
			}))

		body := []dst.Stmt{
			readValue,
		}
		body = append(body, writeValue...)

		stmt := &dst.IfStmt{
			Cond: cond,
			Body: astbuilder.StatementBlock(body...),
			Else: astbuilder.StatementBlock(assignZero...),
		}

		return []dst.Stmt{
			stmt,
		}
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
				astbuilder.CallExpr(localId, conversionContext.functionName, reader))
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

		result := []dst.Stmt{
			declaration,
			conversion,
			checkForError,
		}

		result = append(result, assignment...)
		return result
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
				astbuilder.CallExpr(localId, conversionContext.functionName, reader))
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

		result := []dst.Stmt{
			declaration,
			safeConversion,
		}

		result = append(result, assignment...)
		return result
	}
}

// assignOptionalObjectTypeFromObjectType will generate a conversion if both properties are
// TypeNames referencing ObjectType definitions and only the destination is optional
//
// For ConvertFrom:
//
// var <local> <destinationType>
// err := <local>.ConvertFrom(&<source>)
// if err != nil {
//    return errors.Wrap(err, <message>)
// }
// <destination> = &<local>
//
// For ConvertTo:
//
// var <local> <destinationType>
// err := <source>.ConvertTo(&<local>)
// if err != nil {
//     return errors.Wrap(err, <message>)
// }
// <destination> = &<local>
//
func assignOptionalObjectTypeFromObjectType(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint,
	conversionContext *StorageConversionContext) StorageTypeConversion {

	// Require source to be non-optional
	if _, srcOpt := AsOptionalType(sourceEndpoint.Type()); srcOpt {
		return nil
	}

	// Require destination to be optional
	if _, dstOpt := AsOptionalType(destinationEndpoint.Type()); !dstOpt {
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
				astbuilder.CallExpr(reader, conversionContext.functionName, astbuilder.AddrOf(localId)))
		}

		checkForError := astbuilder.ReturnIfNotNil(
			errLocal,
			astbuilder.WrappedErrorf(
				"populating %s from %s, calling %s()",
				destinationEndpoint.name, sourceEndpoint.name, conversionContext.functionName))

		assignment := writer(astbuilder.AddrOf(dst.NewIdent(copyVar)))

		result := []dst.Stmt{
			declaration,
			conversion,
			checkForError,
		}

		result = append(result, assignment...)
		return result
	}
}

// assignOptionalObjectTypeFromOptionalObjectType will generate a conversion if both properties are
// TypeNames referencing ObjectType definitions and both source and destination are optional
//
// For ConvertFrom:
//
// if <source> != nil {
//     var <local> <destinationType>
//     err := <local>.ConvertFrom(<source>)
//     if err != nil {
//         return errors.Wrap(err, <message>)
//     }
//     <destination> = &<local>
// } else {
//     <destination> = nil
// }
//
// For ConvertTo:
//
// if <source> != nil {
//     var <local> <destinationType>
//     err := <source>.ConvertTo(<local>)
//     if err != nil {
//         return errors.Wrap(err, <message>)
//     }
//     <destination> = &<local>
// } else {
//     <destination> = nil
// }
//
func assignOptionalObjectTypeFromOptionalObjectType(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint,
	conversionContext *StorageConversionContext) StorageTypeConversion {

	// Require source to be optional
	if _, srcOpt := AsOptionalType(sourceEndpoint.Type()); !srcOpt {
		return nil
	}

	// Require destination to be optional
	if _, dstOpt := AsOptionalType(destinationEndpoint.Type()); !dstOpt {
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

		declaration := astbuilder.LocalVariableDeclaration(
			copyVar,
			createTypeDeclaration(dstName, generationContext), "")

		var conversion dst.Stmt
		if dstName.PackageReference.Equals(generationContext.CurrentPackage()) {
			conversion = astbuilder.SimpleAssignment(
				dst.NewIdent("err"),
				token.DEFINE,
				astbuilder.CallExpr(dst.NewIdent(copyVar), conversionContext.functionName, reader))
		} else {
			conversion = astbuilder.SimpleAssignment(
				dst.NewIdent("err"),
				token.DEFINE,
				astbuilder.CallExpr(reader, conversionContext.functionName, dst.NewIdent(copyVar)))
		}

		checkForError := astbuilder.ReturnIfNotNil(
			dst.NewIdent("err"),
			astbuilder.WrappedErrorf(
				"populating %s from %s, calling %s()",
				destinationEndpoint.name, sourceEndpoint.name, conversionContext.functionName))

		assignment := writer(astbuilder.AddrOf(dst.NewIdent(copyVar)))

		assignNil := writer(dst.NewIdent("nil"))

		body := []dst.Stmt{
			declaration,
			conversion,
			checkForError,
		}
		body = append(body, assignment...)

		stmt := astbuilder.SimpleIfElse(
			astbuilder.NotEqual(reader, dst.NewIdent("nil")),
			astbuilder.StatementBlock(body...),
			astbuilder.StatementBlock(assignNil...))

		return []dst.Stmt{
			stmt,
		}
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
