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

// StoragePropertyConversion represents a function that generates the correct AST to convert a single property value
// Different functions will be used, depending on the types of the properties to be converted.
// source is an expression for the source value that will be read.
// destination is an expression the target value that will be written.
type StoragePropertyConversion func(source dst.Expr, destination dst.Expr, ctx *CodeGenerationContext) []dst.Stmt

// StorageConversionFunction represents a function that performs conversions for storage versions
type StorageConversionFunction struct {
	// Name of this conversion function
	name string
	// Name of the ultimate hub type to which we are converting, passed as a parameter
	parameter TypeName
	// Other Type of this conversion stage (if this isn't our parameter type, we'll delegate to this to finish the job)
	staging TypeDefinition
	// Map of all property conversions we are going to use, keyed by name of the receiver property
	conversions map[string]StoragePropertyConversion
	// Reference to our identifier factory
	idFactory IdentifierFactory
	// Which conversionType of conversion are we generating?
	conversionDirection StorageConversionDirection
	// A cached set of local identifiers that have already been used, to avoid conflicts
	knownLocals KnownLocalsSet
}

// StorageConversionDirection specifies the direction of conversion we're implementing with this function
type StorageConversionDirection int

const (
	// Indicates the conversion is from the passed other instance, populating the receiver
	ConvertFrom = StorageConversionDirection(1)
	// Indicate the conversion is to the passed other type, populating other
	ConvertTo = StorageConversionDirection(2)
)

// Ensure that StorageConversionFunction implements Function
var _ Function = &StorageConversionFunction{}

// NewStorageConversionFromFunction creates a new StorageConversionFunction to convert from the specified source
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
		knownLocals:         make(KnownLocalsSet),
	}

	errs := result.createConversions(receiver)
	return result, errs
}

// NewStorageConversionToFunction creates a new StorageConversionFunction to convert to the specified destination
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
		knownLocals:         make(map[string]struct{}),
	}

	errs := result.createConversions(receiver)
	return result, errs
}

// Name returns the name of this function
func (fn *StorageConversionFunction) Name() string {
	return fn.name
}

// RequiredPackageReferences returns the set of package references required by this function
func (fn *StorageConversionFunction) RequiredPackageReferences() *PackageReferenceSet {
	return NewPackageReferenceSet(
		fn.parameter.PackageReference,
		fn.staging.name.PackageReference)
}

// References returns the set of types referenced by this function
func (fn *StorageConversionFunction) References() TypeNameSet {
	return NewTypeNameSet(fn.parameter, fn.staging.name)
}

// Equals checks to see if the supplied function is the same as this one
func (fn *StorageConversionFunction) Equals(f Function) bool {
	if other, ok := f.(*StorageConversionFunction); ok {
		// Only check name for now
		if fn.name != other.name {
			return false
		}
	}

	return false
}

// AsFunc renders this function as an AST for serialization to a Go source file
func (fn *StorageConversionFunction) AsFunc(ctx *CodeGenerationContext, receiver TypeName) *dst.FuncDecl {

	parameterName := "other" // safe default
	if fn.conversionDirection == ConvertTo {
		parameterName = "destination"
	} else if fn.conversionDirection == ConvertFrom {
		parameterName = "source"
	}

	receiverName := fn.idFactory.CreateIdentifier(receiver.Name(), NotExported)

	funcDetails := &astbuilder.FuncDetails{
		ReceiverIdent: receiverName,
		ReceiverType:  receiver.AsType(ctx),
		Name:          fn.Name(),
		Body:          fn.generateBody(receiverName, parameterName, ctx),
	}

	parameterPackage := ctx.MustGetImportedPackageName(fn.parameter.PackageReference)

	funcDetails.AddParameter(
		parameterName,
		&dst.StarExpr{
			X: &dst.SelectorExpr{
				X:   dst.NewIdent(parameterPackage),
				Sel: dst.NewIdent(fn.parameter.name),
			},
		})

	funcDetails.AddReturns("error")

	return funcDetails.DefineFunc()
}

// generateBody returns all of the statements required for the conversion function
// receiver is an expression for access our receiver type, used to qualify field access
// parameter is an expression for access to our parameter passed to the function, also used for field access
// ctx is our code generation context, passed to allow resolving of identifiers in other packages
func (fn *StorageConversionFunction) generateBody(receiver string, parameter string, ctx *CodeGenerationContext) []dst.Stmt {

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
func (fn *StorageConversionFunction) generateDirectConversionFrom(receiver string, parameter string, ctx *CodeGenerationContext) []dst.Stmt {
	receiverIdent := dst.NewIdent(receiver)
	parameterIdent := dst.NewIdent(parameter)

	return fn.generateAssignments(parameterIdent, receiverIdent, ctx)
}

// generateDirectConversionTo returns the method body required to directly copy information from
// our receiver onto the parameter instance
func (fn *StorageConversionFunction) generateDirectConversionTo(receiver string, parameter string, ctx *CodeGenerationContext) []dst.Stmt {
	receiverIdent := dst.NewIdent(receiver)
	parameterIdent := dst.NewIdent(parameter)

	return fn.generateAssignments(receiverIdent, parameterIdent, ctx)
}

// generateIndirectConversionFrom returns the method body required to populate our receiver when
// we don't directly understand the structure of the parameter value.
// To accommodate this, we first convert to an intermediate form:
//
// var staging IntermediateType
// staging.ConvertFrom(parameter)
// [copy values from staging]
//
func (fn *StorageConversionFunction) generateIndirectConversionFrom(receiver string, parameter string, ctx *CodeGenerationContext) []dst.Stmt {

	local := fn.knownLocals.createLocal(receiver + "Temp")

	parameterPackage := ctx.MustGetImportedPackageName(fn.staging.name.PackageReference)
	localDeclaration := astbuilder.LocalVariableDeclaration(
		local,
		&dst.SelectorExpr{
			X:   dst.NewIdent(parameterPackage),
			Sel: dst.NewIdent(fn.staging.name.name),
		},
		fmt.Sprintf("// %s is our intermediate for conversion", local))
	localDeclaration.Decorations().Before = dst.NewLine

	callConvertFrom := astbuilder.InvokeQualifiedFunc(
		local, fn.name, dst.NewIdent(parameter))
	callConvertFrom.Decorations().Before = dst.EmptyLine
	callConvertFrom.Decorations().Start.Append(
		fmt.Sprintf("// Populate %s from %s", local, parameter))

	assignments := fn.generateAssignments(
		dst.NewIdent(local),
		dst.NewIdent(receiver),
		ctx)

	var result []dst.Stmt
	result = append(result, localDeclaration)
	result = append(result, callConvertFrom)
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
func (fn *StorageConversionFunction) generateIndirectConversionTo(receiver string, parameter string, ctx *CodeGenerationContext) []dst.Stmt {

	local := fn.knownLocals.createLocal(receiver + "Temp")

	parameterPackage := ctx.MustGetImportedPackageName(fn.staging.name.PackageReference)
	localDeclaration := astbuilder.LocalVariableDeclaration(
		local,
		&dst.SelectorExpr{
			X:   dst.NewIdent(parameterPackage),
			Sel: dst.NewIdent(fn.staging.name.name),
		},
		fmt.Sprintf("// %s is our intermediate for conversion", local))
	localDeclaration.Decorations().Before = dst.NewLine

	callConvertTo := astbuilder.InvokeQualifiedFunc(
		local, fn.name, dst.NewIdent(parameter))
	callConvertTo.Decorations().Before = dst.EmptyLine
	callConvertTo.Decorations().Start.Append(
		fmt.Sprintf("// Populate %s from %s", parameter, local))

	assignments := fn.generateAssignments(
		dst.NewIdent(receiver),
		dst.NewIdent(local),
		ctx)

	var result []dst.Stmt
	result = append(result, localDeclaration)
	result = append(result, assignments...)
	result = append(result, callConvertTo)
	return result
}

// generateAssignments generates a sequence of statements to copy information between the two types
func (fn *StorageConversionFunction) generateAssignments(source dst.Expr, destination dst.Expr, ctx *CodeGenerationContext) []dst.Stmt {
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

// createConversions iterates through the properties on our receiver type, matching them up with
// our other type and generating conversions where possible
func (fn *StorageConversionFunction) createConversions(receiver TypeDefinition) []error {
	receiverObject := AsObjectType(receiver.Type())
	otherObject := AsObjectType(fn.staging.Type())
	var errs []error

	// Flag receiver name as used
	fn.knownLocals.Add(receiver.name.name)

	for _, receiverProperty := range receiverObject.Properties() {
		otherProperty, ok := otherObject.Property(receiverProperty.propertyName)
		//TODO: Handle renames
		if ok {
			var conv StoragePropertyConversion
			var err error
			if fn.conversionDirection == ConvertFrom {
				conv, err = createPropertyConversion(otherProperty, receiverProperty, fn.knownLocals)
			} else {
				conv, err = createPropertyConversion(receiverProperty, otherProperty, fn.knownLocals)
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
