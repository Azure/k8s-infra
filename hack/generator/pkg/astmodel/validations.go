/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"
	"reflect"
	"strings"
)

type Validation struct {
	name string

	// value is a bool, an int, or a string, or a list of those
	value interface{}
}

func GenerateKubebuilderComment(validation Validation) string {
	const prefix = "// +kubebuilder:validation:"

	if validation.value != nil {
		value := reflect.ValueOf(validation.value)

		if value.Kind() == reflect.Slice {
			// handle slice values which should look like "x,y,z"
			var values []string
			for i := 0; i < value.Len(); i += 1 {
				values = append(values, fmt.Sprintf("%v", value.Index(i)))
			}

			return fmt.Sprintf("%s%s=%s", prefix, validation.name, strings.Join(values, ","))
		}

		// everything else
		return fmt.Sprintf("%s%s=%v", prefix, validation.name, validation.value)
	}

	// validation without argument
	return fmt.Sprintf("%s%s", prefix, validation.name)
}

func ValidateEnum(permittedValues []interface{}) Validation {
	return Validation{"Enum", permittedValues}
}

func ValidateRequired() Validation {
	return Validation{"Required", nil}
}
