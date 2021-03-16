/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import "github.com/pkg/errors"

const (
	genRuntimePathPrefix  = "github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
	GenRuntimePackageName = "genruntime"
	GroupSuffix           = ".infra.azure.com"
)

var MetaV1PackageReference = MakeExternalPackageReference("k8s.io/apimachinery/pkg/apis/meta/v1")

type PackageReference interface {
	// AsLocalPackage attempts conversion to a LocalPackageReference
	AsLocalPackage() (LocalPackageReference, bool)
	// Package returns the package name of this reference
	PackageName() string
	// PackagePath returns the fully qualified package path
	PackagePath() string
	// Equals returns true if the passed package reference references the same package, false otherwise
	Equals(ref PackageReference) bool
	// String returns the string representation of the package reference
	String() string
	// IsPreview returns true if this package reference is a preview
	IsPreview() bool
}

// PackageAsLocalPackage converts the given PackageReference into a LocalPackageReference if possible.
// If the provided PackageReference does not represent a local package an error is returned.
func PackageAsLocalPackage(pkg PackageReference) (LocalPackageReference, error) {
	if localPkg, ok := pkg.AsLocalPackage(); ok {
		return localPkg, nil
	}
	return LocalPackageReference{}, errors.Errorf("%q is not a local package", pkg.PackagePath())
}
