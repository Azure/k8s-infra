/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package config

import (
	"fmt"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"regexp"
	"strings"
)

// Filter contains basic functionality for a filter
type Filter struct {
	// Group is a wildcard matching specifier for which groups are selected by this filter
	Group      string `yaml:",omitempty"`
	groupRegex *regexp.Regexp
	// Version is a wildcard matching specifier for which types are selected by this filter
	Version      string `yaml:",omitempty"`
	versionRegex *regexp.Regexp
	// Name is a wildcard matching specifier for which types are selected by this filter
	Name      string `yaml:",omitempty"`
	nameRegex *regexp.Regexp
	// Because is used to articulate why the filter applied to a type (used to generate explanatory logs in debug mode)
	Because string
}

func (filter *Filter) groupMatches(schema string) bool {
	return filter.matches(filter.Group, &filter.groupRegex, schema)
}

func (filter *Filter) versionMatches(version string) bool {
	return filter.matches(filter.Version, &filter.versionRegex, version)
}

func (filter *Filter) nameMatches(name string) bool {
	return filter.matches(filter.Name, &filter.nameRegex, name)
}

func (filter *Filter) matches(glob string, regex **regexp.Regexp, name string) bool {
	if glob == "" {
		return true
	}

	if *regex == nil {
		*regex = createGlobbingRegex(glob)
	}

	return (*regex).MatchString(name)
}

// AppliesToType indicates whether this filter should be applied to the supplied type definition
func (filter *Filter) AppliesToType(typeName *astmodel.TypeName) bool {
	groupName, packageName, err := typeName.PackageReference.GroupAndPackage()
	if err != nil {
		// TODO: Should this func return an error rather than panic?
		panic(fmt.Sprintf("%v", err))
	}

	result := filter.groupMatches(groupName) &&
		filter.versionMatches(packageName) &&
		filter.nameMatches(typeName.Name())

	return result
}

// create a regex that does globbing of names
// * and ? have their usual (DOS style) meanings as wildcards
func createGlobbingRegex(globbing string) *regexp.Regexp {
	if globbing == "" {
		// nil here as "" is fast-tracked elsewhere
		return nil
	}

	g := regexp.QuoteMeta(globbing)
	g = strings.ReplaceAll(g, "\\*", ".*")
	g = strings.ReplaceAll(g, "\\?", ".")
	g = "^" + g + "$"
	return regexp.MustCompile(g)
}
