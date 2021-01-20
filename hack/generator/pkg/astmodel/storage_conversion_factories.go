package astmodel

import (
	"fmt"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astbuilder"
	"github.com/dave/dst"
	"go/token"
)

/*
 * These are the factory methods for creating property conversions for storage conversions
 */

var conversionFactories = []StoragePropertyConversionFactory{
	PrimitivePropertyConversionFactory,
	OptionalPrimitivePropertyConversionFactory,
}

// createPropertyConversion tries to create a property conversion between the two provided properties, using all of the
// available conversion functions in priority order to do so.
func createPropertyConversion(source *PropertyDefinition, destination *PropertyDefinition) (StoragePropertyConversion, error) {
	for _, f := range conversionFactories {
		result := f(source, destination)
		if result != nil {
			return result, nil
		}
	}

	err := fmt.Errorf(
		"no converstion found to assign %s (%v) from %s (%v)",
		destination.propertyName,
		destination.propertyType,
		source.propertyName,
		source.propertyType)

	return nil, err
}

// PrimitivePropertyConversionFactory generates a conversion for identical primitive types
func PrimitivePropertyConversionFactory(sourceProperty *PropertyDefinition, destinationProperty *PropertyDefinition) StoragePropertyConversion {
	if IsOptionalType(sourceProperty.propertyType) || IsOptionalType(destinationProperty.propertyType) {
		// We don't handle optional types here
		return nil
	}

	sourceType := AsPrimitiveType(sourceProperty.propertyType)
	destinationType := AsPrimitiveType(sourceProperty.propertyType)
	if sourceType == nil || !sourceType.Equals(destinationType) {
		return nil
	}

	// Both properties have the same underlying primitive type, generate a simple assignment
	return func(source func() dst.Expr, destination func() dst.Expr, ctx *CodeGenerationContext) []dst.Stmt {
		left := &dst.SelectorExpr{
			X:   destination(),
			Sel: dst.NewIdent(string(destinationProperty.propertyName)),
		}
		right := &dst.SelectorExpr{
			X:   source(),
			Sel: dst.NewIdent(string(sourceProperty.propertyName)),
		}
		return []dst.Stmt{
			astbuilder.SimpleAssignment(left, token.ASSIGN, right),
		}
	}
}

func OptionalPrimitivePropertyConversionFactory(source *PropertyDefinition, destination *PropertyDefinition) StoragePropertyConversion {
	sourceOptional := IsOptionalType(source.propertyType)
	destinationOptional := IsOptionalType(destination.propertyType)
	if !sourceOptional && !destinationOptional {
		// Neither side is optional, we don't handle it
		return nil
	}

	sourceType := AsPrimitiveType(source.propertyType)
	destinationType := AsPrimitiveType(source.propertyType)
	if sourceType == nil || !sourceType.Equals(destinationType) {
		return nil
	}

	// Both properties have the same underlying primitive type, but one or other or both is optional
	return func(sourceVariable func() dst.Expr, destinationVariable func() dst.Expr, ctx *CodeGenerationContext) []dst.Stmt {
		if sourceOptional == destinationOptional {
			// Can just copy a pointer to a primitive value
			assign := astbuilder.SimpleAssignment(
				astbuilder.Selector(destinationVariable(), string(destination.propertyName)),
				token.ASSIGN,
				astbuilder.Selector(sourceVariable(), string(source.propertyName)))
			return []dst.Stmt{assign}
		}

		if destinationOptional {
			// Need a pointer to the primitive value as the source is not optional
			assign := astbuilder.SimpleAssignment(
				astbuilder.Selector(destinationVariable(), string(destination.propertyName)),
				token.ASSIGN,
				astbuilder.AddrOf(
					astbuilder.Selector(sourceVariable(), string(source.propertyName))))
			return []dst.Stmt{assign}
		}

		if sourceOptional {
			// Need to check for null and only assign if we have a value
			cond := &dst.BinaryExpr{
				X:  astbuilder.Selector(sourceVariable(), string(source.propertyName)),
				Op: token.NEQ,
				Y:  dst.NewIdent("nil"),
			}
			assignValue := astbuilder.SimpleAssignment(
				astbuilder.Selector(destinationVariable(), string(destination.propertyName)),
				token.ASSIGN,
				astbuilder.Dereference(
					astbuilder.Selector(sourceVariable(), string(source.propertyName))))
			assignZero := astbuilder.SimpleAssignment(
				astbuilder.Selector(destinationVariable(), string(destination.propertyName)),
				token.ASSIGN,
				&dst.BasicLit{
					Value: zeroValue(sourceType),
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

		panic("Should never get to the end of OptionalPrimitivePropertyConversionFactory")
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
