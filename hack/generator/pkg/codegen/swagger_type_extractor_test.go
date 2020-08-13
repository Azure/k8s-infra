/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
)

func Example_inferNameFromURLPath() {
	group, name, _ := inferNameFromURLPath("/Microsoft.GroupName/resourceName/{resourceId}")
	fmt.Printf("%s: %s", group, name)
	// Output: Microsoft.GroupName: ResourceName
}

func TestInferNameFromURLPath_FailsWithMultipleParametersInARow(t *testing.T) {
	g := NewGomegaWithT(t)

	_, _, err := inferNameFromURLPath("/Microsoft.GroupName/resourceName/{resourceId}/{anotherParameter}")
	g.Expect(err).To(Not(BeNil()))
	g.Expect(err.Error()).To(ContainSubstring("multiple parameters"))
}

func TestInferNameFromURLPath_FailsWithNoGroupName(t *testing.T) {
	g := NewGomegaWithT(t)

	_, _, err := inferNameFromURLPath("/resourceName/{resourceId}/{anotherParameter}")
	g.Expect(err).To(Not(BeNil()))
	g.Expect(err.Error()).To(ContainSubstring("no group name"))
}
