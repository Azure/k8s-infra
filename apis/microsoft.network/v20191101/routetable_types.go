/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package v20191101

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	v1 "github.com/Azure/k8s-infra/apis/microsoft.network/v1"
)

type (
	// RouteSpecProperties are the resource specific properties
	RouteSpecProperties struct {
		AddressPrefix    string `json:"addressPrefix,omitempty"`
		NextHopIPAddress string `json:"nextHopIpAddress,omitempty"`
		// +kubebuilder:validation:Enum=Internet;None;VirtualAppliance;VirtualNetworkGateway;VnetLocal
		NextHopType string `json:"nextHopType,omitempty"`
	}

	// RouteSpec is a route resource
	// TODO: (dj) I think this should probably be a slice of corev1.ObjectReference
	RouteSpec struct {
		// ID of the subnet resource
		// +kubebuilder:validation:Required
		ID string `json:"id,omitempty"`

		// Name of the subnet
		// +kubebuilder:validation:Required
		Name string `json:"name,omitempty"`

		// Properties of the subnet
		Properties *RouteSpecProperties `json:"properties,omitempty"`
	}

	// RouteTableSpecProperties are the resource specific properties
	RouteTableSpecProperties struct {
		DisableBGPRoutePropagation bool        `json:"disableBgpRoutePropagation,omitempty"`
		Routes                     []RouteSpec `json:"routes,omitempty"`
	}

	// RouteTableSpec defines the desired state of RouteTable
	RouteTableSpec struct {
		// ResourceGroup is the Azure Resource Group the VirtualNetwork resides within
		// +kubebuilder:validation:Required
		ResourceGroup *corev1.ObjectReference `json:"group"`

		// Location of the VNET in Azure
		// +kubebuilder:validation:Required
		Location string `json:"location"`

		// Tags are user defined key value pairs
		// +optional
		Tags map[string]string `json:"tags,omitempty"`

		// Properties of the Virtual Network
		Properties *RouteTableSpecProperties `json:"properties,omitempty"`
	}

	// RouteTableStatus defines the observed state of RouteTable
	RouteTableStatus struct {
		ID                string `json:"id,omitempty"`
		ProvisioningState string `json:"provisioningState,omitempty"`
	}

	// +kubebuilder:object:root=true

	// RouteTable is the Schema for the routetables API
	RouteTable struct {
		metav1.TypeMeta   `json:",inline"`
		metav1.ObjectMeta `json:"metadata,omitempty"`

		Spec   RouteTableSpec   `json:"spec,omitempty"`
		Status RouteTableStatus `json:"status,omitempty"`
	}

	// +kubebuilder:object:root=true

	// RouteTableList contains a list of RouteTable
	RouteTableList struct {
		metav1.TypeMeta `json:",inline"`
		metav1.ListMeta `json:"metadata,omitempty"`
		Items           []RouteTable `json:"items"`
	}
)

func (rt *RouteTable) ConvertTo(dstRaw conversion.Hub) error {
	to := dstRaw.(*v1.RouteTable)
	to.ObjectMeta = rt.ObjectMeta
	to.Spec.ResourceGroup = rt.Spec.ResourceGroup
	to.Spec.APIVersion = "2019-11-01"
	to.Spec.Location = rt.Spec.Location
	to.Spec.Tags = rt.Spec.Tags
	to.Status.ID = rt.Status.ID
	to.Status.ProvisioningState = rt.Status.ProvisioningState
	bits, err := json.Marshal(rt.Spec.Properties)
	if err != nil {
		return err
	}

	var props v1.RouteTableSpecProperties
	if err := json.Unmarshal(bits, &props); err != nil {
		return err
	}

	to.Spec.Properties = &props
	return nil
}

func (rt *RouteTable) ConvertFrom(src conversion.Hub) error {
	from := src.(*v1.RouteTable)
	rt.ObjectMeta = from.ObjectMeta
	rt.Spec.ResourceGroup = from.Spec.ResourceGroup
	rt.Spec.Location = from.Spec.Location
	rt.Spec.Tags = from.Spec.Tags
	rt.Status.ID = from.Status.ID
	rt.Status.ProvisioningState = from.Status.ProvisioningState

	bits, err := json.Marshal(from.Spec.Properties)
	if err != nil {
		return err
	}

	var props RouteTableSpecProperties
	if err := json.Unmarshal(bits, &props); err != nil {
		return err
	}
	rt.Spec.Properties = &props
	return nil
}

func init() {
	SchemeBuilder.Register(&RouteTable{}, &RouteTableList{})
}
