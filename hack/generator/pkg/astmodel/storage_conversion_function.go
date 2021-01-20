/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astbuilder"
	"github.com/dave/dst"
	"sort"
)

// StoragePropertyConversion generates the AST for a given conversion.
// source is a factory function that returns an expression for the source of the conversion.
// destination is a factory function that returns an expression for the destination of the conversion.
// The parameters source and destination are funcs because AST fragments can't be reused, and in
// some cases we need to reference source and destination multiple times in a single fragment.
type StoragePropertyConversion func(source func() dst.Expr, destination func() dst.Expr, ctx *CodeGenerationContext) []dst.Stmt

// StoragePropertyConversionFactory is a factory func that creates a StoragePropertyConversion for later use.
// source is the PropertyDefinition for the origin which will be read.
// destination is the PropertyDefinition for the target which will be written.
type StoragePropertyConversionFactory func(sourceType Type, destinationType *PropertyDefinition) StoragePropertyConversion

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
) (*StorageConversionFunction, []error) {
	result := &StorageConversionFunction{
		name:                "ConvertFrom",
		parameter:           source,
		staging:             staging,
		idFactory:           idFactory,
		conversionDirection: ConvertFrom,
		conversions:         make(map[string]StoragePropertyConversion),
	}

	errs := result.createConversions(receiver)
	return result, errs
}

func NewStorageConversionToFunction(
	receiver TypeDefinition,
	destination TypeName,
	staging TypeDefinition,
	idFactory IdentifierFactory,
) (*StorageConversionFunction, []error) {
	result := &StorageConversionFunction{
		name:                "ConvertTo",
		parameter:           destination,
		staging:             staging,
		idFactory:           idFactory,
		conversionDirection: ConvertTo,
		conversions:         make(map[string]StoragePropertyConversion),
	}

	errs := result.createConversions(receiver)
	return result, errs
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

func (fn *StorageConversionFunction) AsFunc(ctx *CodeGenerationContext, receiver TypeName) *dst.FuncDecl {

	parameterName := fn.parameterName()
	parameterIdent := func() dst.Expr {
		return dst.NewIdent(fn.parameterName())
	}

	receiverName := fn.receiverName(receiver)
	receiverIdent := func() dst.Expr {
		return dst.NewIdent(receiverName)
	}

	funcDetails := &astbuilder.FuncDetails{
		ReceiverIdent: receiverName,
		ReceiverType:  receiver.AsType(ctx),
		Name:          fn.Name(),
		Body:          fn.generateBody(receiverIdent, parameterIdent, ctx),
	}

	funcDetails.AddParameter(
		parameterName,
		&dst.StarExpr{
			X: fn.staging.name.AsType(ctx)})
	funcDetails.AddReturns("error")

	return funcDetails.DefineFunc()
}

// generateBody returns all of the statements required for the conversion function
// receiver a function returning the name of our receiver type, used to qualify field access
// parameter a function returning the name of the parameter passed to the function, also used for field access
// ctx is our code generation context, passed to allow resolving of identifiers in other packages
func (fn *StorageConversionFunction) generateBody(receiver func() dst.Expr, parameter func() dst.Expr, ctx *CodeGenerationContext) []dst.Stmt {

	if fn.parameter.Equals(fn.staging.name) {
		// Last step of conversion, directly to the parameter type we've been given
		if fn.conversionDirection == ConvertFrom {
			return fn.generateDirectConversionFrom(receiver, parameter, ctx)
		} else {
			// fn.conversionType == ConvertTo
			return fn.generateDirectConversionTo(receiver, parameter, ctx)
		}
	}

	// Intermediate step of conversion, not working directly with the parameter type we've been given
	if fn.conversionDirection == ConvertFrom {
		return fn.generateIndirectConversionFrom(receiver, parameter, ctx)
	} else {
		// fn.conversionType == ConvertTo
		return fn.generateIndirectConversionTo(receiver, parameter, ctx)
	}
}

// generateDirectConversionFrom returns the method body required to directly copy information from
// the parameter instance onto our receiver
func (fn *StorageConversionFunction) generateDirectConversionFrom(receiver func() dst.Expr, parameter func() dst.Expr, ctx *CodeGenerationContext) []dst.Stmt {
	return fn.generateAssignments(parameter, receiver, ctx)
}

// generateDirectConversionTo returns the method body required to directly copy information from
// our receiver onto the parameter instance
func (fn *StorageConversionFunction) generateDirectConversionTo(receiver func() dst.Expr, parameter func() dst.Expr, ctx *CodeGenerationContext) []dst.Stmt {
	return fn.generateAssignments(receiver, parameter, ctx)
}

// generateIndirectConversionFrom returns the method body required to populate our receiver when
// we don't directly understand the structure of the parameter value.
// To accommodate this, we first convert to an intermediate form:
//
// var staging IntermediateType
// staging.ConvertFrom(parameter)
// [copy values from staging]
//
func (fn *StorageConversionFunction) generateIndirectConversionFrom(receiver func() dst.Expr, parameter func() dst.Expr, ctx *CodeGenerationContext) []dst.Stmt {
	staging := astbuilder.LocalVariableDeclaration(
		"staging", dst.NewIdent(fn.staging.name.name), "// staging is our intermediate type for conversion")
	staging.Decorations().Before = dst.NewLine

	convertFrom := astbuilder.InvokeQualifiedFunc(
		"staging", fn.name, parameter())
	convertFrom.Decorations().Before = dst.EmptyLine
	convertFrom.Decorations().Start.Append("// first populate staging")

	assignments := fn.generateAssignments(
		func() dst.Expr {
			return dst.NewIdent("staging")
		},
		receiver,
		ctx)

	var result []dst.Stmt
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
func (fn *StorageConversionFunction) generateIndirectConversionTo(receiver func() dst.Expr, parameter func() dst.Expr, ctx *CodeGenerationContext) []dst.Stmt {
	staging := astbuilder.LocalVariableDeclaration(
		"staging", dst.NewIdent(fn.staging.name.name), "// staging is our intermediate type for conversion")
	staging.Decorations().Before = dst.NewLine

	convertTo := astbuilder.InvokeQualifiedFunc(
		"staging", fn.name, parameter())
	convertTo.Decorations().Before = dst.EmptyLine
	convertTo.Decorations().Start.Append("// use staging to populate")

	assignments := fn.generateAssignments(
		receiver,
		func() dst.Expr {
			return dst.NewIdent("staging")
		},
		ctx)

	var result []dst.Stmt
	result = append(result, staging)
	result = append(result, assignments...)
	result = append(result, convertTo)
	return result
}

func (fn *StorageConversionFunction) generateAssignments(source func() dst.Expr, destination func() dst.Expr, ctx *CodeGenerationContext) []dst.Stmt {
	var result []dst.Stmt

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
			firstStatement.Decorations().Before = dst.EmptyLine
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

func (fn *StorageConversionFunction) createConversions(receiver TypeDefinition) []error {
	receiverObject := AsObjectType(receiver.Type())
	otherObject := AsObjectType(fn.staging.Type())
	var errs []error

	for _, receiverProperty := range receiverObject.Properties() {
		otherProperty, ok := otherObject.Property(receiverProperty.propertyName)
		//TODO: Handle renames
		if ok {
			var conv StoragePropertyConversion
			var err error
			if fn.conversionDirection == ConvertFrom {
				conv, err = createPropertyConversion(otherProperty, receiverProperty)
			} else {
				conv, err = createPropertyConversion(receiverProperty, otherProperty)
			}

			if conv != nil {
				// A conversion was created, keep it for later
				fn.conversions[string(receiverProperty.propertyName)] = conv
			}

			if err != nil {
				// An error was returned; this can happen even if a conversion was created as well.
				errs = append(errs, err)
			}
		}
	}

	return errs
}
