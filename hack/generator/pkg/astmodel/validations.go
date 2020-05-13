package astmodel

import (
	"fmt"
	"reflect"
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
			valueString := ""
			for i := 0; i < value.Len(); i += 1 {
				if i != 0 {
					valueString += ","
				}

				valueString += fmt.Sprintf("%v", value.Index(i))
			}

			return fmt.Sprintf("%s%s=%s", prefix, validation.name, valueString)
		} else {
			// everything else
			return fmt.Sprintf("%s%s=%v", prefix, validation.name, validation.value)
		}
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
