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
type IpAllocations struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              IpAllocations_Spec  `json:"spec,omitempty"`
	Status            IpAllocation_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type IpAllocationsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IpAllocations `json:"items"`
}

//Generated from:
type IpAllocation_Status struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Location: Resource location.
	Location *string `json:"location,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the IpAllocation.
	Properties *IpAllocationPropertiesFormat_Status `json:"properties,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type IpAllocations_Spec struct {
	ForProvider IpAllocationsParameters `json:"forProvider"`
}

//Generated from:
type IpAllocationPropertiesFormat_Status struct {

	//AllocationTags: IpAllocation tags.
	AllocationTags map[string]string `json:"allocationTags,omitempty"`

	//IpamAllocationId: The IPAM allocation ID.
	IpamAllocationId *string `json:"ipamAllocationId,omitempty"`

	//Prefix: The address prefix for the IpAllocation.
	Prefix *string `json:"prefix,omitempty"`

	//PrefixLength: The address prefix length for the IpAllocation.
	PrefixLength *int `json:"prefixLength,omitempty"`

	//PrefixType: The address prefix Type for the IpAllocation.
	PrefixType *IPVersion_Status `json:"prefixType,omitempty"`

	//Subnet: The Subnet that using the prefix of this IpAllocation resource.
	Subnet *SubResource_Status `json:"subnet,omitempty"`

	//Type: The type for the IpAllocation.
	Type *IpAllocationType_Status `json:"type,omitempty"`

	//VirtualNetwork: The VirtualNetwork that using the prefix of this IpAllocation
	//resource.
	VirtualNetwork *SubResource_Status `json:"virtualNetwork,omitempty"`
}

type IpAllocationsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion IpAllocationsSpecApiVersion `json:"apiVersion"`
	Comments   *string                     `json:"comments,omitempty"`

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
	//Properties: Properties of the IpAllocation.
	Properties IpAllocationPropertiesFormat `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type IpAllocationsSpecType `json:"type"`
}

//Generated from:
// +kubebuilder:validation:Enum={"IPv4","IPv6"}
type IPVersion_Status string

const (
	IPVersion_StatusIPv4 = IPVersion_Status("IPv4")
	IPVersion_StatusIPv6 = IPVersion_Status("IPv6")
)

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/IpAllocationPropertiesFormat
type IpAllocationPropertiesFormat struct {

	//AllocationTags: IpAllocation tags.
	AllocationTags map[string]string `json:"allocationTags,omitempty"`

	//IpamAllocationId: The IPAM allocation ID.
	IpamAllocationId *string `json:"ipamAllocationId,omitempty"`

	//Prefix: The address prefix for the IpAllocation.
	Prefix *string `json:"prefix,omitempty"`

	//PrefixLength: The address prefix length for the IpAllocation.
	PrefixLength *int `json:"prefixLength,omitempty"`

	//PrefixType: The address prefix Type for the IpAllocation.
	PrefixType *IpAllocationPropertiesFormatPrefixType `json:"prefixType,omitempty"`

	//Type: The type for the IpAllocation.
	Type *IpAllocationPropertiesFormatType `json:"type,omitempty"`
}

//Generated from:
// +kubebuilder:validation:Enum={"Hypernet","Undefined"}
type IpAllocationType_Status string

const (
	IpAllocationType_StatusHypernet  = IpAllocationType_Status("Hypernet")
	IpAllocationType_StatusUndefined = IpAllocationType_Status("Undefined")
)

// +kubebuilder:validation:Enum={"2020-05-01"}
type IpAllocationsSpecApiVersion string

const IpAllocationsSpecApiVersion20200501 = IpAllocationsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/IpAllocations"}
type IpAllocationsSpecType string

const IpAllocationsSpecTypeMicrosoftNetworkIpAllocations = IpAllocationsSpecType("Microsoft.Network/IpAllocations")

// +kubebuilder:validation:Enum={"IPv4","IPv6"}
type IpAllocationPropertiesFormatPrefixType string

const (
	IpAllocationPropertiesFormatPrefixTypeIPv4 = IpAllocationPropertiesFormatPrefixType("IPv4")
	IpAllocationPropertiesFormatPrefixTypeIPv6 = IpAllocationPropertiesFormatPrefixType("IPv6")
)

// +kubebuilder:validation:Enum={"Hypernet","Undefined"}
type IpAllocationPropertiesFormatType string

const (
	IpAllocationPropertiesFormatTypeHypernet  = IpAllocationPropertiesFormatType("Hypernet")
	IpAllocationPropertiesFormatTypeUndefined = IpAllocationPropertiesFormatType("Undefined")
)

func init() {
	SchemeBuilder.Register(&IpAllocations{}, &IpAllocationsList{})
}
