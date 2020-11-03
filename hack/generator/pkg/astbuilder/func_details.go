/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astbuilder

import (
	"fmt"
	"go/ast"
	"strings"
)

type FuncDetails struct {
	ReceiverIdent *ast.Ident
	ReceiverType  ast.Expr
	Name          *ast.Ident
	Comments      []string
	Params        []*ast.Field
	Returns       []*ast.Field
	Body          []ast.Stmt
}

// NewTestFuncDetails() returns a FuncDetails for a test method
// Tests require a particular signature, so this makes it simpler to create test functions
func NewTestFuncDetails(testName string, body ...ast.Stmt) *FuncDetails {

	// Ensure the method name starts with `Test` as required
	var name string
	if strings.HasPrefix(testName, "Test") {
		name = testName
	} else {
		name = "Test_" + testName
	}

	result := &FuncDetails{
		Name: ast.NewIdent(name),
		Body: body,
	}

	result.AddParameter("t",
		&ast.StarExpr{
			X: &ast.SelectorExpr{
				X:   ast.NewIdent("testing"),
				Sel: ast.NewIdent("T"),
			}},
	)

	return result
}

// DefineFunc defines a function (header, body, etc), like:
// 	<comment>
//	func (<receiverIdent> <receiverType>) <name>(<params...>) (<returns...>) {
// 		<body...>
//	}
func DefineFunc(funcDetails *FuncDetails) *ast.FuncDecl {

	// Safety check that we are making something valid
	if (funcDetails.ReceiverIdent == nil) != (funcDetails.ReceiverType == nil) {
		reason := fmt.Sprintf(
			"ReceiverIdent and ReceiverType must both be specified, or both omitted. ReceiverIdent: %q, ReceiverType: %q",
			funcDetails.ReceiverIdent,
			funcDetails.ReceiverType)
		panic(reason)
	}

	// Filter out any nil statements
	// this helps creation of the funcDetails go simpler
	var body []ast.Stmt
	for _, s := range funcDetails.Body {
		if s != nil {
			body = append(body, s)
		}
	}

	var comment []*ast.Comment
	if len(funcDetails.Comments) > 0 {
		funcDetails.Comments[0] = fmt.Sprintf("// %s %s", funcDetails.Name, funcDetails.Comments[0])
		AddComments(&comment, funcDetails.Comments)
	}

	result := &ast.FuncDecl{
		Name: funcDetails.Name,
		Doc: &ast.CommentGroup{
			List: comment,
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: funcDetails.Params,
			},
			Results: &ast.FieldList{
				List: funcDetails.Returns,
			},
		},
		Body: &ast.BlockStmt{
			List: body,
		},
	}

	if funcDetails.ReceiverIdent != nil {
		// We have a receiver, so include it

		field := &ast.Field{
			Names: []*ast.Ident{
				funcDetails.ReceiverIdent,
			},
			Type: funcDetails.ReceiverType,
		}

		recv := ast.FieldList{
			List: []*ast.Field{field},
		}

		result.Recv = &recv
	}

	return result
}

// AddStatements adds additional statements to the function
func (fn *FuncDetails) AddStatements(statements ...ast.Stmt) {
	fn.Body = append(fn.Body, statements...)
}

// AddParameter adds another parameter to the function definition
func (fn *FuncDetails) AddParameter(id string, parameterType ast.Expr) {
	field := &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(id)},
		Type:  parameterType,
	}
	fn.Params = append(fn.Params, field)
}

func (fn *FuncDetails) AddComments(comment ...string) {
	fn.Comments = append(fn.Comments, comment...)
}
