/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package v20191101

import (
	"encoding/json"

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

	// RouteSpec defines the desired state of Route
	RouteSpec struct {
		// Properties of the subnet
		Properties *RouteSpecProperties `json:"properties,omitempty"`
	}

	// RouteStatus defines the observed state of Route
	RouteStatus struct {
		ID                string `json:"id,omitempty"`
		ProvisioningState string `json:"provisioningState,omitempty"`
	}

	// +kubebuilder:object:root=true

	// Route is the Schema for the routes API
	Route struct {
		metav1.TypeMeta   `json:",inline"`
		metav1.ObjectMeta `json:"metadata,omitempty"`

		Spec   RouteSpec   `json:"spec,omitempty"`
		Status RouteStatus `json:"status,omitempty"`
	}

	// +kubebuilder:object:root=true

	// RouteList contains a list of Route
	RouteList struct {
		metav1.TypeMeta `json:",inline"`
		metav1.ListMeta `json:"metadata,omitempty"`
		Items           []Route `json:"items"`
	}
)

func (r *Route) ConvertTo(dstRaw conversion.Hub) error {
	to := dstRaw.(*v1.RouteTable)
	to.ObjectMeta = r.ObjectMeta
	to.Spec.APIVersion = "2019-11-01"
	to.Status.ID = r.Status.ID
	to.Status.ProvisioningState = r.Status.ProvisioningState
	bits, err := json.Marshal(r.Spec.Properties)
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

func (r *Route) ConvertFrom(src conversion.Hub) error {
	from := src.(*v1.Route)
	r.ObjectMeta = from.ObjectMeta
	r.Status.ID = from.Status.ID
	r.Status.ProvisioningState = from.Status.ProvisioningState

	bits, err := json.Marshal(from.Spec.Properties)
	if err != nil {
		return err
	}

	var props RouteSpecProperties
	if err := json.Unmarshal(bits, &props); err != nil {
		return err
	}
	r.Spec.Properties = &props
	return nil
}

func init() {
	SchemeBuilder.Register(&Route{}, &RouteList{})
}
