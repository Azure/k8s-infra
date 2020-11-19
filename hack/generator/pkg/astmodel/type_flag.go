package astmodel

import "fmt"

type TypeFlag string

const (
	StorageFlag = TypeFlag("storage")
	ArmFlag     = TypeFlag("arm")
	OneOfFlag   = TypeFlag("oneof")
)

var _ fmt.Stringer = TypeFlag("")

func (f TypeFlag) String() string {
	return string(f)
}
