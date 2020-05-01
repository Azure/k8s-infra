/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"go/ast"
	"go/format"
	"go/token"
	"os"
)

// FileDefinition is the content of a file we're generating
type FileDefinition struct {
	// Name for the package
	PackageReference

	// Structs to include in this file
	structs []*StructDefinition
}

// FileDefinition must implement Definition
var _ Definition = &FileDefinition{}

// NewFileDefinition creates a file definition containing specified structs
func NewFileDefinition(structs ...*StructDefinition) *FileDefinition {
	// TODO: check that all structs are from same package
	return &FileDefinition{
		PackageReference: structs[0].PackageReference,
		structs:          structs,
	}
}

// AsAst generates an AST node representing this file
func (file *FileDefinition) AsAst() ast.Node {

	// Create import header:
	var imports []ast.Spec
	for _, s := range file.structs {
		for _, requiredImport := range s.RequiredImports() {
			imports = append(imports, &ast.ImportSpec{
				Name: nil,
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "\"github.com/Azure/k8s-infra/hack/generator/apis/" + requiredImport.PackagePath() + "\"",
				}},
			)
		}
	}

	var decls []ast.Decl
	if len(imports) > 0 {
		decls = append(decls, &ast.GenDecl{Tok: token.IMPORT, Specs: imports})
	}

	// Emit all structs:
	for _, s := range file.structs {
		decls = append(decls, s.AsDeclaration())
	}

	result := &ast.File{
		Name:  ast.NewIdent(file.PackageName()),
		Decls: decls,
	}

	return result
}

// SaveTo writes this generated file to disk
func (file *FileDefinition) SaveTo(filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err != nil {
		return err
	}

	content := file.AsAst()

	err = format.Node(f, token.NewFileSet(), content)
	if err != nil {
		return err
	}

	return nil
}
