package astmodel

// EnumValue captures a single value of the enumeration
type EnumValue struct {
	// Identifier is a Go identifer for the value
	Identifier string
	// Value is the actual value expected by ARM
	Value string
}

func (value *EnumValue) Equals(v *EnumValue) bool {
	if value == v {
		return true
	}

	return value.Identifier == v.Identifier && value.Value == v.Value
}