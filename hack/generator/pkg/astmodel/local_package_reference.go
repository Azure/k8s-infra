/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"
	"strings"
)

// LocalPackageReference specifies a local package name or reference
type LocalPackageReference struct {
	localPathPrefix string
	group           string
	version         string
}

var _ PackageReference = LocalPackageReference{}
var _ fmt.Stringer = LocalPackageReference{}

// MakeLocalPackageReference Creates a new local package reference from a group and version
func MakeLocalPackageReference(prefix string, group string, version string) LocalPackageReference {
	return LocalPackageReference{localPathPrefix: prefix, group: group, version: version}
}

// AsLocalPackage returns this instance and true
func (l LocalPackageReference) AsLocalPackage() (LocalPackageReference, bool) {
	return l, true
}

// Group returns the group of this local reference
func (l LocalPackageReference) Group() string {
	return l.group
}

// Version returns the version of this local reference
func (l LocalPackageReference) Version() string {
	return l.version
}

// PackageName returns the package name of this reference
func (l LocalPackageReference) PackageName() string {
	return l.version
}

// PackagePath returns the fully qualified package path
func (l LocalPackageReference) PackagePath() string {
	url := l.localPathPrefix + "/" + l.group + "/" + l.version
	return url
}

// Equals returns true if the passed package reference references the same package, false otherwise
func (l LocalPackageReference) Equals(ref PackageReference) bool {
	if ref == nil {
		return false
	}

	if other, ok := ref.AsLocalPackage(); ok {
		return l.localPathPrefix == other.localPathPrefix &&
			l.version == other.version &&
			l.group == other.group
	}

	return false
}

// String returns the string representation of the package reference
func (l LocalPackageReference) String() string {
	return l.PackagePath()
}

// IsPreview returns true if this package reference is a preview
func (l LocalPackageReference) IsPreview() bool {
	lc := strings.ToLower(l.version)
	return strings.Contains(lc, "alpha") ||
		strings.Contains(lc, "beta") ||
		strings.Contains(lc, "preview")
}

// IsLocalPackageReference returns true if the supplied reference is a local one
func IsLocalPackageReference(ref PackageReference) bool {
	_, ok := ref.(LocalPackageReference)
	return ok
}
