// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
// Code generated by k8s-infra-gen. DO NOT EDIT.
package v20200501

import (
	"fmt"
	"github.com/Azure/k8s-infra/hack/generated/apis/deploymenttemplate/v20150101"
	"github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
type VirtualHubsIpConfigurations struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VirtualHubsIpConfigurations_Spec `json:"spec,omitempty"`
	Status            HubIpConfiguration_Status        `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type VirtualHubsIpConfigurationsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualHubsIpConfigurations `json:"items"`
}

//Generated from:
type HubIpConfiguration_Status struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: Name of the Ip Configuration.
	Name *string `json:"name,omitempty"`

	//Properties: The properties of the Virtual Hub IPConfigurations.
	Properties *HubIPConfigurationPropertiesFormat_Status `json:"properties,omitempty"`

	//Type: Ipconfiguration type.
	Type *string `json:"type,omitempty"`
}

type VirtualHubsIpConfigurations_Spec struct {
	ForProvider VirtualHubsIpConfigurationsParameters `json:"forProvider"`
}

//Generated from:
type HubIPConfigurationPropertiesFormat_Status struct {

	//PrivateIPAddress: The private IP address of the IP configuration.
	PrivateIPAddress *string `json:"privateIPAddress,omitempty"`

	//PrivateIPAllocationMethod: The private IP address allocation method.
	PrivateIPAllocationMethod *IPAllocationMethod_Status `json:"privateIPAllocationMethod,omitempty"`

	//ProvisioningState: The provisioning state of the IP configuration resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//PublicIPAddress: The reference to the public IP resource.
	PublicIPAddress *PublicIPAddress_Status `json:"publicIPAddress,omitempty"`

	//Subnet: The reference to the subnet resource.
	Subnet *Subnet_Status `json:"subnet,omitempty"`
}

type VirtualHubsIpConfigurationsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion VirtualHubsIpConfigurationsSpecApiVersion `json:"apiVersion"`
	Comments   *string                                   `json:"comments,omitempty"`

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
	//Properties: The properties of the Virtual Hub IPConfigurations.
	Properties HubIPConfigurationPropertiesFormat `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type VirtualHubsIpConfigurationsSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/HubIPConfigurationPropertiesFormat
type HubIPConfigurationPropertiesFormat struct {

	//PrivateIPAddress: The private IP address of the IP configuration.
	PrivateIPAddress *string `json:"privateIPAddress,omitempty"`

	//PrivateIPAllocationMethod: The private IP address allocation method.
	PrivateIPAllocationMethod *HubIPConfigurationPropertiesFormatPrivateIPAllocationMethod `json:"privateIPAllocationMethod,omitempty"`

	//PublicIPAddress: The reference to the public IP resource.
	PublicIPAddress *PublicIPAddress `json:"publicIPAddress,omitempty"`

	//Subnet: The reference to the subnet resource.
	Subnet *Subnet `json:"subnet,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type VirtualHubsIpConfigurationsSpecApiVersion string

const VirtualHubsIpConfigurationsSpecApiVersion20200501 = VirtualHubsIpConfigurationsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/virtualHubs/ipConfigurations"}
type VirtualHubsIpConfigurationsSpecType string

const VirtualHubsIpConfigurationsSpecTypeMicrosoftNetworkVirtualHubsIpConfigurations = VirtualHubsIpConfigurationsSpecType("Microsoft.Network/virtualHubs/ipConfigurations")

// +kubebuilder:validation:Enum={"Dynamic","Static"}
type HubIPConfigurationPropertiesFormatPrivateIPAllocationMethod string

const (
	HubIPConfigurationPropertiesFormatPrivateIPAllocationMethodDynamic = HubIPConfigurationPropertiesFormatPrivateIPAllocationMethod("Dynamic")
	HubIPConfigurationPropertiesFormatPrivateIPAllocationMethodStatic  = HubIPConfigurationPropertiesFormatPrivateIPAllocationMethod("Static")
)

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/PublicIPAddress
type PublicIPAddress struct {

	// +kubebuilder:validation:Required
	//Location: Resource location.
	Location string `json:"location"`

	//Properties: Public IP address properties.
	Properties *PublicIPAddressPropertiesFormat `json:"properties,omitempty"`

	//Sku: The public IP address SKU.
	Sku *PublicIPAddressSku `json:"sku,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Zones: A list of availability zones denoting the IP allocated for the resource
	//needs to come from.
	Zones []string `json:"zones,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/Subnet
type Subnet struct {

	// +kubebuilder:validation:Required
	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name string `json:"name"`

	//Properties: Properties of the subnet.
	Properties *SubnetPropertiesFormat `json:"properties,omitempty"`
}

func init() {
	SchemeBuilder.Register(&VirtualHubsIpConfigurations{}, &VirtualHubsIpConfigurationsList{})
}
