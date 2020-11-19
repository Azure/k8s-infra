package astmodel

import "fmt"

type TypeFlag string

const (
	StorageFlag = TypeFlag("storage")
	ArmFlag     = TypeFlag("arm")
	OneOfFlag   = TypeFlag("oneof")
)

var _ fmt.Stringer = TypeFlag("")

// String() renders the tag as a string
func (f TypeFlag) String() string {
	return string(f)
}

// ApplyTo() applies the tag to the provided type
func (f TypeFlag) ApplyTo(t Type) *FlaggedType {
	return NewFlaggedType(t, f)
}
