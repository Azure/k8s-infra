/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"
	"go/ast"
	"go/token"
)

// PropertyName is a semantic type
type PropertyName string

// PropertyDefinition encapsulates the definition of a property
type PropertyDefinition struct {
	propertyName PropertyName
	propertyType Type
	jsonName     string
	description  string
	validations  []Validation
}

// NewPropertyDefinition is a factory method for creating a new PropertyDefinition
// name is the name for the new property (mandatory)
// propertyType is the type for the new property (mandatory)
func NewPropertyDefinition(propertyName PropertyName, jsonName string, propertyType Type) *PropertyDefinition {
	return &PropertyDefinition{
		propertyName: propertyName,
		propertyType: propertyType,
		jsonName:     jsonName,
		description:  "",
	}
}

// PropertyName returns the name of the property
func (property *PropertyDefinition) PropertyName() PropertyName {
	return property.propertyName
}

// PropertyType returns the data type of the property
func (property *PropertyDefinition) PropertyType() Type {
	return property.propertyType
}

// WithDescription returns a new PropertyDefinition with the specified description
func (property *PropertyDefinition) WithDescription(description *string) *PropertyDefinition {
	if description == nil {
		// Special handling for nil
		d := ""
		return property.WithDescription(&d)
	}

	if *description == property.description {
		return property
	}

	result := *property
	result.description = *description
	return &result
}

// WithType clones the property and returns it with a new type
func (property *PropertyDefinition) WithType(newType Type) *PropertyDefinition {
	if property.propertyType == newType {
		return property
	}

	result := *property
	result.propertyType = newType
	return &result
}

// WithValidation adds the given validation to the property's set of validations
func (property *PropertyDefinition) WithValidation(validation Validation) *PropertyDefinition {
	result := *property
	result.validations = append(result.validations, validation)
	return &result
}

// MakeRequired returns a new PropertyDefinition that is marked as required
func (property *PropertyDefinition) MakeRequired() *PropertyDefinition {
	if !property.HasOptionalType() && property.HasRequiredValidation() {
		return property
	}

	result := *property

	if property.HasOptionalType() {
		// Need to remove the optionality
		ot := property.propertyType.(*OptionalType)
		result.propertyType = ot.BaseType()
	}

	if !property.HasRequiredValidation() {
		result = *result.WithValidation(ValidateRequired())
	}

	return &result
}

// MakeOptional returns a new PropertyDefinition that has an optional value
func (property *PropertyDefinition) MakeOptional() *PropertyDefinition {
	if property.HasOptionalType() && !property.HasRequiredValidation() {
		// No change required
		return property
	}

	result := *property

	if !property.HasOptionalType() {
		// Need to make the type optional
		result.propertyType = NewOptionalType(result.propertyType)
	}

	if property.HasRequiredValidation() {
		// Need to remove the Required validation
		var validations []Validation
		for _, v := range result.validations {
			if !v.HasName(RequiredValidationName) {
				validations = append(validations, v)
			}
		}

		result.validations = validations
	}

	return &result
}

// HasRequiredValidation returns true if the property has validation specifying that it is required;
// returns false otherwise.
func (property *PropertyDefinition) HasRequiredValidation() bool {
	required := ValidateRequired()
	for _, v := range property.validations {
		if v == required {
			return true
		}
	}

	return false
}

// HasOptionalType returns true if the type of this property is an optioan reference to a value
// (and might therefore be nil).
func (property *PropertyDefinition) HasOptionalType() bool {
	_, ok := property.propertyType.(*OptionalType)
	return ok
}

// AsField generates a Go AST field node representing this property definition
func (property *PropertyDefinition) AsField(codeGenerationContext *CodeGenerationContext) *ast.Field {

	result := &ast.Field{
		Doc:   &ast.CommentGroup{},
		Names: []*ast.Ident{ast.NewIdent(string(property.propertyName))},
		Type:  property.PropertyType().AsType(codeGenerationContext),
		Tag: &ast.BasicLit{
			Kind:  token.STRING,
			Value: fmt.Sprintf("`json:%q`", property.jsonName),
		},
	}

	// generate validation comments:
	for _, validation := range property.validations {
		// these are not doc comments but they must go here to be emitted before the property
		addDocComment(&result.Doc.List, GenerateKubebuilderComment(validation), 200)
	}

	// generate doc comment:
	if property.description != "" {
		addDocComment(&result.Doc.List, fmt.Sprintf("%s: %s", property.propertyName, property.description), 80)
	}

	return result
}

// Equals tests to see if the specified PropertyDefinition specifies the same property
func (property *PropertyDefinition) Equals(f *PropertyDefinition) bool {
	return property == f || (property.propertyName == f.propertyName && property.propertyType.Equals(f.propertyType))
}
