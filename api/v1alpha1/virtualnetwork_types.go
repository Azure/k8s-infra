/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/k8s-infra/api/azmetav1"
)

type AddressSpace struct {
	AddressPrefixes []string `json:"addressPrefixes,omitempty"`
}

type VirtualNetworkProperties struct {
	AddressSpace AddressSpace `json:"addressSpace,omitempty"`
}

type SubnetProperties struct {
	AddressPrefix string `json:"addressPrefix,omitempty"`
}

type Subnet struct {
	azmetav1.NestedResourceSpec `json:",inline"`
}

// VirtualNetworkSpec defines the desired state of VirtualNetwork
type VirtualNetworkSpec struct {
	azmetav1.TrackedResourceSpec `json:",inline"`
}

// VirtualNetworkStatus defines the observed state of VirtualNetwork
type VirtualNetworkStatus struct {
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true

// VirtualNetwork is the Schema for the virtualnetworks API
type VirtualNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualNetworkSpec   `json:"spec,omitempty"`
	Status VirtualNetworkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VirtualNetworkList contains a list of VirtualNetwork
type VirtualNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualNetwork `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualNetwork{}, &VirtualNetworkList{})
}
