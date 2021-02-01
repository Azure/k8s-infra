/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

// StorageConversionContext captures additional supporting information that may be needed when a
// storage conversion factory creates a conversion
type StorageConversionContext struct {
	// types is a map of all known type definitions, used to resolve TypeNames to actual types
	types Types
}

// NewStorageConversionContext creates a new instance of a StorageConversionContext
func NewStorageConversionContext(types Types) *StorageConversionContext {
	return &StorageConversionContext{
		types: types,
	}
}
