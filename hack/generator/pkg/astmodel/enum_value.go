/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

// EnumValue captures a single value of the enumeration
type EnumValue struct {
	// Identifier is a Go identifier for the value
	Identifier string
	// Value is the actual value expected by ARM
	Value string
}
