/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package genruntime

import (
	"strings"
)

type ArmTransformer interface {
	// TODO: if a transform can never fail... then don't bother returning error here?
	// owningName is the "a/b/c" name for the owner. For example this would be "VNet1" when deploying subnet1
	// or myaccount when creating Batch pool1
	ToArm(owningName string) (interface{}, error)
	FromArm(owner KnownResourceReference, input interface{}) error
}

func CreateArmResourceNameForDeployment(owningName string, name string) string {
	result := owningName + "/" + name
	return result
}

func ExtractKubernetesResourceNameFromArmName(armName string) string {
	// TODO: Possibly need to worry about preserving case here, although ARM should be already
	strs := strings.Split(armName, "/")
	return strs[len(strs)-1]
}
