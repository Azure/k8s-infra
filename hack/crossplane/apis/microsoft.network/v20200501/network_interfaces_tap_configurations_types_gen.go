// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
// Code generated by k8s-infra-gen. DO NOT EDIT.
package v20200501

import (
	"github.com/Azure/k8s-infra/hack/crossplane/apis/deploymenttemplate/v20150101"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
type NetworkInterfacesTapConfigurations struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              NetworkInterfacesTapConfigurations_Spec `json:"spec,omitempty"`
	Status            NetworkInterfaceTapConfiguration_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type NetworkInterfacesTapConfigurationsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkInterfacesTapConfigurations `json:"items"`
}

type NetworkInterfaceTapConfiguration_Status struct {
	AtProvider NetworkInterfacesTapConfigurationsObservation `json:"atProvider"`
}

type NetworkInterfacesTapConfigurations_Spec struct {
	ForProvider NetworkInterfacesTapConfigurationsParameters `json:"forProvider"`
}

type NetworkInterfacesTapConfigurationsObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the Virtual Network Tap configuration.
	Properties *NetworkInterfaceTapConfigurationPropertiesFormat_Status `json:"properties,omitempty"`

	//Type: Sub Resource type.
	Type *string `json:"type,omitempty"`
}

type NetworkInterfacesTapConfigurationsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion NetworkInterfacesTapConfigurationsSpecApiVersion `json:"apiVersion"`
	Comments   *string                                          `json:"comments,omitempty"`

	//Condition: Condition of the resource
	Condition *bool                   `json:"condition,omitempty"`
	Copy      *v20150101.ResourceCopy `json:"copy,omitempty"`

	//DependsOn: Collection of resources this resource depends on
	DependsOn []string `json:"dependsOn,omitempty"`

	//Location: Location to deploy resource to
	Location *v20150101.ResourceLocations `json:"location,omitempty"`

	// +kubebuilder:validation:Required
	//Name: Name of the resource
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	//Properties: Properties of the Virtual Network Tap configuration.
	Properties NetworkInterfaceTapConfigurationPropertiesFormat `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type NetworkInterfacesTapConfigurationsSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/NetworkInterfaceTapConfigurationPropertiesFormat
type NetworkInterfaceTapConfigurationPropertiesFormat struct {

	//VirtualNetworkTap: The reference to the Virtual Network Tap resource.
	VirtualNetworkTap *SubResource `json:"virtualNetworkTap,omitempty"`
}

//Generated from:
type NetworkInterfaceTapConfigurationPropertiesFormat_Status struct {

	//ProvisioningState: The provisioning state of the network interface tap
	//configuration resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//VirtualNetworkTap: The reference to the Virtual Network Tap resource.
	VirtualNetworkTap *VirtualNetworkTap_Status `json:"virtualNetworkTap,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type NetworkInterfacesTapConfigurationsSpecApiVersion string

const NetworkInterfacesTapConfigurationsSpecApiVersion20200501 = NetworkInterfacesTapConfigurationsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/networkInterfaces/tapConfigurations"}
type NetworkInterfacesTapConfigurationsSpecType string

const NetworkInterfacesTapConfigurationsSpecTypeMicrosoftNetworkNetworkInterfacesTapConfigurations = NetworkInterfacesTapConfigurationsSpecType("Microsoft.Network/networkInterfaces/tapConfigurations")

func init() {
	SchemeBuilder.Register(&NetworkInterfacesTapConfigurations{}, &NetworkInterfacesTapConfigurationsList{})
}
