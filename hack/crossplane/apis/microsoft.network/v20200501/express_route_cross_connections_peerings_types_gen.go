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
type ExpressRouteCrossConnectionsPeerings struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ExpressRouteCrossConnectionsPeerings_Spec `json:"spec,omitempty"`
	Status            ExpressRouteCrossConnectionPeering_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ExpressRouteCrossConnectionsPeeringsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExpressRouteCrossConnectionsPeerings `json:"items"`
}

type ExpressRouteCrossConnectionPeering_Status struct {
	AtProvider ExpressRouteCrossConnectionsPeeringsObservation `json:"atProvider"`
}

type ExpressRouteCrossConnectionsPeerings_Spec struct {
	ForProvider ExpressRouteCrossConnectionsPeeringsParameters `json:"forProvider"`
}

type ExpressRouteCrossConnectionsPeeringsObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the express route cross connection peering.
	Properties *ExpressRouteCrossConnectionPeeringProperties_Status `json:"properties,omitempty"`
}

type ExpressRouteCrossConnectionsPeeringsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion ExpressRouteCrossConnectionsPeeringsSpecApiVersion `json:"apiVersion"`
	Comments   *string                                            `json:"comments,omitempty"`

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
	//Properties: Properties of the express route cross connection peering.
	Properties ExpressRouteCrossConnectionPeeringProperties `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type ExpressRouteCrossConnectionsPeeringsSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/ExpressRouteCrossConnectionPeeringProperties
type ExpressRouteCrossConnectionPeeringProperties struct {

	//GatewayManagerEtag: The GatewayManager Etag.
	GatewayManagerEtag *string `json:"gatewayManagerEtag,omitempty"`

	//Ipv6PeeringConfig: The IPv6 peering configuration.
	Ipv6PeeringConfig *Ipv6ExpressRouteCircuitPeeringConfig `json:"ipv6PeeringConfig,omitempty"`

	//MicrosoftPeeringConfig: The Microsoft peering configuration.
	MicrosoftPeeringConfig *ExpressRouteCircuitPeeringConfig `json:"microsoftPeeringConfig,omitempty"`

	//PeerASN: The peer ASN.
	PeerASN *int `json:"peerASN,omitempty"`

	//PeeringType: The peering type.
	PeeringType *ExpressRouteCrossConnectionPeeringPropertiesPeeringType `json:"peeringType,omitempty"`

	//PrimaryPeerAddressPrefix: The primary address prefix.
	PrimaryPeerAddressPrefix *string `json:"primaryPeerAddressPrefix,omitempty"`

	//SecondaryPeerAddressPrefix: The secondary address prefix.
	SecondaryPeerAddressPrefix *string `json:"secondaryPeerAddressPrefix,omitempty"`

	//SharedKey: The shared key.
	SharedKey *string `json:"sharedKey,omitempty"`

	//State: The peering state.
	State *ExpressRouteCrossConnectionPeeringPropertiesState `json:"state,omitempty"`

	//VlanId: The VLAN ID.
	VlanId *int `json:"vlanId,omitempty"`
}

//Generated from:
type ExpressRouteCrossConnectionPeeringProperties_Status struct {

	//AzureASN: The Azure ASN.
	AzureASN *int `json:"azureASN,omitempty"`

	//GatewayManagerEtag: The GatewayManager Etag.
	GatewayManagerEtag *string `json:"gatewayManagerEtag,omitempty"`

	//Ipv6PeeringConfig: The IPv6 peering configuration.
	Ipv6PeeringConfig *Ipv6ExpressRouteCircuitPeeringConfig_Status `json:"ipv6PeeringConfig,omitempty"`

	//LastModifiedBy: Who was the last to modify the peering.
	LastModifiedBy *string `json:"lastModifiedBy,omitempty"`

	//MicrosoftPeeringConfig: The Microsoft peering configuration.
	MicrosoftPeeringConfig *ExpressRouteCircuitPeeringConfig_Status `json:"microsoftPeeringConfig,omitempty"`

	//PeerASN: The peer ASN.
	PeerASN *int `json:"peerASN,omitempty"`

	//PeeringType: The peering type.
	PeeringType *ExpressRoutePeeringType_Status `json:"peeringType,omitempty"`

	//PrimaryAzurePort: The primary port.
	PrimaryAzurePort *string `json:"primaryAzurePort,omitempty"`

	//PrimaryPeerAddressPrefix: The primary address prefix.
	PrimaryPeerAddressPrefix *string `json:"primaryPeerAddressPrefix,omitempty"`

	//ProvisioningState: The provisioning state of the express route cross connection
	//peering resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//SecondaryAzurePort: The secondary port.
	SecondaryAzurePort *string `json:"secondaryAzurePort,omitempty"`

	//SecondaryPeerAddressPrefix: The secondary address prefix.
	SecondaryPeerAddressPrefix *string `json:"secondaryPeerAddressPrefix,omitempty"`

	//SharedKey: The shared key.
	SharedKey *string `json:"sharedKey,omitempty"`

	//State: The peering state.
	State *ExpressRoutePeeringState_Status `json:"state,omitempty"`

	//VlanId: The VLAN ID.
	VlanId *int `json:"vlanId,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type ExpressRouteCrossConnectionsPeeringsSpecApiVersion string

const ExpressRouteCrossConnectionsPeeringsSpecApiVersion20200501 = ExpressRouteCrossConnectionsPeeringsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/expressRouteCrossConnections/peerings"}
type ExpressRouteCrossConnectionsPeeringsSpecType string

const ExpressRouteCrossConnectionsPeeringsSpecTypeMicrosoftNetworkExpressRouteCrossConnectionsPeerings = ExpressRouteCrossConnectionsPeeringsSpecType("Microsoft.Network/expressRouteCrossConnections/peerings")

// +kubebuilder:validation:Enum={"AzurePrivatePeering","AzurePublicPeering","MicrosoftPeering"}
type ExpressRouteCrossConnectionPeeringPropertiesPeeringType string

const (
	ExpressRouteCrossConnectionPeeringPropertiesPeeringTypeAzurePrivatePeering = ExpressRouteCrossConnectionPeeringPropertiesPeeringType("AzurePrivatePeering")
	ExpressRouteCrossConnectionPeeringPropertiesPeeringTypeAzurePublicPeering  = ExpressRouteCrossConnectionPeeringPropertiesPeeringType("AzurePublicPeering")
	ExpressRouteCrossConnectionPeeringPropertiesPeeringTypeMicrosoftPeering    = ExpressRouteCrossConnectionPeeringPropertiesPeeringType("MicrosoftPeering")
)

// +kubebuilder:validation:Enum={"Disabled","Enabled"}
type ExpressRouteCrossConnectionPeeringPropertiesState string

const (
	ExpressRouteCrossConnectionPeeringPropertiesStateDisabled = ExpressRouteCrossConnectionPeeringPropertiesState("Disabled")
	ExpressRouteCrossConnectionPeeringPropertiesStateEnabled  = ExpressRouteCrossConnectionPeeringPropertiesState("Enabled")
)

func init() {
	SchemeBuilder.Register(&ExpressRouteCrossConnectionsPeerings{}, &ExpressRouteCrossConnectionsPeeringsList{})
}
