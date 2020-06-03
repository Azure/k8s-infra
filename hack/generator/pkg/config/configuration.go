/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package config

import (
	"errors"
	"fmt"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
)

// Configuration is used to control which types get generated
type Configuration struct {
	// Base URL for the JSON schema to generate
	SchemaURL string
	// Filters used to control which types are included
	ExportFilters []*ExportFilter
	// Filters used to control which types are included in the type graph during JSON schema parsing
	TypeFilters []*TypeFilter
	// TypeTransformers used to remap types
	TypeTransformers []*TypeTransformer
}

// ShouldExportResult is returned by ShouldExport to indicate whether the supplied type should be exported
type ShouldExportResult string

const (
	// Export indicates the specified type should be exported to disk
	Export ShouldExportResult = "export"
	// Skip indicates the specified type should be skipped and not exported
	Skip ShouldExportResult = "skip"
)

// ShouldPruneResult is returned by ShouldPrune to indicate whether the supplied type should be exported
type ShouldPruneResult string

const (
	// Include indicates the specified type should be included in the type graph
	Include ShouldPruneResult = "include"
	// Prune indicates the type (and all types only referenced by it) should be pruned from the type graph
	Prune ShouldPruneResult = "prune"
)

// NewConfiguration is a convenience factory for Configuration
func NewConfiguration(filters ...*ExportFilter) *Configuration {
	result := Configuration{
		ExportFilters: filters,
	}

	return &result
}

// Initialize checks for common errors and initializes structures inside the configuration
// which need additional setup after json deserialization
func (config *Configuration) Initialize() error {
	if config.SchemaURL == "" {
		return errors.New("SchemaURL missing")
	}

	for _, transformer := range config.TypeTransformers {
		err := transformer.Init()
		if err != nil {
			return err
		}
	}

	return nil
}

// ShouldExport tests for whether a given struct should be exported
// Returns a result indicating whether export should occur as well as a reason for logging
func (config *Configuration) ShouldExport(typeName *astmodel.TypeName) (result ShouldExportResult, because string) {
	for _, f := range config.ExportFilters {
		if f.AppliesToType(typeName) {
			switch f.Action {
			case ExportFilterActionExclude:
				return Skip, f.Because
			case ExportFilterActionInclude:
				return Export, f.Because
			default:
				panic(fmt.Errorf("unknown exportfilter directive: %s", f.Action))
			}
		}
	}

	// By default we export all types
	return Export, ""
}

// ShouldPrune tests for whether a given type should be processed or pruned
func (config *Configuration) ShouldPrune(typeName *astmodel.TypeName) (result ShouldPruneResult, because string) {
	for _, f := range config.TypeFilters {
		if f.AppliesToType(typeName) {
			switch f.Action {
			case TypeFilterPruneType:
				return Prune, f.Because
			case TypeFilterIncludeType:
				return Include, f.Because
			default:
				panic(fmt.Errorf("unknown typefilter directive: %s", f.Action))
			}
		}
	}

	// By default we export all types
	return Include, ""
}

// TransformType uses the configured type transformers to transform a type name (reference) to a different type.
// If no transformation is performed, nil is returned
func (config *Configuration) TransformType(name *astmodel.TypeName) (astmodel.Type, string) {
	for _, transformer := range config.TypeTransformers {
		result := transformer.TransformTypeName(name)
		if result != nil {
			return result, transformer.Because
		}
	}

	// No matches, return nil
	return nil, ""
}
