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
type PublicIPPrefixes struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PublicIPPrefixes_Spec `json:"spec,omitempty"`
	Status            PublicIPPrefix_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type PublicIPPrefixesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PublicIPPrefixes `json:"items"`
}

type PublicIPPrefix_Status struct {
	AtProvider PublicIPPrefixesObservation `json:"atProvider"`
}

type PublicIPPrefixes_Spec struct {
	ForProvider PublicIPPrefixesParameters `json:"forProvider"`
}

type PublicIPPrefixesObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Location: Resource location.
	Location *string `json:"location,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Public IP prefix properties.
	Properties *PublicIPPrefixPropertiesFormat_Status `json:"properties,omitempty"`

	//Sku: The public IP prefix SKU.
	Sku *PublicIPPrefixSku_Status `json:"sku,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`

	//Zones: A list of availability zones denoting the IP allocated for the resource
	//needs to come from.
	Zones []string `json:"zones,omitempty"`
}

type PublicIPPrefixesParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion PublicIPPrefixesSpecApiVersion `json:"apiVersion"`
	Comments   *string                        `json:"comments,omitempty"`

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
	//Properties: Public IP prefix properties.
	Properties PublicIPPrefixPropertiesFormat `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Sku: The public IP prefix SKU.
	Sku *PublicIPPrefixSku `json:"sku,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type PublicIPPrefixesSpecType `json:"type"`

	//Zones: A list of availability zones denoting the IP allocated for the resource
	//needs to come from.
	Zones []string `json:"zones,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/PublicIPPrefixPropertiesFormat
type PublicIPPrefixPropertiesFormat struct {

	//IpTags: The list of tags associated with the public IP prefix.
	IpTags []IpTag `json:"ipTags,omitempty"`

	//PrefixLength: The Length of the Public IP Prefix.
	PrefixLength *int `json:"prefixLength,omitempty"`

	//PublicIPAddressVersion: The public IP address version.
	PublicIPAddressVersion *PublicIPPrefixPropertiesFormatPublicIPAddressVersion `json:"publicIPAddressVersion,omitempty"`
}

//Generated from:
type PublicIPPrefixPropertiesFormat_Status struct {

	//IpPrefix: The allocated Prefix.
	IpPrefix *string `json:"ipPrefix,omitempty"`

	//IpTags: The list of tags associated with the public IP prefix.
	IpTags []IpTag_Status `json:"ipTags,omitempty"`

	//LoadBalancerFrontendIpConfiguration: The reference to load balancer frontend IP
	//configuration associated with the public IP prefix.
	LoadBalancerFrontendIpConfiguration *SubResource_Status `json:"loadBalancerFrontendIpConfiguration,omitempty"`

	//PrefixLength: The Length of the Public IP Prefix.
	PrefixLength *int `json:"prefixLength,omitempty"`

	//ProvisioningState: The provisioning state of the public IP prefix resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//PublicIPAddressVersion: The public IP address version.
	PublicIPAddressVersion *IPVersion_Status `json:"publicIPAddressVersion,omitempty"`

	//PublicIPAddresses: The list of all referenced PublicIPAddresses.
	PublicIPAddresses []ReferencedPublicIpAddress_Status `json:"publicIPAddresses,omitempty"`

	//ResourceGuid: The resource GUID property of the public IP prefix resource.
	ResourceGuid *string `json:"resourceGuid,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/PublicIPPrefixSku
type PublicIPPrefixSku struct {

	//Name: Name of a public IP prefix SKU.
	Name *PublicIPPrefixSkuName `json:"name,omitempty"`
}

//Generated from:
type PublicIPPrefixSku_Status struct {

	//Name: Name of a public IP prefix SKU.
	Name *PublicIPPrefixSkuStatusName `json:"name,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type PublicIPPrefixesSpecApiVersion string

const PublicIPPrefixesSpecApiVersion20200501 = PublicIPPrefixesSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/publicIPPrefixes"}
type PublicIPPrefixesSpecType string

const PublicIPPrefixesSpecTypeMicrosoftNetworkPublicIPPrefixes = PublicIPPrefixesSpecType("Microsoft.Network/publicIPPrefixes")

// +kubebuilder:validation:Enum={"IPv4","IPv6"}
type PublicIPPrefixPropertiesFormatPublicIPAddressVersion string

const (
	PublicIPPrefixPropertiesFormatPublicIPAddressVersionIPv4 = PublicIPPrefixPropertiesFormatPublicIPAddressVersion("IPv4")
	PublicIPPrefixPropertiesFormatPublicIPAddressVersionIPv6 = PublicIPPrefixPropertiesFormatPublicIPAddressVersion("IPv6")
)

// +kubebuilder:validation:Enum={"Standard"}
type PublicIPPrefixSkuName string

const PublicIPPrefixSkuNameStandard = PublicIPPrefixSkuName("Standard")

// +kubebuilder:validation:Enum={"Standard"}
type PublicIPPrefixSkuStatusName string

const PublicIPPrefixSkuStatusNameStandard = PublicIPPrefixSkuStatusName("Standard")

//Generated from:
type ReferencedPublicIpAddress_Status struct {

	//Id: The PublicIPAddress Reference.
	Id *string `json:"id,omitempty"`
}

func init() {
	SchemeBuilder.Register(&PublicIPPrefixes{}, &PublicIPPrefixesList{})
}
