package astmodel

import (
	"fmt"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astbuilder"
	"github.com/dave/dst"
	"github.com/pkg/errors"
	"go/token"
)

// createPropertyConversion tries to create a property conversion between the two provided properties, using all of the
// available conversion functions in priority order to do so.
func createPropertyConversion(
	sourceProperty *PropertyDefinition,
	destinationProperty *PropertyDefinition,
	knownLocals map[string]struct{}) (StoragePropertyConversion, error) {

	sourceEndpoint := NewStorageConversionEndpoint(
		sourceProperty.propertyType, string(sourceProperty.propertyName), knownLocals)
	destinationEndpoint := NewStorageConversionEndpoint(
		destinationProperty.propertyType, string(destinationProperty.propertyName), knownLocals)

	conversion, err := createTypeConversion(sourceEndpoint, destinationEndpoint)

	if err != nil {
		return nil, errors.Wrapf(
			err,
			"trying to assign %s from %s",
			destinationProperty.propertyName,
			sourceProperty.propertyName)
	}

	return func(source dst.Expr, destination dst.Expr, ctx *CodeGenerationContext) []dst.Stmt {

		var reader = astbuilder.Selector(source, string(sourceProperty.PropertyName()))
		var writer = astbuilder.Selector(destination, string(destinationProperty.PropertyName()))

		return conversion(reader, writer, ctx)
	}, nil
}

// StorageTypeConversion generates the AST for a given conversion.
// source is an expression to read the original value
// destination is an expression to write the converted value
// The parameters source and destination are funcs because AST fragments can't be reused, and in
// some cases we need to reference source and destination multiple times in a single fragment.
type StorageTypeConversion func(reader dst.Expr, writer dst.Expr, ctx *CodeGenerationContext) []dst.Stmt

// StorageTypeConversionFactory represents factory methods that can be used to create StorageTypeConversions
// for a specific pair of types
// source is the endpoint that will be read
// destination is the endpoint that will be written
type StorageTypeConversionFactory func(source *StorageConversionEndpoint, destination *StorageConversionEndpoint) StorageTypeConversion

// A list of all known type conversion factory methods
var typeConversionFactories []StorageTypeConversionFactory

func init() {
	typeConversionFactories = []StorageTypeConversionFactory{
		assignPrimitiveTypeFromPrimitiveType,
		assignOptionalPrimitiveTypeFromPrimitiveType,
		assignPrimitiveTypeFromOptionalPrimitiveType,
		assignArrayFromArray,
		assignMapFromMap,
	}
}

// createTypeConversion tries to create a type conversion between the two provided types, using
// all of the available type conversion functions in priority order to do so.
func createTypeConversion(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint) (StorageTypeConversion, error) {
	for _, f := range typeConversionFactories {
		result := f(sourceEndpoint, destinationEndpoint)
		if result != nil {
			return result, nil
		}
	}

	err := fmt.Errorf(
		"no conversion found to assign %s from %s",
		destinationEndpoint.name,
		sourceEndpoint.name)

	return nil, err
}

// assignPrimitiveTypeFromPrimitiveType will generate a direct assignment if both types have the
// same underlying primitive type and have the same optionality
// <destination> = <source>
func assignPrimitiveTypeFromPrimitiveType(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint) StorageTypeConversion {
	st := AsPrimitiveType(sourceEndpoint.Type())
	dt := AsPrimitiveType(destinationEndpoint.theType)

	if st == nil || dt == nil || !st.Equals(dt) {
		// Either or both sides are not primitive types, or not the same primitive type
		return nil
	}

	if IsOptionalType(sourceEndpoint.Type()) != isTypeOptional(destinationEndpoint.Type()) {
		// Different optionality, handled elsewhere
		return nil
	}

	return func(reader dst.Expr, writer dst.Expr, ctx *CodeGenerationContext) []dst.Stmt {
		return []dst.Stmt{
			astbuilder.SimpleAssignment(writer, token.ASSIGN, reader),
		}
	}
}

// assignOptionalPrimitiveTypeFromPrimitiveType will generate a direct assignment if both types
// have the same underlying primitive type and have the same optionality
// <destination> = &<source>
func assignOptionalPrimitiveTypeFromPrimitiveType(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint) StorageTypeConversion {
	st := AsPrimitiveType(sourceEndpoint.Type())
	dt := AsPrimitiveType(destinationEndpoint.Type())

	if st == nil || dt == nil || !st.Equals(dt) {
		// Either or both sides are not primitive types, or not the same primitive type
		return nil
	}

	if IsOptionalType(sourceEndpoint.Type()) || !IsOptionalType(destinationEndpoint.Type()) {
		// Different optionality than we handle here
		return nil
	}

	return func(reader dst.Expr, writer dst.Expr, ctx *CodeGenerationContext) []dst.Stmt {
		return []dst.Stmt{
			astbuilder.SimpleAssignment(writer, token.ASSIGN, astbuilder.AddrOf(reader)),
		}
	}
}

// assignPrimitiveTypeFromOptionalPrimitiveType will generate a direct assignment if both types
// have the same underlying primitive type and have the same optionality
//
// if <source> != nil {
//    <destination> = *<source>
// } else {
//    <destination> = <zero>
// }
func assignPrimitiveTypeFromOptionalPrimitiveType(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint) StorageTypeConversion {
	st := AsPrimitiveType(sourceEndpoint.Type())
	dt := AsPrimitiveType(destinationEndpoint.Type())

	if st == nil || dt == nil || !st.Equals(dt) {
		// Either or both sides are not primitive types, or not the same primitive type
		return nil
	}

	if !IsOptionalType(sourceEndpoint.Type()) || IsOptionalType(destinationEndpoint.Type()) {
		// Different optionality than we handle here
		return nil
	}

	return func(reader dst.Expr, writer dst.Expr, ctx *CodeGenerationContext) []dst.Stmt {
		// Need to check for null and only assign if we have a value
		cond := astbuilder.NotEqual(reader, dst.NewIdent("nil"))

		assignValue := astbuilder.SimpleAssignment(writer, token.ASSIGN, astbuilder.Dereference(reader))

		assignZero := astbuilder.SimpleAssignment(
			writer,
			token.ASSIGN,
			&dst.BasicLit{
				Value: zeroValue(st),
			})

		stmt := &dst.IfStmt{
			Cond: cond,
			Body: &dst.BlockStmt{
				List: []dst.Stmt{
					assignValue,
				},
			},
			Else: &dst.BlockStmt{
				List: []dst.Stmt{
					assignZero,
				},
			},
		}

		return []dst.Stmt{stmt}
	}
}

// assignArrayFromArray will generate a code fragment to populate an array, assuming the
// underlying types of the two arrays are compatible
//
// var <arr> []<type>
// for _, <value> := range <reader> {
//     arr := append(arr, <value>)
// }
// <writer> = <arr>
func assignArrayFromArray(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint) StorageTypeConversion {
	st := AsArrayType(sourceEndpoint.Type())
	dt := AsArrayType(destinationEndpoint.Type())

	if st == nil || dt == nil {
		// One or other type is not an array
		return nil
	}

	srcEp := sourceEndpoint.WithType(st.element)
	dstEp := destinationEndpoint.WithType(dt.element)
	conversion, _ := createTypeConversion(srcEp, dstEp)

	if conversion == nil {
		// No conversion between the elements of the array
		return nil
	}

	return func(reader dst.Expr, writer dst.Expr, ctx *CodeGenerationContext) []dst.Stmt {
		itemId := sourceEndpoint.CreateSingularLocal()
		tempId := sourceEndpoint.CreatePluralLocal("List")

		declaration := astbuilder.SimpleAssignment(
			tempId,
			token.DEFINE,
			astbuilder.MakeList(dt.AsType(ctx), astbuilder.CallFunc("len", reader)))

		body := conversion(
			itemId,
			&dst.IndexExpr{
				X:     tempId,
				Index: dst.NewIdent("index"),
			},
			ctx,
		)

		assign := astbuilder.SimpleAssignment(writer, token.ASSIGN, tempId)

		loop := astbuilder.IterateOverListWithIndex("index", itemId.Name, reader, body...)
		loop.Decs.After = dst.EmptyLine

		return []dst.Stmt{
			declaration,
			loop,
			assign,
		}
	}
}

// assignMapFromMap will generate a code fragment to populate an array, assuming the
// underlying types of the two arrays are compatible
//
// var <map> map[<key>]<type>
// for key, <item> := range <reader> {
//     <map>[<key>] := <item>
// }
// <writer> = <map>
func assignMapFromMap(
	sourceEndpoint *StorageConversionEndpoint,
	destinationEndpoint *StorageConversionEndpoint) StorageTypeConversion {
	st := AsMapType(sourceEndpoint.Type())
	dt := AsMapType(destinationEndpoint.Type())

	if st == nil || dt == nil {
		// One or other type is not a map
		return nil
	}

	if !st.key.Equals(dt.key) {
		// Keys are different types
		return nil
	}

	srcEp := sourceEndpoint.WithType(st.value)
	dstEp := destinationEndpoint.WithType(dt.value)
	conversion, _ := createTypeConversion(srcEp, dstEp)

	if conversion == nil {
		// No conversion between the elements of the map
		return nil
	}

	return func(reader dst.Expr, writer dst.Expr, ctx *CodeGenerationContext) []dst.Stmt {
		itemId := sourceEndpoint.CreateSingularLocal()
		tempId := sourceEndpoint.CreatePluralLocal("Map")

		declaration := astbuilder.SimpleAssignment(
			tempId,
			token.DEFINE,
			astbuilder.MakeMap(dt.key.AsType(ctx), dt.value.AsType(ctx)))

		body := conversion(
			itemId,
			&dst.IndexExpr{
				X:     tempId,
				Index: dst.NewIdent("key"),
			},
			ctx,
		)

		assign := astbuilder.SimpleAssignment(writer, token.ASSIGN, tempId)

		loop := astbuilder.IterateOverMapWithValue("key", itemId.Name, reader, body...)
		loop.Decs.After = dst.EmptyLine

		return []dst.Stmt{
			declaration,
			loop,
			assign,
		}
	}
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
	}

	return "##DOESNOTCOMPUTE##"
}
