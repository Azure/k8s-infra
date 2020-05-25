/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"go/ast"
	"go/token"
)

// OneOfJSONMarshalFunction is a function for marshalling discriminated unions
// (types with only mutually exclusive properties) to JSON
type OneOfJSONMarshalFunction struct {
	oneOfStruct *StructType
	idFactory   IdentifierFactory // TODO: It's this or pass it in the AsFunc method
}

// NewOneOfJSONMarshalFunction creates a new OneOfJSONMarshalFunction struct
func NewOneOfJSONMarshalFunction(oneOfStruct *StructType, idFactory IdentifierFactory) *OneOfJSONMarshalFunction {
	return &OneOfJSONMarshalFunction{oneOfStruct, idFactory}
}

// Ensure OneOfJSONMarshalFunction implements Function interface correctly
var _ Function = (*OneOfJSONMarshalFunction)(nil)

// Equals determines if this function is equal to the passed in function
func (f *OneOfJSONMarshalFunction) Equals(other Function) bool {
	if o, ok := other.(*OneOfJSONMarshalFunction); ok {
		return f.oneOfStruct.Equals(o.oneOfStruct)
	}

	return false
}

// References indicates whether this function includes any direct references to the given type
func (f *OneOfJSONMarshalFunction) References(name *TypeName) bool {
	// Defer this check to the owning struct as we only refer to its fields and it
	return f.oneOfStruct.References(name)
}

// AsFunc returns the function as a go ast
func (f *OneOfJSONMarshalFunction) AsFunc(receiver *TypeName, methodName string) *ast.FuncDecl {
	receiverName := f.idFactory.CreateIdentifier(receiver.name, Internal)

	header, _ := createComments("Marshal marshals the object as JSON")

	result := &ast.FuncDecl{
		Doc: &ast.CommentGroup{
			List: header,
		},
		Name: ast.NewIdent(methodName),
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Type:  receiver.AsType(),
					Names: []*ast.Ident{ast.NewIdent(receiverName)},
				},
			},
		},
		Type: &ast.FuncType{
			Func: token.HighestPrec,
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: ast.NewIdent("[]byte"),
					},
					{
						Type: ast.NewIdent("error"),
					},
				},
			},
		},
	}

	var statements []ast.Stmt

	for _, field := range f.oneOfStruct.fields {
		fieldSelectorExpr := &ast.SelectorExpr{
			X:   ast.NewIdent(receiverName),
			Sel: ast.NewIdent(string(field.fieldName)),
		}

		ifStatement := ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  fieldSelectorExpr,
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   ast.NewIdent("json"),
									Sel: ast.NewIdent("Marshal"),
								},
								Args: []ast.Expr{
									fieldSelectorExpr,
								},
							},
						},
					},
				},
			},
		}

		statements = append(statements, &ifStatement)
	}

	finalReturnStatement := &ast.ReturnStmt{
		Results: []ast.Expr{
			ast.NewIdent("nil, nil"), // TODO: Is this the right way to do this?
		},
	}
	statements = append(statements, finalReturnStatement)

	result.Body = &ast.BlockStmt{
		List: statements,
	}

	return result
}

// RequiredImports returns a list of packages required by this
func (f *OneOfJSONMarshalFunction) RequiredImports() []PackageReference {
	return []PackageReference{
		{"encoding/json"},
	}
}
