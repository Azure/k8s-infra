/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"go/ast"
	"go/token"
)

// PackageReference indicates which package
// a struct belongs to.
type PackageImport struct {
	PackageReference PackageReference // This is used as the key in a map so can't be pointer
	name             *string
}

// NewPackageImport creates a new package import from a reference
func NewPackageImport(packageReference PackageReference) *PackageImport {
	return &PackageImport{
		PackageReference: packageReference,
	}
}

// WithName creates a new package reference with a friendly name
func (pi *PackageImport) WithName(name string) *PackageImport {
	result := NewPackageImport(pi.PackageReference)
	result.name = &name

	return result
}

func (pi *PackageImport) AsImportSpec() *ast.ImportSpec {
	var name *ast.Ident
	if pi.name != nil {
		name = ast.NewIdent(*pi.name)
	}

	return &ast.ImportSpec{
		Name: name,
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: "\"" + pi.PackageReference.PackagePath() + "\"",
		},
	}
}

// PackageName is the package name of the package reference
func (pi *PackageImport) PackageName() string {
	if pi.name != nil {
		return *pi.name
	}

	return pi.PackageReference.PackageName()
}

// Equals returns true if the passed package reference references the same package, false otherwise
func (pi *PackageImport) Equal(ref *PackageImport) bool {
	return pi.PackageReference.Equal(&ref.PackageReference) && pi.name == ref.name
}
