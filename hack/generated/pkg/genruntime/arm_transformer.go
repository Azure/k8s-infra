/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package genruntime

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
)

type MetaObject interface {
	runtime.Object
	metav1.Object
	KubernetesResource
}

type KubernetesResource interface {
	// Owner returns the ResourceReference so that we can extract the Group/Kind for easier lookups
	Owner() *ResourceReference

	AzureName() string
}

type ArmResourceSpec interface {
	AzureApiVersion() string // TODO: This is technically not needed?
}

// TODO: Consider ArmSpecTransformer and ArmTransformer, so we don't have to pass owningName/name through all the calls

// ArmTransformer is a type which can be converted to/from an Arm object shape.
// Each CRD resource must implement these methods.
type ArmTransformer interface {
	// owningName is the "a/b/c" name for the owner. For example this would be "VNet1" when deploying subnet1
	// or myaccount when creating Batch pool1
	ConvertToArm(owningName string) (interface{}, error)
	PopulateFromArm(owner KnownResourceReference, input interface{}) error
}
//
//func LookupOwnerGroupKind(v MetaObject) (string, string) {
//	t := reflect.TypeOf(v)
//	field, _ := t.FieldByName("Spec")
//	group, ok := field.Tag.Lookup("group")
//	if !ok {
//		panic(fmt.Sprintf("Couldn't find %s owner group tag", v.GetName()))
//	}
//	kind, ok := field.Tag.Lookup("kind")
//	if !ok {
//		panic(fmt.Sprintf("Couldn't find %s owner kind tag", v.GetName()))
//	}
//
//	return group, kind
//}

// CombineArmNames creates a "fully qualified" resource name for use
// in an Azure template deployment.
// See https://docs.microsoft.com/en-us/azure/azure-resource-manager/templates/child-resource-name-type#outside-parent-resource
// for details on the format of the name field in ARM templates.
func CombineArmNames(names ...string) string {
	return strings.Join(names, "/")
}

// ExtractKubernetesResourceNameFromArmName extracts the Kubernetes resource name from an ARM name.
// See https://docs.microsoft.com/en-us/azure/azure-resource-manager/templates/child-resource-name-type#outside-parent-resource
// for details on the format of the name field in ARM templates.
func ExtractKubernetesResourceNameFromArmName(armName string) string {
	if len(armName) == 0 {
		return ""
	}

	// TODO: Possibly need to worry about preserving case here, although ARM should be already
	strs := strings.Split(armName, "/")
	return strs[len(strs)-1]
}
