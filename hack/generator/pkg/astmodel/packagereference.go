/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import "path/filepath"

// Package reference indicates which package
// a struct belongs to.
type PackageReference struct {
	groupName   string
	packageName string
}

func (pr *PackageReference) PackagePath() string {
	return filepath.Join(pr.GroupName(), pr.PackageName())
}

func (pr *PackageReference) GroupName() string {
	return pr.groupName
}

func (pr *PackageReference) PackageName() string {
	return pr.packageName
}

// Equals returns true if the passed package reference references the same package, false otherwise
func (pr *PackageReference) Equals(ref *PackageReference) bool {
	return pr.groupName == ref.groupName && pr.packageName == pr.packageName
}
