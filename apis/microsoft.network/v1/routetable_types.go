/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package v1

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/k8s-infra/pkg/zips"
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
		APIVersion string `json:"apiVersion,omitempty"`

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
		DeploymentID      string `json:"deploymentId,omitempty"`
		ProvisioningState string `json:"provisioningState,omitempty"`
	}

	// +kubebuilder:object:root=true
	// +kubebuilder:subresource:status
	// +kubebuilder:storageversion

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

func (*RouteTable) Hub() {}

func (rt *RouteTable) ToResource() (zips.Resource, error) {
	rgName := ""
	if rt.Spec.ResourceGroup != nil {
		rgName = rt.Spec.ResourceGroup.Name
	}

	res := zips.Resource{
		ID:                rt.Status.ID,
		DeploymentID:      rt.Status.DeploymentID,
		Type:              "Microsoft.Network/routeTables",
		ResourceGroup:     rgName,
		Name:              rt.Name,
		APIVersion:        rt.Spec.APIVersion,
		Location:          rt.Spec.Location,
		Tags:              rt.Spec.Tags,
		ProvisioningState: zips.ProvisioningState(rt.Status.ProvisioningState),
	}

	bits, err := json.Marshal(rt.Spec.Properties)
	if err != nil {
		return res, err
	}
	res.Properties = bits

	return *res.SetAnnotations(rt.Annotations), nil
}

func (rt *RouteTable) FromResource(res zips.Resource) error {
	rt.Status.ID = res.ID
	rt.Status.DeploymentID = res.DeploymentID
	rt.Status.ProvisioningState = string(res.ProvisioningState)
	rt.Spec.Tags = res.Tags

	var props RouteTableSpecProperties
	if err := json.Unmarshal(res.Properties, &props); err != nil {
		return err
	}
	rt.Spec.Properties = &props
	return nil
}

func init() {
	SchemeBuilder.Register(&RouteTable{}, &RouteTableList{})
}
