/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"github.com/gobuffalo/flect"
	"strconv"
	"strings"
	"unicode"
)

type KnownLocalsSet map[string]struct{}

// StorageConversionEndpoint represents either a source or a destination field for a storage conversion
type StorageConversionEndpoint struct {
	// theType is the Type of the value accessible via this endpoint
	theType Type
	// name is the name of the underlying property, used to generate useful local identifiers
	name string
	// knownLocals is a shared map of locals that have already been created within a given function, to prevent duplicates
	knownLocals KnownLocalsSet
}

func NewStorageConversionEndpoint(theType Type, name string, knownLocals KnownLocalsSet) *StorageConversionEndpoint {
	return &StorageConversionEndpoint{
		theType:     theType,
		name:        strings.ToLower(name),
		knownLocals: knownLocals,
	}
}

// Expr returns a clone of the expression for this endpoint
// (returning a clone avoids issues with reuse of fragments within the dst)
func (endpoint *StorageConversionEndpoint) Type() Type {
	return endpoint.theType
}

// Create an identifier for a local variable that represents a single item
// Each call will return a unique identifier
func (endpoint *StorageConversionEndpoint) CreateSingularLocal() string {
	singular := flect.Singularize(endpoint.name)
	return endpoint.knownLocals.createLocal(singular)
}

// CreatePluralLocal creates an identifier for a local variable that represents multiple items
// Each call will return a unique identifier
func (endpoint *StorageConversionEndpoint) CreatePluralLocal(suffix string) string {
	plural := flect.Pluralize(endpoint.name)
	return endpoint.knownLocals.createLocal(plural + suffix)
}

// WithType creates a new endpoint with a different type
func (endpoint *StorageConversionEndpoint) WithType(theType Type) *StorageConversionEndpoint {
	return NewStorageConversionEndpoint(
		theType,
		endpoint.name,
		endpoint.knownLocals)
}

// createLocal creates a new unique local with the specified suffix
// Has to be deterministic, so we use an incrementing number to make them unique
func (locals KnownLocalsSet) createLocal(nameHint string) string {
	baseName := locals.toPrivate(nameHint)
	id := baseName
	index := 0
	for {
		_, found := locals[id]
		if !found {
			break
		}

		index++
		id = baseName + strconv.Itoa(index)
	}

	locals[id] = struct{}{}

	return id
}

// Add allows identifiers that have already been used to be registered, avoiding duplicates
func (locals KnownLocalsSet) Add(local string) {
	name := locals.toPrivate(local)
	locals[name] = struct{}{}
}

// toPrivate converts a Go identifier into a private form
func (locals KnownLocalsSet) toPrivate(s string) string {
	// Just lowercase the first character
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}
