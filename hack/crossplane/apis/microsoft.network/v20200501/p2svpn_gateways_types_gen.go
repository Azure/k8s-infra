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
type P2svpnGateways struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              P2svpnGateways_Spec  `json:"spec,omitempty"`
	Status            P2SVpnGateway_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type P2svpnGatewaysList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []P2svpnGateways `json:"items"`
}

type P2SVpnGateway_Status struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              P2svpnGatewaysObservation `json:"atProvider"`
}

type P2svpnGateways_Spec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           P2svpnGatewaysParameters `json:"forProvider"`
}

type P2svpnGatewaysObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Location: Resource location.
	Location *string `json:"location,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the P2SVpnGateway.
	Properties *P2SVpnGatewayProperties_Status `json:"properties,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type P2svpnGatewaysParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion P2svpnGatewaysSpecApiVersion `json:"apiVersion"`
	Comments   *string                      `json:"comments,omitempty"`

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
	//Properties: Properties of the P2SVpnGateway.
	Properties P2SVpnGatewayProperties `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type P2svpnGatewaysSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/P2SVpnGatewayProperties
type P2SVpnGatewayProperties struct {

	//CustomDnsServers: List of all customer specified DNS servers IP addresses.
	CustomDnsServers []string `json:"customDnsServers,omitempty"`

	//P2SConnectionConfigurations: List of all p2s connection configurations of the
	//gateway.
	P2SConnectionConfigurations []P2SConnectionConfiguration `json:"p2SConnectionConfigurations,omitempty"`

	//VirtualHub: The VirtualHub to which the gateway belongs.
	VirtualHub *SubResource `json:"virtualHub,omitempty"`

	//VpnGatewayScaleUnit: The scale unit for this p2s vpn gateway.
	VpnGatewayScaleUnit *int `json:"vpnGatewayScaleUnit,omitempty"`

	//VpnServerConfiguration: The VpnServerConfiguration to which the p2sVpnGateway is
	//attached to.
	VpnServerConfiguration *SubResource `json:"vpnServerConfiguration,omitempty"`
}

//Generated from:
type P2SVpnGatewayProperties_Status struct {

	//CustomDnsServers: List of all customer specified DNS servers IP addresses.
	CustomDnsServers []string `json:"customDnsServers,omitempty"`

	//P2SConnectionConfigurations: List of all p2s connection configurations of the
	//gateway.
	P2SConnectionConfigurations []P2SConnectionConfiguration_Status `json:"p2SConnectionConfigurations,omitempty"`

	//ProvisioningState: The provisioning state of the P2S VPN gateway resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//VirtualHub: The VirtualHub to which the gateway belongs.
	VirtualHub *SubResource_Status `json:"virtualHub,omitempty"`

	//VpnClientConnectionHealth: All P2S VPN clients' connection health status.
	VpnClientConnectionHealth *VpnClientConnectionHealth_Status `json:"vpnClientConnectionHealth,omitempty"`

	//VpnGatewayScaleUnit: The scale unit for this p2s vpn gateway.
	VpnGatewayScaleUnit *int `json:"vpnGatewayScaleUnit,omitempty"`

	//VpnServerConfiguration: The VpnServerConfiguration to which the p2sVpnGateway is
	//attached to.
	VpnServerConfiguration *SubResource_Status `json:"vpnServerConfiguration,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type P2svpnGatewaysSpecApiVersion string

const P2svpnGatewaysSpecApiVersion20200501 = P2svpnGatewaysSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/p2svpnGateways"}
type P2svpnGatewaysSpecType string

const P2svpnGatewaysSpecTypeMicrosoftNetworkP2svpnGateways = P2svpnGatewaysSpecType("Microsoft.Network/p2svpnGateways")

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/P2SConnectionConfiguration
type P2SConnectionConfiguration struct {

	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the P2S connection configuration.
	Properties *P2SConnectionConfigurationProperties `json:"properties,omitempty"`
}

//Generated from:
type P2SConnectionConfiguration_Status struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the P2S connection configuration.
	Properties *P2SConnectionConfigurationProperties_Status `json:"properties,omitempty"`
}

//Generated from:
type VpnClientConnectionHealth_Status struct {

	//AllocatedIpAddresses: List of allocated ip addresses to the connected p2s vpn
	//clients.
	AllocatedIpAddresses []string `json:"allocatedIpAddresses,omitempty"`

	//TotalEgressBytesTransferred: Total of the Egress Bytes Transferred in this
	//connection.
	TotalEgressBytesTransferred *int `json:"totalEgressBytesTransferred,omitempty"`

	//TotalIngressBytesTransferred: Total of the Ingress Bytes Transferred in this P2S
	//Vpn connection.
	TotalIngressBytesTransferred *int `json:"totalIngressBytesTransferred,omitempty"`

	//VpnClientConnectionsCount: The total of p2s vpn clients connected at this time
	//to this P2SVpnGateway.
	VpnClientConnectionsCount *int `json:"vpnClientConnectionsCount,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/P2SConnectionConfigurationProperties
type P2SConnectionConfigurationProperties struct {

	//RoutingConfiguration: The Routing Configuration indicating the associated and
	//propagated route tables on this connection.
	RoutingConfiguration *RoutingConfiguration `json:"routingConfiguration,omitempty"`

	//VpnClientAddressPool: The reference to the address space resource which
	//represents Address space for P2S VpnClient.
	VpnClientAddressPool *AddressSpace `json:"vpnClientAddressPool,omitempty"`
}

//Generated from:
type P2SConnectionConfigurationProperties_Status struct {

	//ProvisioningState: The provisioning state of the P2SConnectionConfiguration
	//resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//RoutingConfiguration: The Routing Configuration indicating the associated and
	//propagated route tables on this connection.
	RoutingConfiguration *RoutingConfiguration_Status `json:"routingConfiguration,omitempty"`

	//VpnClientAddressPool: The reference to the address space resource which
	//represents Address space for P2S VpnClient.
	VpnClientAddressPool *AddressSpace_Status `json:"vpnClientAddressPool,omitempty"`
}

func init() {
	SchemeBuilder.Register(&P2svpnGateways{}, &P2svpnGatewaysList{})
}
