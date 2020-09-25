/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"
	"strings"
)

// ExternalPackageReference indicates a library package that needs to be imported
type ExternalPackageReference struct {
	packagePath string
}

var _ PackageReference = ExternalPackageReference{}
var _ fmt.Stringer = ExternalPackageReference{}

// MakeExternalPackageReference creates a new package reference from a path
func MakeExternalPackageReference(packagePath string) ExternalPackageReference {
	return ExternalPackageReference{packagePath: packagePath}
}

// IsLocalPackage returns false to indicate that library packages are not local
func (pr ExternalPackageReference) AsLocalPackage() (LocalPackageReference, bool) {
	return LocalPackageReference{}, false
}

// Package returns the package name of this reference
func (pr ExternalPackageReference) PackageName() string {
	l := strings.Split(pr.packagePath, "/")
	return l[len(l)-1]
}

// PackagePath returns the fully qualified package path
func (pr ExternalPackageReference) PackagePath() string {
	return pr.packagePath
}

// Equals returns true if the passed package reference references the same package, false otherwise
func (pr ExternalPackageReference) Equals(ref PackageReference) bool {
	if other, ok := ref.(ExternalPackageReference); ok {
		return pr.packagePath == other.packagePath
	}

	return false
}

// String returns the string representation of the package reference
func (pr ExternalPackageReference) String() string {
	return pr.packagePath
}
