/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"go/ast"
	"go/token"
	"sort"
)

// EnumType represents a set of mutually exclusive predefined options
type EnumType struct {
	// baseType is the underlying type used to define the values
	baseType *PrimitiveType
	// Options is the set of all unique values
	options []EnumValue
}

// EnumType must implement the Type interface correctly
var _ Type = (*EnumType)(nil)

// NewEnumType defines a new enumeration including the legal values
func NewEnumType(baseType *PrimitiveType, options []EnumValue) *EnumType {
	sort.Slice(options, func(left int, right int) bool {
		return options[left].Identifier < options[right].Identifier
	})

	return &EnumType{baseType: baseType, options: options}
}

// AsTypeAst renders a Go abstract syntax tree for referencing the type.
// As all enums are named, this will only be called when the parent NamedType is generating a declaration
func (enum *EnumType) AsTypeAst(codeGenerationContext *CodeGenerationContext) ast.Expr {

	//TODO: Need to include validation comments the declaration of the enumeration type
	/*
		validationComment := GenerateKubebuilderComment(enum.baseType.CreateValidation())
		declaration.Doc.List = append(
			declaration.Doc.List,
			&ast.Comment{Text: "\n" + validationComment})
	*/

	return enum.baseType.AsTypeAst(codeGenerationContext)
}

// AsDeclarationAsts implements Type for EnumType
func (enum *EnumType) AsDeclarationAsts(nameHint string, _ *CodeGenerationContext) []ast.Decl {
	var specs []ast.Spec
	for _, v := range enum.Options() {
		s := enum.createValueDeclaration(nameHint, v)
		specs = append(specs, s)
	}

	valuesDeclaration := &ast.GenDecl{
		Tok:   token.CONST,
		Doc:   &ast.CommentGroup{},
		Specs: specs,
	}

	return []ast.Decl{valuesDeclaration}
}

// References indicates whether this Type includes any direct references to the given Type
func (enum *EnumType) References(tn *TypeName) bool {
	return enum.baseType.References(tn)
}

// Equals will return true if the supplied type has the same base type and options
func (enum *EnumType) Equals(t Type) bool {
	if e, ok := t.(*EnumType); ok {
		if !enum.baseType.Equals(e.baseType) {
			return false
		}

		if len(enum.options) != len(e.options) {
			// Different number of fields, not equal
			return false
		}

		for i := range enum.options {
			if !enum.options[i].Equals(&e.options[i]) {
				return false
			}
		}

		// All options match, equal
		return true
	}

	return false
}

// RequiredImports indicates that Enums never need additional imports
func (enum *EnumType) RequiredImports() []*PackageReference {
	return nil
}

// CreateInternalDefinitions defines a named type for this enum and returns that type to be used in place
// of this "raw" enum type
func (enum *EnumType) CreateInternalDefinitions(nameHint *TypeName, idFactory IdentifierFactory) (Type, []*NamedType) {
	// an internal enum must always be named:
	definedEnum, otherTypes := enum.CreateNamedTypes(nameHint, idFactory, false)
	return definedEnum, append(otherTypes, definedEnum)
}

// CreateNamedTypes defines a named type for this "raw" enum type
func (enum *EnumType) CreateNamedTypes(name *TypeName, idFactory IdentifierFactory, _ bool) (*NamedType, []*NamedType) {
	identifier := idFactory.CreateEnumIdentifier(name.name)
	canonicalName := NewTypeName(name.PackageReference, identifier)
	return NewNamedType(canonicalName, enum), nil
}

// Options returns all the enum options
// A copy of the slice is returned to preserve immutability
func (enum *EnumType) Options() []EnumValue {
	//TODO: Do we need to sort these to guarantee determinism?
	return append(enum.options[:0:0], enum.options...)
}

// CreateValidation creates the validation annotation for this Enum
func (enum *EnumType) CreateValidation() Validation {
	var values []interface{}
	for _, opt := range enum.Options() {
		values = append(values, opt.Value)
	}

	return ValidateEnum(values)
}

func (enum *EnumType) Visit(visitor func(t Type)) {
	visitor(enum)
	enum.baseType.Visit(visitor)
}

func (enum *EnumType) createValueDeclaration(enumName string, value EnumValue) ast.Spec {
	enumIdentifier := ast.NewIdent(enumName)
	valueIdentifier := ast.NewIdent(enumName + value.Identifier)
	valueLiteral := ast.BasicLit{
		Kind:  token.STRING,
		Value: value.Value,
	}

	valueSpec := &ast.ValueSpec{
		Names: []*ast.Ident{valueIdentifier},
		Values: []ast.Expr{
			&ast.CallExpr{
				Fun:  enumIdentifier,
				Args: []ast.Expr{&valueLiteral},
			},
		},
	}

	return valueSpec
}
