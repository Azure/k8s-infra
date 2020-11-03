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
type VpnGatewaysVpnConnections struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VpnGatewaysVpnConnections_Spec `json:"spec,omitempty"`
	Status            VpnConnection_Status           `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type VpnGatewaysVpnConnectionsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VpnGatewaysVpnConnections `json:"items"`
}

type VpnConnection_Status struct {
	AtProvider VpnGatewaysVpnConnectionsObservation `json:"atProvider"`
}

type VpnGatewaysVpnConnections_Spec struct {
	ForProvider VpnGatewaysVpnConnectionsParameters `json:"forProvider"`
}

type VpnGatewaysVpnConnectionsObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the VPN connection.
	Properties *VpnConnectionProperties_Status `json:"properties,omitempty"`
}

type VpnGatewaysVpnConnectionsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion VpnGatewaysVpnConnectionsSpecApiVersion `json:"apiVersion"`
	Comments   *string                                 `json:"comments,omitempty"`

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
	//Properties: Properties of the VPN connection.
	Properties VpnConnectionProperties `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type VpnGatewaysVpnConnectionsSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/VpnConnectionProperties
type VpnConnectionProperties struct {

	//ConnectionBandwidth: Expected bandwidth in MBPS.
	ConnectionBandwidth *int `json:"connectionBandwidth,omitempty"`

	//ConnectionStatus: The connection status.
	ConnectionStatus *VpnConnectionPropertiesConnectionStatus `json:"connectionStatus,omitempty"`

	//DpdTimeoutSeconds: The dead peer detection timeout for a vpn connection in
	//seconds.
	DpdTimeoutSeconds *int `json:"dpdTimeoutSeconds,omitempty"`

	//EnableBgp: EnableBgp flag.
	EnableBgp *bool `json:"enableBgp,omitempty"`

	//EnableInternetSecurity: Enable internet security.
	EnableInternetSecurity *bool `json:"enableInternetSecurity,omitempty"`

	//EnableRateLimiting: EnableBgp flag.
	EnableRateLimiting *bool `json:"enableRateLimiting,omitempty"`

	//IpsecPolicies: The IPSec Policies to be considered by this connection.
	IpsecPolicies []IpsecPolicy `json:"ipsecPolicies,omitempty"`

	//RemoteVpnSite: Id of the connected vpn site.
	RemoteVpnSite *SubResource `json:"remoteVpnSite,omitempty"`

	//RoutingConfiguration: The Routing Configuration indicating the associated and
	//propagated route tables on this connection.
	RoutingConfiguration *RoutingConfiguration `json:"routingConfiguration,omitempty"`

	//RoutingWeight: Routing weight for vpn connection.
	RoutingWeight *int `json:"routingWeight,omitempty"`

	//SharedKey: SharedKey for the vpn connection.
	SharedKey *string `json:"sharedKey,omitempty"`

	//UseLocalAzureIpAddress: Use local azure ip to initiate connection.
	UseLocalAzureIpAddress *bool `json:"useLocalAzureIpAddress,omitempty"`

	//UsePolicyBasedTrafficSelectors: Enable policy-based traffic selectors.
	UsePolicyBasedTrafficSelectors *bool `json:"usePolicyBasedTrafficSelectors,omitempty"`

	//VpnConnectionProtocolType: Connection protocol used for this connection.
	VpnConnectionProtocolType *VpnConnectionPropertiesVpnConnectionProtocolType `json:"vpnConnectionProtocolType,omitempty"`

	//VpnLinkConnections: List of all vpn site link connections to the gateway.
	VpnLinkConnections []VpnSiteLinkConnection `json:"vpnLinkConnections,omitempty"`
}

//Generated from:
type VpnConnectionProperties_Status struct {

	//ConnectionBandwidth: Expected bandwidth in MBPS.
	ConnectionBandwidth *int `json:"connectionBandwidth,omitempty"`

	//ConnectionStatus: The connection status.
	ConnectionStatus *VpnConnectionStatus_Status `json:"connectionStatus,omitempty"`

	//DpdTimeoutSeconds: The dead peer detection timeout for a vpn connection in
	//seconds.
	DpdTimeoutSeconds *int `json:"dpdTimeoutSeconds,omitempty"`

	//EgressBytesTransferred: Egress bytes transferred.
	EgressBytesTransferred *int `json:"egressBytesTransferred,omitempty"`

	//EnableBgp: EnableBgp flag.
	EnableBgp *bool `json:"enableBgp,omitempty"`

	//EnableInternetSecurity: Enable internet security.
	EnableInternetSecurity *bool `json:"enableInternetSecurity,omitempty"`

	//EnableRateLimiting: EnableBgp flag.
	EnableRateLimiting *bool `json:"enableRateLimiting,omitempty"`

	//IngressBytesTransferred: Ingress bytes transferred.
	IngressBytesTransferred *int `json:"ingressBytesTransferred,omitempty"`

	//IpsecPolicies: The IPSec Policies to be considered by this connection.
	IpsecPolicies []IpsecPolicy_Status `json:"ipsecPolicies,omitempty"`

	//ProvisioningState: The provisioning state of the VPN connection resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//RemoteVpnSite: Id of the connected vpn site.
	RemoteVpnSite *SubResource_Status `json:"remoteVpnSite,omitempty"`

	//RoutingConfiguration: The Routing Configuration indicating the associated and
	//propagated route tables on this connection.
	RoutingConfiguration *RoutingConfiguration_Status `json:"routingConfiguration,omitempty"`

	//RoutingWeight: Routing weight for vpn connection.
	RoutingWeight *int `json:"routingWeight,omitempty"`

	//SharedKey: SharedKey for the vpn connection.
	SharedKey *string `json:"sharedKey,omitempty"`

	//UseLocalAzureIpAddress: Use local azure ip to initiate connection.
	UseLocalAzureIpAddress *bool `json:"useLocalAzureIpAddress,omitempty"`

	//UsePolicyBasedTrafficSelectors: Enable policy-based traffic selectors.
	UsePolicyBasedTrafficSelectors *bool `json:"usePolicyBasedTrafficSelectors,omitempty"`

	//VpnConnectionProtocolType: Connection protocol used for this connection.
	VpnConnectionProtocolType *ConnectionProtocol_Status `json:"vpnConnectionProtocolType,omitempty"`

	//VpnLinkConnections: List of all vpn site link connections to the gateway.
	VpnLinkConnections []VpnSiteLinkConnection_Status `json:"vpnLinkConnections,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type VpnGatewaysVpnConnectionsSpecApiVersion string

const VpnGatewaysVpnConnectionsSpecApiVersion20200501 = VpnGatewaysVpnConnectionsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/vpnGateways/vpnConnections"}
type VpnGatewaysVpnConnectionsSpecType string

const VpnGatewaysVpnConnectionsSpecTypeMicrosoftNetworkVpnGatewaysVpnConnections = VpnGatewaysVpnConnectionsSpecType("Microsoft.Network/vpnGateways/vpnConnections")

// +kubebuilder:validation:Enum={"Connected","Connecting","NotConnected","Unknown"}
type VpnConnectionPropertiesConnectionStatus string

const (
	VpnConnectionPropertiesConnectionStatusConnected    = VpnConnectionPropertiesConnectionStatus("Connected")
	VpnConnectionPropertiesConnectionStatusConnecting   = VpnConnectionPropertiesConnectionStatus("Connecting")
	VpnConnectionPropertiesConnectionStatusNotConnected = VpnConnectionPropertiesConnectionStatus("NotConnected")
	VpnConnectionPropertiesConnectionStatusUnknown      = VpnConnectionPropertiesConnectionStatus("Unknown")
)

// +kubebuilder:validation:Enum={"IKEv1","IKEv2"}
type VpnConnectionPropertiesVpnConnectionProtocolType string

const (
	VpnConnectionPropertiesVpnConnectionProtocolTypeIKEv1 = VpnConnectionPropertiesVpnConnectionProtocolType("IKEv1")
	VpnConnectionPropertiesVpnConnectionProtocolTypeIKEv2 = VpnConnectionPropertiesVpnConnectionProtocolType("IKEv2")
)

//Generated from:
// +kubebuilder:validation:Enum={"Connected","Connecting","NotConnected","Unknown"}
type VpnConnectionStatus_Status string

const (
	VpnConnectionStatus_StatusConnected    = VpnConnectionStatus_Status("Connected")
	VpnConnectionStatus_StatusConnecting   = VpnConnectionStatus_Status("Connecting")
	VpnConnectionStatus_StatusNotConnected = VpnConnectionStatus_Status("NotConnected")
	VpnConnectionStatus_StatusUnknown      = VpnConnectionStatus_Status("Unknown")
)

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/VpnSiteLinkConnection
type VpnSiteLinkConnection struct {

	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the VPN site link connection.
	Properties *VpnSiteLinkConnectionProperties `json:"properties,omitempty"`
}

//Generated from:
type VpnSiteLinkConnection_Status struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the VPN site link connection.
	Properties *VpnSiteLinkConnectionProperties_Status `json:"properties,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/VpnSiteLinkConnectionProperties
type VpnSiteLinkConnectionProperties struct {

	//ConnectionBandwidth: Expected bandwidth in MBPS.
	ConnectionBandwidth *int `json:"connectionBandwidth,omitempty"`

	//ConnectionStatus: The connection status.
	ConnectionStatus *VpnSiteLinkConnectionPropertiesConnectionStatus `json:"connectionStatus,omitempty"`

	//EnableBgp: EnableBgp flag.
	EnableBgp *bool `json:"enableBgp,omitempty"`

	//EnableRateLimiting: EnableBgp flag.
	EnableRateLimiting *bool `json:"enableRateLimiting,omitempty"`

	//IpsecPolicies: The IPSec Policies to be considered by this connection.
	IpsecPolicies []IpsecPolicy `json:"ipsecPolicies,omitempty"`

	//RoutingWeight: Routing weight for vpn connection.
	RoutingWeight *int `json:"routingWeight,omitempty"`

	//SharedKey: SharedKey for the vpn connection.
	SharedKey *string `json:"sharedKey,omitempty"`

	//UseLocalAzureIpAddress: Use local azure ip to initiate connection.
	UseLocalAzureIpAddress *bool `json:"useLocalAzureIpAddress,omitempty"`

	//UsePolicyBasedTrafficSelectors: Enable policy-based traffic selectors.
	UsePolicyBasedTrafficSelectors *bool `json:"usePolicyBasedTrafficSelectors,omitempty"`

	//VpnConnectionProtocolType: Connection protocol used for this connection.
	VpnConnectionProtocolType *VpnSiteLinkConnectionPropertiesVpnConnectionProtocolType `json:"vpnConnectionProtocolType,omitempty"`

	//VpnSiteLink: Id of the connected vpn site link.
	VpnSiteLink *SubResource `json:"vpnSiteLink,omitempty"`
}

//Generated from:
type VpnSiteLinkConnectionProperties_Status struct {

	//ConnectionBandwidth: Expected bandwidth in MBPS.
	ConnectionBandwidth *int `json:"connectionBandwidth,omitempty"`

	//ConnectionStatus: The connection status.
	ConnectionStatus *VpnConnectionStatus_Status `json:"connectionStatus,omitempty"`

	//EgressBytesTransferred: Egress bytes transferred.
	EgressBytesTransferred *int `json:"egressBytesTransferred,omitempty"`

	//EnableBgp: EnableBgp flag.
	EnableBgp *bool `json:"enableBgp,omitempty"`

	//EnableRateLimiting: EnableBgp flag.
	EnableRateLimiting *bool `json:"enableRateLimiting,omitempty"`

	//IngressBytesTransferred: Ingress bytes transferred.
	IngressBytesTransferred *int `json:"ingressBytesTransferred,omitempty"`

	//IpsecPolicies: The IPSec Policies to be considered by this connection.
	IpsecPolicies []IpsecPolicy_Status `json:"ipsecPolicies,omitempty"`

	//ProvisioningState: The provisioning state of the VPN site link connection
	//resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//RoutingWeight: Routing weight for vpn connection.
	RoutingWeight *int `json:"routingWeight,omitempty"`

	//SharedKey: SharedKey for the vpn connection.
	SharedKey *string `json:"sharedKey,omitempty"`

	//UseLocalAzureIpAddress: Use local azure ip to initiate connection.
	UseLocalAzureIpAddress *bool `json:"useLocalAzureIpAddress,omitempty"`

	//UsePolicyBasedTrafficSelectors: Enable policy-based traffic selectors.
	UsePolicyBasedTrafficSelectors *bool `json:"usePolicyBasedTrafficSelectors,omitempty"`

	//VpnConnectionProtocolType: Connection protocol used for this connection.
	VpnConnectionProtocolType *ConnectionProtocol_Status `json:"vpnConnectionProtocolType,omitempty"`

	//VpnSiteLink: Id of the connected vpn site link.
	VpnSiteLink *SubResource_Status `json:"vpnSiteLink,omitempty"`
}

// +kubebuilder:validation:Enum={"Connected","Connecting","NotConnected","Unknown"}
type VpnSiteLinkConnectionPropertiesConnectionStatus string

const (
	VpnSiteLinkConnectionPropertiesConnectionStatusConnected    = VpnSiteLinkConnectionPropertiesConnectionStatus("Connected")
	VpnSiteLinkConnectionPropertiesConnectionStatusConnecting   = VpnSiteLinkConnectionPropertiesConnectionStatus("Connecting")
	VpnSiteLinkConnectionPropertiesConnectionStatusNotConnected = VpnSiteLinkConnectionPropertiesConnectionStatus("NotConnected")
	VpnSiteLinkConnectionPropertiesConnectionStatusUnknown      = VpnSiteLinkConnectionPropertiesConnectionStatus("Unknown")
)

// +kubebuilder:validation:Enum={"IKEv1","IKEv2"}
type VpnSiteLinkConnectionPropertiesVpnConnectionProtocolType string

const (
	VpnSiteLinkConnectionPropertiesVpnConnectionProtocolTypeIKEv1 = VpnSiteLinkConnectionPropertiesVpnConnectionProtocolType("IKEv1")
	VpnSiteLinkConnectionPropertiesVpnConnectionProtocolTypeIKEv2 = VpnSiteLinkConnectionPropertiesVpnConnectionProtocolType("IKEv2")
)

func init() {
	SchemeBuilder.Register(&VpnGatewaysVpnConnections{}, &VpnGatewaysVpnConnectionsList{})
}
