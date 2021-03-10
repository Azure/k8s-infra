/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"
	"strings"
)

// ExternalPackageReference indicates a package to be imported from an external source, such as github or the go standard library.
type ExternalPackageReference struct {
	packagePath string
}

var _ PackageReference = ExternalPackageReference{}
var _ fmt.Stringer = ExternalPackageReference{}

// MakeExternalPackageReference creates a new package reference from a path
func MakeExternalPackageReference(packagePath string) ExternalPackageReference {
	return ExternalPackageReference{packagePath: packagePath}
}

// AsLocalPackage returns an empty local reference and false to indicate that library packages
// are not local
func (e ExternalPackageReference) AsLocalPackage() (LocalPackageReference, bool) {
	return LocalPackageReference{}, false
}

// PackageName returns the package name of this reference
func (e ExternalPackageReference) PackageName() string {
	l := strings.Split(e.packagePath, "/")
	return l[len(l)-1]
}

// PackagePath returns the fully qualified package path
func (e ExternalPackageReference) PackagePath() string {
	return e.packagePath
}

// Equals returns true if the passed package reference references the same package, false otherwise
func (e ExternalPackageReference) Equals(ref PackageReference) bool {
	if other, ok := ref.(ExternalPackageReference); ok {
		return e.packagePath == other.packagePath
	}

	return false
}

// IsPreview returns false because external references are never previews
func (e ExternalPackageReference) IsPreview() bool {
	return false
}

// String returns the string representation of the package reference
func (e ExternalPackageReference) String() string {
	return e.packagePath
}
