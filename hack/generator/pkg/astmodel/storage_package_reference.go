/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"
	"strings"
)

const (
	StoragePackageSuffix = "storage"
)

type StoragePackageReference struct {
	LocalPackageReference
}

var _ PackageReference = StoragePackageReference{}

// MakeStoragePackageReference creates a new storage package reference from a local package reference
func MakeStoragePackageReference(local LocalPackageReference) StoragePackageReference {
	return StoragePackageReference{
		LocalPackageReference{
			localPathPrefix: local.localPathPrefix,
			group:           local.group,
			version:         local.version + StoragePackageSuffix,
		},
	}
}

// String returns the string representation of the package reference
func (spr StoragePackageReference) String() string {
	return fmt.Sprintf("storage:%v", spr.PackagePath())
}

// IsPreview returns true if this package reference is a preview
func (spr StoragePackageReference) IsPreview() bool {
	lc := strings.ToLower(spr.version)
	return strings.Contains(lc, "alpha") ||
		strings.Contains(lc, "beta") ||
		strings.Contains(lc, "preview")
}

// IsStoragePackageReference returns true if the reference is to a storage package
func IsStoragePackageReference(reference PackageReference) bool {
	_, ok := reference.(StoragePackageReference)
	return ok
}
