/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astbuilder"
	ast "github.com/dave/dst"
	"go/token"
	"sort"
)

type StoragePropertyConversion func(sourceVariable string, destinationVariable string, ctx *CodeGenerationContext) []ast.Stmt

type StoragePropertyConversionFactory func(source *PropertyDefinition, destination *PropertyDefinition) StoragePropertyConversion

// Represents a function that performs conversions for storage versions
type StorageConversionFunction struct {
	// Name of this conversion function
	name string
	// Name of the ultimate hub type to which we are converting, passed as a parameter
	parameter TypeName
	// Other Type of this conversion stage (if this isn't our parameter type, we'll delegate to this to finish the job)
	staging TypeDefinition
	// Map of all property conversions we are going to use
	conversions map[string]StoragePropertyConversion
	// Reference to our identifier factory
	idFactory IdentifierFactory
	// Which conversionType of conversion are we generating?
	conversionDirection StorageConversionDirection
}

// Direction of conversion we're implementing with this function
type StorageConversionDirection int

const (
	// Indicates the conversion is from the passed other instance, populating the receiver
	ConvertFrom = StorageConversionDirection(1)
	// Indicate the conversion is to the passed other type, populating other
	ConvertTo = StorageConversionDirection(2)
)

var _ Function = &StorageConversionFunction{}

func NewStorageConversionFromFunction(
	receiver TypeDefinition,
	source TypeName,
	staging TypeDefinition,
	idFactory IdentifierFactory,
) *StorageConversionFunction {
	result := &StorageConversionFunction{
		name:                "ConvertFrom",
		parameter:           source,
		staging:             staging,
		idFactory:           idFactory,
		conversionDirection: ConvertFrom,
		conversions:         make(map[string]StoragePropertyConversion),
	}

	result.createConversions(receiver)
	return result
}

func NewStorageConversionToFunction(
	receiver TypeDefinition,
	destination TypeName,
	staging TypeDefinition,
	idFactory IdentifierFactory,
) *StorageConversionFunction {
	result := &StorageConversionFunction{
		name:                "ConvertTo",
		parameter:           destination,
		staging:             staging,
		idFactory:           idFactory,
		conversionDirection: ConvertTo,
		conversions:         make(map[string]StoragePropertyConversion),
	}

	result.createConversions(receiver)
	return result
}

func (fn *StorageConversionFunction) Name() string {
	return fn.name
}

func (fn *StorageConversionFunction) RequiredPackageReferences() *PackageReferenceSet {
	return NewPackageReferenceSet(
		fn.parameter.PackageReference,
		fn.staging.name.PackageReference)
}

func (fn *StorageConversionFunction) References() TypeNameSet {
	return NewTypeNameSet(fn.parameter, fn.staging.name)
}

func (fn *StorageConversionFunction) Equals(f Function) bool {
	if other, ok := f.(*StorageConversionFunction); ok {
		// Only check name for now
		if fn.name != other.name {
			return false
		}
	}

	return false
}

func (fn *StorageConversionFunction) AsFunc(ctx *CodeGenerationContext, receiver TypeName) *ast.FuncDecl {

	parameterName := fn.parameterName()
	receiverName := fn.receiverName(receiver)

	funcDetails := &astbuilder.FuncDetails{
		ReceiverIdent: receiverName,
		ReceiverType:  receiver.AsType(ctx),
		Name:          fn.Name(),
		Body:          fn.generateBody(receiverName, parameterName, ctx),
	}

	funcDetails.AddParameter(
		parameterName,
		&ast.StarExpr{
			X: fn.staging.name.AsType(ctx)})
	funcDetails.AddReturns("error")

	return funcDetails.DefineFunc()
}

// generateBody returns all of the statements required for the conversion function
// receiverName is the name of our receiver type, used to qualify field access
// parameterName is the name of the parameter passed to the function, also used for field access
// ctx is our code generation context, passed to allow resolving of identifiers in other packages
func (fn *StorageConversionFunction) generateBody(receiverName string, parameterName string, ctx *CodeGenerationContext) []ast.Stmt {

	if fn.parameter.Equals(fn.staging.name) {
		// Last step of conversion, directly to the parameter type we've been given
		if fn.conversionDirection == ConvertFrom {
			return fn.generateDirectConversionFrom(receiverName, parameterName, ctx)
		} else {
			// fn.conversionType == ConvertTo
			return fn.generateDirectConversionTo(receiverName, parameterName, ctx)
		}
	}

	// Intermediate step of conversion, not working directly with the parameter type we've been given
	if fn.conversionDirection == ConvertFrom {
		return fn.generateIndirectConversionFrom(receiverName, parameterName, ctx)
	} else {
		// fn.conversionType == ConvertTo
		return fn.generateIndirectConversionTo(receiverName, parameterName, ctx)
	}
}

// generateDirectConversionFrom returns the method body required to directly copy information from
// the parameter instance onto our receiver
func (fn *StorageConversionFunction) generateDirectConversionFrom(receiverName string, parameterName string, ctx *CodeGenerationContext) []ast.Stmt {
	return fn.generateAssignments(parameterName, receiverName, ctx)
}

// generateDirectConversionTo returns the method body required to directly copy information from
// our receiver onto the parameter instance
func (fn *StorageConversionFunction) generateDirectConversionTo(receiverName string, parameterName string, ctx *CodeGenerationContext) []ast.Stmt {
	return fn.generateAssignments(receiverName, parameterName, ctx)
}

// generateIndirectConversionFrom returns the method body required to populate our receiver when
// we don't directly understand the structure of the parameter value.
// To accommodate this, we first convert to an intermediate form:
//
// var staging IntermediateType
// staging.ConvertFrom(parameter)
// [copy values from staging]
//
func (fn *StorageConversionFunction) generateIndirectConversionFrom(receiverName string, parameterName string, ctx *CodeGenerationContext) []ast.Stmt {
	staging := astbuilder.LocalVariableDeclaration(
		"staging", ast.NewIdent(fn.staging.name.name), "// staging is our intermediate type for conversion")
	staging.Decorations().Before = ast.NewLine

	convertFrom := astbuilder.InvokeQualifiedFunc(
		"staging", fn.name, ast.NewIdent(parameterName))
	convertFrom.Decorations().Before = ast.EmptyLine
	convertFrom.Decorations().Start.Append("// populate staging from " + parameterName)

	assignments := fn.generateAssignments("staging", receiverName, ctx)

	var result []ast.Stmt
	result = append(result, staging)
	result = append(result, convertFrom)
	result = append(result, assignments...)
	return result
}

// generateIndirectConversionTo returns the method body required to populate our parameter
// instance when we don't directly understand the structure of the parameter value.
// To accommodate this, we first populate an intermediate form that is then converted.
//
// var staging IntermediateType
// [copy values to staging]
// staging.ConvertTo(parameter)
//
func (fn *StorageConversionFunction) generateIndirectConversionTo(receiverName string, parameterName string, ctx *CodeGenerationContext) []ast.Stmt {
	staging := astbuilder.LocalVariableDeclaration(
		"staging", ast.NewIdent(fn.staging.name.name), "// staging is our intermediate type for conversion")
	staging.Decorations().Before = ast.NewLine

	convertTo := astbuilder.InvokeQualifiedFunc(
		"staging", fn.name, ast.NewIdent(parameterName))
	convertTo.Decorations().Before = ast.EmptyLine
	convertTo.Decorations().Start.Append("// use staging to populate " + parameterName)

	assignments := fn.generateAssignments(receiverName, "staging", ctx)

	var result []ast.Stmt
	result = append(result, staging)
	result = append(result, assignments...)
	result = append(result, convertTo)
	return result
}

func (fn *StorageConversionFunction) generateAssignments(source string, destination string, ctx *CodeGenerationContext) []ast.Stmt {
	var result []ast.Stmt

	// Find all the properties for which we have a conversion
	var properties []string
	for p := range fn.conversions {
		properties = append(properties, p)
	}

	// Sort the properties into alphabetical order to ensure deterministic generation
	sort.Slice(properties, func(i, j int) bool {
		return properties[i] < properties[j]
	})

	// Accumulate all the statements required for conversions, in alphabetical order
	for _, prop := range properties {
		conversion := fn.conversions[prop]
		block := conversion(source, destination, ctx)
		if len(block) > 0 {
			//TODO: Tidy
			firstStatement := block[0]
			firstStatement.Decorations().Before = ast.EmptyLine
			firstStatement.Decorations().Start.Append("// " + prop)
			result = append(result, block...)
		}
	}

	return result
}

func (fn *StorageConversionFunction) receiverName(receiver TypeName) string {
	return fn.idFactory.CreateIdentifier(receiver.Name(), NotExported)
}

func (fn *StorageConversionFunction) parameterName() string {
	if fn.conversionDirection == ConvertTo {
		return "destination"
	}

	if fn.conversionDirection == ConvertFrom {
		return "source"
	}

	panic(fmt.Sprintf("Unexpected conversion conversionType %v", fn.conversionDirection))
}

func (fn *StorageConversionFunction) createConversions(receiver TypeDefinition) {
	receiverObject := fn.unwrapObject(receiver.Type())
	otherObject := fn.unwrapObject(fn.staging.Type())

	for _, receiverProperty := range receiverObject.Properties() {
		otherProperty, ok := otherObject.Property(receiverProperty.propertyName)
		//TODO: Handle renames
		if ok {
			var conv StoragePropertyConversion
			if fn.conversionDirection == ConvertFrom {
				conv = createPropertyConversion(otherProperty, receiverProperty)
			} else {
				conv = createPropertyConversion(receiverProperty, otherProperty)
			}

			fn.conversions[string(receiverProperty.propertyName)] = conv
		}
	}
}

func (fn *StorageConversionFunction) unwrapObject(aType Type) *ObjectType {
	switch t := aType.(type) {
	case *ObjectType:
		return t

	case *FlaggedType:
		return fn.unwrapObject(t.element)

	case *ErroredType:
		return fn.unwrapObject(t.inner)

	default:
		return nil
	}
}

var conversionFactories = []StoragePropertyConversionFactory{
	PrimitivePropertyConversionFactory,
	OptionalPrimitivePropertyConversionFactory,
}

func createPropertyConversion(source *PropertyDefinition, destination *PropertyDefinition) StoragePropertyConversion {
	for _, f := range conversionFactories {
		result := f(source, destination)
		if result != nil {
			return result
		}
	}

	return nil
}

// PrimitivePropertyConversionFactory generates a conversion for identical primitive types
func PrimitivePropertyConversionFactory(source *PropertyDefinition, destination *PropertyDefinition) StoragePropertyConversion {
	if IsOptionalType(source.propertyType) || IsOptionalType(destination.propertyType) {
		// We don't handle optional types here
		return nil
	}

	sourceType := AsPrimitiveType(source.propertyType)
	destinationType := AsPrimitiveType(source.propertyType)
	if sourceType == nil || !sourceType.Equals(destinationType) {
		return nil
	}

	// Both properties have the same underlying primitive type, generate a simple assignment
	return func(sourceVariable string, destinationVariable string, _ *CodeGenerationContext) []ast.Stmt {
		left := &ast.SelectorExpr{
			X:   ast.NewIdent(destinationVariable),
			Sel: ast.NewIdent(string(destination.propertyName)),
		}
		right := &ast.SelectorExpr{
			X:   ast.NewIdent(sourceVariable),
			Sel: ast.NewIdent(string(source.propertyName)),
		}
		return []ast.Stmt{
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
	return func(sourceVariable string, destinationVariable string, _ *CodeGenerationContext) []ast.Stmt {
		if sourceOptional == destinationOptional {
			// Can just copy a pointer to a primitive value
			assign := astbuilder.SimpleAssignment(
				astbuilder.QualifiedTypeName(destinationVariable, string(destination.propertyName)),
				token.ASSIGN,
				astbuilder.QualifiedTypeName(sourceVariable, string(source.propertyName)))
			return []ast.Stmt{assign}
		}

		if destinationOptional {
			// Need a pointer to the primitive value as the source is not optional
			assign := astbuilder.SimpleAssignment(
				astbuilder.QualifiedTypeName(destinationVariable, string(destination.propertyName)),
				token.ASSIGN,
				astbuilder.AddrOf(
					astbuilder.QualifiedTypeName(sourceVariable, string(source.propertyName))))
			return []ast.Stmt{assign}
		}

		if sourceOptional {
			// Need to check for null and only assign if we have a value
			cond := &ast.BinaryExpr{
				X:  astbuilder.QualifiedTypeName(sourceVariable, string(source.propertyName)),
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			}
			assignValue := astbuilder.SimpleAssignment(
				astbuilder.QualifiedTypeName(destinationVariable, string(destination.propertyName)),
				token.ASSIGN,
				astbuilder.Dereference(
					astbuilder.QualifiedTypeName(sourceVariable, string(source.propertyName))))
			assignZero := astbuilder.SimpleAssignment(
				astbuilder.QualifiedTypeName(destinationVariable, string(destination.propertyName)),
				token.ASSIGN,
				&ast.BasicLit{
					Value: zeroValue(sourceType),
				})
			stmt := &ast.IfStmt{
				Cond: cond,
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						assignValue,
					},
				},
				Else: &ast.BlockStmt{
					List: []ast.Stmt{
						assignZero,
					},
				},
			}
			return []ast.Stmt{stmt}
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
