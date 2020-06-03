/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package config

import (
	"fmt"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"regexp"
)

type TransformTarget struct {
	PackagePath string `yaml:",omitempty"`
	Name        string `yaml:",omitempty"`
}

// A TypeTransformer is used to remap types
type TypeTransformer struct {
	// Group is a wildcard matching specifier for which groups are selected by this transformer
	Group      string `yaml:",omitempty"`
	groupRegex *regexp.Regexp
	// Version is a wildcard matching specifier for which types are selected by this transformer
	Version      string `yaml:",omitempty"`
	versionRegex *regexp.Regexp
	// Name is a wildcard matching specifier for which types are selected by this transformer
	Name      string `yaml:",omitempty"`
	nameRegex *regexp.Regexp
	// Because is used to articulate why the targetType applied to a type (used to generate explanatory logs in debug mode)
	Because string

	Transform  TransformTarget // TODO: Don't love this property name or type name...
	targetType astmodel.Type
}

func (transformer *TypeTransformer) groupMatches(schema string) bool {
	return transformer.matches(transformer.Group, transformer.groupRegex, schema)
}

func (transformer *TypeTransformer) versionMatches(version string) bool {
	return transformer.matches(transformer.Version, transformer.versionRegex, version)
}

func (transformer *TypeTransformer) nameMatches(name string) bool {
	return transformer.matches(transformer.Name, transformer.nameRegex, name)
}

func (transformer *TypeTransformer) matches(glob string, regex *regexp.Regexp, name string) bool {
	if glob == "" {
		return true
	}

	return regex.MatchString(name)
}

func (transformer *TypeTransformer) Init() error {
	if transformer.Transform.Name == "" {
		return fmt.Errorf(
			"type transformer for group: %s, version: %s, name: %s is missing type name to transform to",
			transformer.Group,
			transformer.Version,
			transformer.Name)
	}

	transformer.groupRegex = createGlobbingRegex(transformer.Group)
	transformer.versionRegex = createGlobbingRegex(transformer.Version)
	transformer.nameRegex = createGlobbingRegex(transformer.Name)

	// Type has no package -- must be a primitive type
	if transformer.Transform.PackagePath == "" {
		switch transformer.Transform.Name {
		case "bool":
			transformer.targetType = astmodel.BoolType
		case "float":
			transformer.targetType = astmodel.FloatType
		case "int":
			transformer.targetType = astmodel.IntType
		case "string":
			transformer.targetType = astmodel.StringType
		default:
			return fmt.Errorf(
				"type transformer for group: %s, version: %s, name: %s has unknown"+
					"primtive type transformation target: %s",
				transformer.Group,
				transformer.Version,
				transformer.Name,
				transformer.Transform.Name)

		}

		return nil
	}

	transformer.targetType = astmodel.NewTypeName(
		*astmodel.NewPackageReference(transformer.Transform.PackagePath),
		transformer.Transform.Name)
	return nil
}

func (transformer *TypeTransformer) TransformTypeName(typeName *astmodel.TypeName) astmodel.Type {
	name := typeName.Name()

	if typeName.PackageReference.IsLocalPackage() {
		group, version, err := typeName.PackageReference.GroupAndPackage()
		if err != nil {
			// This shouldn't ever happen because IsLocalPackage is true -- checking just to be safe
			panic(fmt.Sprintf("%s was flagged as a local package but has no group and package", typeName.PackageReference))
		}

		if transformer.groupMatches(group) && transformer.versionMatches(version) && transformer.nameMatches(name) {
			return transformer.targetType
		}
	} else {
		// TODO: Support external types better rather than doing everything in terms of GVK?
		if transformer.nameMatches(name) {
			return transformer.targetType
		}
	}

	// Didn't match so return nil
	return nil
}
