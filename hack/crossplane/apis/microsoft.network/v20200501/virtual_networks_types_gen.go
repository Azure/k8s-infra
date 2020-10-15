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
type VirtualNetworks struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VirtualNetworks_Spec  `json:"spec,omitempty"`
	Status            VirtualNetwork_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type VirtualNetworksList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualNetworks `json:"items"`
}

//Generated from:
type VirtualNetwork_Status struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Location: Resource location.
	Location *string `json:"location,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the virtual network.
	Properties *VirtualNetworkPropertiesFormat_Status `json:"properties,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type VirtualNetworks_Spec struct {
	ForProvider VirtualNetworksParameters `json:"forProvider"`
}

//Generated from:
type VirtualNetworkPropertiesFormat_Status struct {

	//AddressSpace: The AddressSpace that contains an array of IP address ranges that
	//can be used by subnets.
	AddressSpace *AddressSpace_Status `json:"addressSpace,omitempty"`

	//BgpCommunities: Bgp Communities sent over ExpressRoute with each route
	//corresponding to a prefix in this VNET.
	BgpCommunities *VirtualNetworkBgpCommunities_Status `json:"bgpCommunities,omitempty"`

	//DdosProtectionPlan: The DDoS protection plan associated with the virtual network.
	DdosProtectionPlan *SubResource_Status `json:"ddosProtectionPlan,omitempty"`

	//DhcpOptions: The dhcpOptions that contains an array of DNS servers available to
	//VMs deployed in the virtual network.
	DhcpOptions *DhcpOptions_Status `json:"dhcpOptions,omitempty"`

	//EnableDdosProtection: Indicates if DDoS protection is enabled for all the
	//protected resources in the virtual network. It requires a DDoS protection plan
	//associated with the resource.
	EnableDdosProtection *bool `json:"enableDdosProtection,omitempty"`

	//EnableVmProtection: Indicates if VM protection is enabled for all the subnets in
	//the virtual network.
	EnableVmProtection *bool `json:"enableVmProtection,omitempty"`

	//IpAllocations: Array of IpAllocation which reference this VNET.
	IpAllocations []SubResource_Status `json:"ipAllocations,omitempty"`

	//ProvisioningState: The provisioning state of the virtual network resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//ResourceGuid: The resourceGuid property of the Virtual Network resource.
	ResourceGuid *string `json:"resourceGuid,omitempty"`

	//Subnets: A list of subnets in a Virtual Network.
	Subnets []Subnet_Status `json:"subnets,omitempty"`

	//VirtualNetworkPeerings: A list of peerings in a Virtual Network.
	VirtualNetworkPeerings []VirtualNetworkPeering_Status `json:"virtualNetworkPeerings,omitempty"`
}

type VirtualNetworksParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion VirtualNetworksSpecApiVersion `json:"apiVersion"`
	Comments   *string                       `json:"comments,omitempty"`

	//Condition: Condition of the resource
	Condition *bool                   `json:"condition,omitempty"`
	Copy      *v20150101.ResourceCopy `json:"copy,omitempty"`

	//DependsOn: Collection of resources this resource depends on
	DependsOn []string `json:"dependsOn,omitempty"`

	//Location: Location to deploy resource to
	Location string `json:"location,omitempty"`

	// +kubebuilder:validation:Required
	//Name: Name of the resource
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	//Properties: Properties of the virtual network.
	Properties VirtualNetworkPropertiesFormat `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type VirtualNetworksSpecType `json:"type"`
}

//Generated from:
type DhcpOptions_Status struct {

	//DnsServers: The list of DNS servers IP addresses.
	DnsServers []string `json:"dnsServers,omitempty"`
}

//Generated from:
type VirtualNetworkBgpCommunities_Status struct {

	//RegionalCommunity: The BGP community associated with the region of the virtual
	//network.
	RegionalCommunity *string `json:"regionalCommunity,omitempty"`

	// +kubebuilder:validation:Required
	//VirtualNetworkCommunity: The BGP community associated with the virtual network.
	VirtualNetworkCommunity string `json:"virtualNetworkCommunity"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/VirtualNetworkPropertiesFormat
type VirtualNetworkPropertiesFormat struct {

	// +kubebuilder:validation:Required
	//AddressSpace: The AddressSpace that contains an array of IP address ranges that
	//can be used by subnets.
	AddressSpace AddressSpace `json:"addressSpace"`

	//BgpCommunities: Bgp Communities sent over ExpressRoute with each route
	//corresponding to a prefix in this VNET.
	BgpCommunities *VirtualNetworkBgpCommunities `json:"bgpCommunities,omitempty"`

	//DdosProtectionPlan: The DDoS protection plan associated with the virtual network.
	DdosProtectionPlan *SubResource `json:"ddosProtectionPlan,omitempty"`

	//DhcpOptions: The dhcpOptions that contains an array of DNS servers available to
	//VMs deployed in the virtual network.
	DhcpOptions *DhcpOptions `json:"dhcpOptions,omitempty"`

	//EnableDdosProtection: Indicates if DDoS protection is enabled for all the
	//protected resources in the virtual network. It requires a DDoS protection plan
	//associated with the resource.
	EnableDdosProtection *bool `json:"enableDdosProtection,omitempty"`

	//EnableVmProtection: Indicates if VM protection is enabled for all the subnets in
	//the virtual network.
	EnableVmProtection *bool `json:"enableVmProtection,omitempty"`

	//IpAllocations: Array of IpAllocation which reference this VNET.
	IpAllocations []SubResource `json:"ipAllocations,omitempty"`

	//Subnets: A list of subnets in a Virtual Network.
	Subnets []Subnet `json:"subnets,omitempty"`

	//VirtualNetworkPeerings: A list of peerings in a Virtual Network.
	VirtualNetworkPeerings []VirtualNetworkPeering `json:"virtualNetworkPeerings,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type VirtualNetworksSpecApiVersion string

const VirtualNetworksSpecApiVersion20200501 = VirtualNetworksSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/virtualNetworks"}
type VirtualNetworksSpecType string

const VirtualNetworksSpecTypeMicrosoftNetworkVirtualNetworks = VirtualNetworksSpecType("Microsoft.Network/virtualNetworks")

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/DhcpOptions
type DhcpOptions struct {

	// +kubebuilder:validation:Required
	//DnsServers: The list of DNS servers IP addresses.
	DnsServers []string `json:"dnsServers"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/VirtualNetworkBgpCommunities
type VirtualNetworkBgpCommunities struct {

	// +kubebuilder:validation:Required
	//VirtualNetworkCommunity: The BGP community associated with the virtual network.
	VirtualNetworkCommunity string `json:"virtualNetworkCommunity"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/VirtualNetworkPeering
type VirtualNetworkPeering struct {

	// +kubebuilder:validation:Required
	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name string `json:"name"`

	//Properties: Properties of the virtual network peering.
	Properties *VirtualNetworkPeeringPropertiesFormat `json:"properties,omitempty"`
}

func init() {
	SchemeBuilder.Register(&VirtualNetworks{}, &VirtualNetworksList{})
}
