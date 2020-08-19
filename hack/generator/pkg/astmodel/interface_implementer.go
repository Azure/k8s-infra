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

// TODO: need to add tests for this probably
type InterfaceImplementer struct { // TODO: Do we like this name?
	interfaces map[TypeName]*InterfaceImplementation
}

// MakeInterfaceImplementer returns an interface implementer
func MakeInterfaceImplementer() InterfaceImplementer {
	return InterfaceImplementer{}
}

// WithInterface creates a new ObjectType with a function (method) attached to it
func (interfaceImplementer InterfaceImplementer) WithInterface(iface *InterfaceImplementation) InterfaceImplementer {
	result := interfaceImplementer.copy()
	result.interfaces[iface.Name()] = iface

	return result
}

//func (interfaceImplementer *InterfaceImplementer) Interfaces() {
//	panic("TODO")
//}

func (interfaceImplementer InterfaceImplementer) References() TypeNameSet {
	var results TypeNameSet
	for _, iface := range interfaceImplementer.interfaces {
		for ref := range iface.References() {
			results = results.Add(ref)
		}
	}

	return results
}

func (interfaceImplementer InterfaceImplementer) AsType(codeGenerationContext *CodeGenerationContext) ast.Expr {
	panic("InterfaceImplementer cannot be used as a standalone type")
}

func (interfaceImplementer InterfaceImplementer) AsDeclarations(
	codeGenerationContext *CodeGenerationContext,
	typeName TypeName,
	_ []string) []ast.Decl {

	var result []ast.Decl

	// First interfaces must be ordered by name for deterministic output
	var ifaceNames []TypeName
	for ifaceName := range interfaceImplementer.interfaces {
		ifaceNames = append(ifaceNames, ifaceName)
	}

	sort.Slice(ifaceNames, func(i int, j int) bool {
		return ifaceNames[i].name < ifaceNames[j].name
	})

	for _, ifaceName := range ifaceNames {
		iface := interfaceImplementer.interfaces[ifaceName]

		result = append(result, interfaceImplementer.generateInterfaceImplAssertion(codeGenerationContext, iface, typeName))

		var funcNames []string
		for funcName := range iface.functions {
			funcNames = append(funcNames, funcName)
		}

		sort.Strings(funcNames)

		for _, methodName := range funcNames {
			function := iface.functions[methodName]
			result = append(result, function.AsFunc(codeGenerationContext, typeName, methodName))
		}
	}

	return result
}

func (interfaceImplementer InterfaceImplementer) Equals(t Type) bool {

	otherInterfaceImplementer, ok := t.(InterfaceImplementer)
	if !ok {
		return false
	}

	if len(interfaceImplementer.interfaces) != len(otherInterfaceImplementer.interfaces) {
		return false
	}

	for ifaceName, iface := range interfaceImplementer.interfaces {
		otherIface, ok := otherInterfaceImplementer.interfaces[ifaceName]
		if !ok {
			return false
		}

		if !iface.Equals(otherIface) {
			return false
		}
	}

	return true
}

// TODO: Do we actually want to impl this - I kinda feel like we actually don't?
// TODO: maybe we want an interface with _most_ of these methods but not AsType?
var _ Type = InterfaceImplementer{}

func (interfaceImplementer InterfaceImplementer) RequiredImports() []PackageReference {
	var result []PackageReference

	for _, i := range interfaceImplementer.interfaces {
		result = append(result, i.RequiredImports()...)
	}

	return result
}

func (interfaceImplementer InterfaceImplementer) generateInterfaceImplAssertion(
	codeGenerationContext *CodeGenerationContext,
	iface *InterfaceImplementation,
	typeName TypeName) ast.Decl {

	ifacePackageName, err := codeGenerationContext.GetImportedPackageName(iface.name.PackageReference)
	if err != nil {
		panic(err)
	}

	typeAssertion := &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Type: &ast.SelectorExpr{
					X:   ast.NewIdent(ifacePackageName),
					Sel: ast.NewIdent(iface.name.name),
				},
				Names: []*ast.Ident{
					ast.NewIdent("_"),
				},
				Values: []ast.Expr{
					&ast.UnaryExpr{
						Op: token.AND,
						X: &ast.CompositeLit{
							Type: ast.NewIdent(typeName.name),
						},
					},
				},
			},
		},
	}

	return typeAssertion
}

func (interfaceImplementer InterfaceImplementer) copy() InterfaceImplementer {
	result := interfaceImplementer

	result.interfaces = make(map[TypeName]*InterfaceImplementation, len(interfaceImplementer.interfaces))
	for k, v := range interfaceImplementer.interfaces {
		result.interfaces[k] = v
	}

	return result
}
