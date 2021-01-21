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
func createPropertyConversion(sourceProperty *PropertyDefinition, destinationProperty *PropertyDefinition) (StoragePropertyConversion, error) {

	conversion, err := createTypeConversion(sourceProperty.propertyType, destinationProperty.propertyType)

	if err != nil {
		return nil, errors.Wrapf(
			err,
			"trying to assign %s from %s",
			destinationProperty.propertyName,
			sourceProperty.propertyName)
	}

	return func(source func() dst.Expr, destination func() dst.Expr, ctx *CodeGenerationContext) []dst.Stmt {

		var reader = func() dst.Expr {
			return astbuilder.Selector(source(), string(sourceProperty.PropertyName()))
		}

		var writer = func() dst.Expr {
			return astbuilder.Selector(destination(), string(destinationProperty.PropertyName()))
		}

		return conversion(reader, writer)
	}, nil
}

// StorageTypeConversion generates the AST for a given conversion.
// source is a factory function that returns an expression to read the original value
// destination is a factory function that returns an expression to write the converted value
// The parameters source and destination are funcs because AST fragments can't be reused, and in
// some cases we need to reference source and destination multiple times in a single fragment.
type StorageTypeConversion func(reader func() dst.Expr, writer func() dst.Expr) []dst.Stmt

// StorageTypeConversionFactory represents factory methods that can be used to create StorageTypeConversions
// for a specific pair of types
// sourceType is the type of the value that will be read
// destinationType is the type of the value that will be written
type StorageTypeConversionFactory func(sourceType Type, destinationType Type) StorageTypeConversion

// A list of all known type conversion factory methods
var typeConversionFactories = []StorageTypeConversionFactory{
	assignPrimitiveTypeFromPrimitiveType,
	assignOptionalPrimitiveTypeFromPrimitiveType,
	assignPrimitiveTypeFromOptionalPrimitiveType,
}

// createTypeConversion tries to create a type conversion between the two provided types, using
// all of the available type conversion functions in priority order to do so.
func createTypeConversion(sourceType Type, destinationType Type) (StorageTypeConversion, error) {
	for _, f := range typeConversionFactories {
		result := f(sourceType, destinationType)
		if result != nil {
			return result, nil
		}
	}

	err := fmt.Errorf(
		"no conversion found to assign %s from %s",
		destinationType,
		sourceType)

	return nil, err
}

// assignPrimitiveTypeFromPrimitiveType will generate a direct assignment if both types have the
// same underlying primitive type and have the same optionality
// <destination> = <source>
func assignPrimitiveTypeFromPrimitiveType(sourceType Type, destinationType Type) StorageTypeConversion {
	st := AsPrimitiveType(sourceType)
	dt := AsPrimitiveType(destinationType)

	if st == nil || dt == nil || !st.Equals(dt) {
		// Either or both sides are not primitive types, or not the same primitive type
		return nil
	}

	if IsOptionalType(sourceType) != isTypeOptional(destinationType) {
		// Different optionality, handled elsewhere
		return nil
	}

	return func(reader func() dst.Expr, writer func() dst.Expr) []dst.Stmt {
		return []dst.Stmt{
			astbuilder.SimpleAssignment(writer(), token.ASSIGN, reader()),
		}
	}
}

// assignOptionalPrimitiveTypeFromPrimitiveType will generate a direct assignment if both types
// have the same underlying primitive type and have the same optionality
// <destination> = &<source>
func assignOptionalPrimitiveTypeFromPrimitiveType(sourceType Type, destinationType Type) StorageTypeConversion {
	st := AsPrimitiveType(sourceType)
	dt := AsPrimitiveType(destinationType)

	if st == nil || dt == nil || !st.Equals(dt) {
		// Either or both sides are not primitive types, or not the same primitive type
		return nil
	}

	if IsOptionalType(sourceType) || !IsOptionalType(destinationType) {
		// Different optionality than we handle here
		return nil
	}

	return func(reader func() dst.Expr, writer func() dst.Expr) []dst.Stmt {
		return []dst.Stmt{
			astbuilder.SimpleAssignment(writer(), token.ASSIGN, astbuilder.AddrOf(reader())),
		}
	}
}

// assignPrimitiveTypeFromOptionalPrimitiveType will generate a direct assignment if both types
// have the same underlying primitive type and have the same optionality
// if <source> != nil {
//    <destination> = *<source>
// } else {
//    <destination> = <zero>
// }
func assignPrimitiveTypeFromOptionalPrimitiveType(sourceType Type, destinationType Type) StorageTypeConversion {
	st := AsPrimitiveType(sourceType)
	dt := AsPrimitiveType(destinationType)

	if st == nil || dt == nil || !st.Equals(dt) {
		// Either or both sides are not primitive types, or not the same primitive type
		return nil
	}

	if !IsOptionalType(sourceType) || IsOptionalType(destinationType) {
		// Different optionality than we handle here
		return nil
	}

	return func(reader func() dst.Expr, writer func() dst.Expr) []dst.Stmt {
		// Need to check for null and only assign if we have a value
		cond := &dst.BinaryExpr{
			X:  reader(),
			Op: token.NEQ,
			Y:  dst.NewIdent("nil"),
		}

		assignValue := astbuilder.SimpleAssignment(
			writer(),
			token.ASSIGN,
			astbuilder.Dereference(reader()))

		assignZero := astbuilder.SimpleAssignment(
			writer(),
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
