// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
// Code generated by k8s-infra-gen. DO NOT EDIT.
package v20200501

import (
	"github.com/Azure/k8s-infra/hack/crossplane/apis/deploymenttemplate/v20150101"
	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
type NetworkVirtualAppliances struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              NetworkVirtualAppliances_Spec  `json:"spec,omitempty"`
	Status            NetworkVirtualAppliance_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type NetworkVirtualAppliancesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkVirtualAppliances `json:"items"`
}

type NetworkVirtualAppliance_Status struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              NetworkVirtualAppliancesObservation `json:"atProvider"`
}

type NetworkVirtualAppliances_Spec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           NetworkVirtualAppliancesParameters `json:"forProvider"`
}

type NetworkVirtualAppliancesObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Identity: The service principal that has read access to cloud-init and config
	//blob.
	Identity *ManagedServiceIdentity_Status `json:"identity,omitempty"`

	//Location: Resource location.
	Location *string `json:"location,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the Network Virtual Appliance.
	Properties *NetworkVirtualAppliancePropertiesFormat_Status `json:"properties,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type NetworkVirtualAppliancesParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion NetworkVirtualAppliancesSpecApiVersion `json:"apiVersion"`
	Comments   *string                                `json:"comments,omitempty"`

	//Condition: Condition of the resource
	Condition *bool                   `json:"condition,omitempty"`
	Copy      *v20150101.ResourceCopy `json:"copy,omitempty"`

	//DependsOn: Collection of resources this resource depends on
	DependsOn []string `json:"dependsOn,omitempty"`

	//Identity: The service principal that has read access to cloud-init and config
	//blob.
	Identity *ManagedServiceIdentity `json:"identity,omitempty"`

	//Location: Location to deploy resource to
	Location string `json:"location,omitempty"`

	// +kubebuilder:validation:Required
	//Name: Name of the resource
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	//Properties: Properties of the Network Virtual Appliance.
	Properties NetworkVirtualAppliancePropertiesFormat `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type NetworkVirtualAppliancesSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/NetworkVirtualAppliancePropertiesFormat
type NetworkVirtualAppliancePropertiesFormat struct {

	//BootStrapConfigurationBlobs: BootStrapConfigurationBlobs storage URLs.
	BootStrapConfigurationBlobs []string `json:"bootStrapConfigurationBlobs,omitempty"`

	//CloudInitConfiguration: CloudInitConfiguration string in plain text.
	CloudInitConfiguration *string `json:"cloudInitConfiguration,omitempty"`

	//CloudInitConfigurationBlobs: CloudInitConfigurationBlob storage URLs.
	CloudInitConfigurationBlobs []string `json:"cloudInitConfigurationBlobs,omitempty"`

	//NvaSku: Network Virtual Appliance SKU.
	NvaSku *VirtualApplianceSkuProperties `json:"nvaSku,omitempty"`

	//VirtualApplianceAsn: VirtualAppliance ASN.
	VirtualApplianceAsn *int `json:"virtualApplianceAsn,omitempty"`

	//VirtualHub: The Virtual Hub where Network Virtual Appliance is being deployed.
	VirtualHub *SubResource `json:"virtualHub,omitempty"`
}

//Generated from:
type NetworkVirtualAppliancePropertiesFormat_Status struct {

	//BootStrapConfigurationBlobs: BootStrapConfigurationBlobs storage URLs.
	BootStrapConfigurationBlobs []string `json:"bootStrapConfigurationBlobs,omitempty"`

	//CloudInitConfiguration: CloudInitConfiguration string in plain text.
	CloudInitConfiguration *string `json:"cloudInitConfiguration,omitempty"`

	//CloudInitConfigurationBlobs: CloudInitConfigurationBlob storage URLs.
	CloudInitConfigurationBlobs []string `json:"cloudInitConfigurationBlobs,omitempty"`

	//NvaSku: Network Virtual Appliance SKU.
	NvaSku *VirtualApplianceSkuProperties_Status `json:"nvaSku,omitempty"`

	//ProvisioningState: The provisioning state of the resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//VirtualApplianceAsn: VirtualAppliance ASN.
	VirtualApplianceAsn *int `json:"virtualApplianceAsn,omitempty"`

	//VirtualApplianceNics: List of Virtual Appliance Network Interfaces.
	VirtualApplianceNics []VirtualApplianceNicProperties_Status `json:"virtualApplianceNics,omitempty"`

	//VirtualApplianceSites: List of references to VirtualApplianceSite.
	VirtualApplianceSites []SubResource_Status `json:"virtualApplianceSites,omitempty"`

	//VirtualHub: The Virtual Hub where Network Virtual Appliance is being deployed.
	VirtualHub *SubResource_Status `json:"virtualHub,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type NetworkVirtualAppliancesSpecApiVersion string

const NetworkVirtualAppliancesSpecApiVersion20200501 = NetworkVirtualAppliancesSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/networkVirtualAppliances"}
type NetworkVirtualAppliancesSpecType string

const NetworkVirtualAppliancesSpecTypeMicrosoftNetworkNetworkVirtualAppliances = NetworkVirtualAppliancesSpecType("Microsoft.Network/networkVirtualAppliances")

//Generated from:
type VirtualApplianceNicProperties_Status struct {

	//Name: NIC name.
	Name *string `json:"name,omitempty"`

	//PrivateIpAddress: Private IP address.
	PrivateIpAddress *string `json:"privateIpAddress,omitempty"`

	//PublicIpAddress: Public IP address.
	PublicIpAddress *string `json:"publicIpAddress,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/VirtualApplianceSkuProperties
type VirtualApplianceSkuProperties struct {

	//BundledScaleUnit: Virtual Appliance Scale Unit.
	BundledScaleUnit *string `json:"bundledScaleUnit,omitempty"`

	//MarketPlaceVersion: Virtual Appliance Version.
	MarketPlaceVersion *string `json:"marketPlaceVersion,omitempty"`

	//Vendor: Virtual Appliance Vendor.
	Vendor *string `json:"vendor,omitempty"`
}

//Generated from:
type VirtualApplianceSkuProperties_Status struct {

	//BundledScaleUnit: Virtual Appliance Scale Unit.
	BundledScaleUnit *string `json:"bundledScaleUnit,omitempty"`

	//MarketPlaceVersion: Virtual Appliance Version.
	MarketPlaceVersion *string `json:"marketPlaceVersion,omitempty"`

	//Vendor: Virtual Appliance Vendor.
	Vendor *string `json:"vendor,omitempty"`
}

func init() {
	SchemeBuilder.Register(&NetworkVirtualAppliances{}, &NetworkVirtualAppliancesList{})
}
