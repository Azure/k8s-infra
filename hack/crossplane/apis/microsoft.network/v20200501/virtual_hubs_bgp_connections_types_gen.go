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
type VirtualHubsBgpConnections struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VirtualHubsBgpConnections_Spec `json:"spec,omitempty"`
	Status            BgpConnection_Status           `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type VirtualHubsBgpConnectionsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualHubsBgpConnections `json:"items"`
}

type BgpConnection_Status struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              VirtualHubsBgpConnectionsObservation `json:"atProvider"`
}

type VirtualHubsBgpConnections_Spec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           VirtualHubsBgpConnectionsParameters `json:"forProvider"`
}

type VirtualHubsBgpConnectionsObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: Name of the connection.
	Name *string `json:"name,omitempty"`

	//Properties: The properties of the Bgp connections.
	Properties *BgpConnectionProperties_Status `json:"properties,omitempty"`

	//Type: Connection type.
	Type *string `json:"type,omitempty"`
}

type VirtualHubsBgpConnectionsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion VirtualHubsBgpConnectionsSpecApiVersion `json:"apiVersion"`
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
	//Properties: The properties of the Bgp connections.
	Properties BgpConnectionProperties `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type VirtualHubsBgpConnectionsSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/BgpConnectionProperties
type BgpConnectionProperties struct {

	//PeerAsn: Peer ASN.
	PeerAsn *int `json:"peerAsn,omitempty"`

	//PeerIp: Peer IP.
	PeerIp *string `json:"peerIp,omitempty"`
}

//Generated from:
type BgpConnectionProperties_Status struct {

	//ConnectionState: The current state of the VirtualHub to Peer.
	ConnectionState *BgpConnectionPropertiesStatusConnectionState `json:"connectionState,omitempty"`

	//PeerAsn: Peer ASN.
	PeerAsn *int `json:"peerAsn,omitempty"`

	//PeerIp: Peer IP.
	PeerIp *string `json:"peerIp,omitempty"`

	//ProvisioningState: The provisioning state of the resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type VirtualHubsBgpConnectionsSpecApiVersion string

const VirtualHubsBgpConnectionsSpecApiVersion20200501 = VirtualHubsBgpConnectionsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/virtualHubs/bgpConnections"}
type VirtualHubsBgpConnectionsSpecType string

const VirtualHubsBgpConnectionsSpecTypeMicrosoftNetworkVirtualHubsBgpConnections = VirtualHubsBgpConnectionsSpecType("Microsoft.Network/virtualHubs/bgpConnections")

// +kubebuilder:validation:Enum={"Connected","Connecting","NotConnected","Unknown"}
type BgpConnectionPropertiesStatusConnectionState string

const (
	BgpConnectionPropertiesStatusConnectionStateConnected    = BgpConnectionPropertiesStatusConnectionState("Connected")
	BgpConnectionPropertiesStatusConnectionStateConnecting   = BgpConnectionPropertiesStatusConnectionState("Connecting")
	BgpConnectionPropertiesStatusConnectionStateNotConnected = BgpConnectionPropertiesStatusConnectionState("NotConnected")
	BgpConnectionPropertiesStatusConnectionStateUnknown      = BgpConnectionPropertiesStatusConnectionState("Unknown")
)

func init() {
	SchemeBuilder.Register(&VirtualHubsBgpConnections{}, &VirtualHubsBgpConnectionsList{})
}
