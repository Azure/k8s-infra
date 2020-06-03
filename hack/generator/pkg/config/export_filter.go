/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package config

// ExportFilterAction defines the possible actions that should happen for types matching the filter
type ExportFilterAction string

const (
	// ExportFilterActionInclude indicates that any type matched by the filter should be exported to disk by the generator
	ExportFilterActionInclude ExportFilterAction = "include"

	// ExportFilterActionExclude indicates that any type matched by the filter should not be exported as a struct.
	// This skips generation of types but may not prevent references to the type by other structs (which
	// will cause compiler errors).
	ExportFilterActionExclude ExportFilterAction = "exclude"
)

// A ExportFilter is used to control which types should be exported by the generator
type ExportFilter struct {
	Action ExportFilterAction
	Filter `yaml:",inline"`
}
