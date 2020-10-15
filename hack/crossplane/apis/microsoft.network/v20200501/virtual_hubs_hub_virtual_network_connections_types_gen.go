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
type VirtualHubsHubVirtualNetworkConnections struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VirtualHubsHubVirtualNetworkConnections_Spec `json:"spec,omitempty"`
	Status            HubVirtualNetworkConnection_Status           `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type VirtualHubsHubVirtualNetworkConnectionsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualHubsHubVirtualNetworkConnections `json:"items"`
}

type HubVirtualNetworkConnection_Status struct {
	AtProvider VirtualHubsHubVirtualNetworkConnectionsObservation `json:"atProvider"`
}

type VirtualHubsHubVirtualNetworkConnections_Spec struct {
	ForProvider VirtualHubsHubVirtualNetworkConnectionsParameters `json:"forProvider"`
}

type VirtualHubsHubVirtualNetworkConnectionsObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the hub virtual network connection.
	Properties *HubVirtualNetworkConnectionProperties_Status `json:"properties,omitempty"`
}

type VirtualHubsHubVirtualNetworkConnectionsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion VirtualHubsHubVirtualNetworkConnectionsSpecApiVersion `json:"apiVersion"`
	Comments   *string                                               `json:"comments,omitempty"`

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
	//Properties: Properties of the hub virtual network connection.
	Properties HubVirtualNetworkConnectionProperties `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type VirtualHubsHubVirtualNetworkConnectionsSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/HubVirtualNetworkConnectionProperties
type HubVirtualNetworkConnectionProperties struct {

	//AllowHubToRemoteVnetTransit: Deprecated: VirtualHub to RemoteVnet transit to
	//enabled or not.
	AllowHubToRemoteVnetTransit *bool `json:"allowHubToRemoteVnetTransit,omitempty"`

	//AllowRemoteVnetToUseHubVnetGateways: Deprecated: Allow RemoteVnet to use Virtual
	//Hub's gateways.
	AllowRemoteVnetToUseHubVnetGateways *bool `json:"allowRemoteVnetToUseHubVnetGateways,omitempty"`

	//EnableInternetSecurity: Enable internet security.
	EnableInternetSecurity *bool `json:"enableInternetSecurity,omitempty"`

	//RemoteVirtualNetwork: Reference to the remote virtual network.
	RemoteVirtualNetwork *SubResource `json:"remoteVirtualNetwork,omitempty"`

	//RoutingConfiguration: The Routing Configuration indicating the associated and
	//propagated route tables on this connection.
	RoutingConfiguration *RoutingConfiguration `json:"routingConfiguration,omitempty"`
}

//Generated from:
type HubVirtualNetworkConnectionProperties_Status struct {

	//AllowHubToRemoteVnetTransit: Deprecated: VirtualHub to RemoteVnet transit to
	//enabled or not.
	AllowHubToRemoteVnetTransit *bool `json:"allowHubToRemoteVnetTransit,omitempty"`

	//AllowRemoteVnetToUseHubVnetGateways: Deprecated: Allow RemoteVnet to use Virtual
	//Hub's gateways.
	AllowRemoteVnetToUseHubVnetGateways *bool `json:"allowRemoteVnetToUseHubVnetGateways,omitempty"`

	//EnableInternetSecurity: Enable internet security.
	EnableInternetSecurity *bool `json:"enableInternetSecurity,omitempty"`

	//ProvisioningState: The provisioning state of the hub virtual network connection
	//resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//RemoteVirtualNetwork: Reference to the remote virtual network.
	RemoteVirtualNetwork *SubResource_Status `json:"remoteVirtualNetwork,omitempty"`

	//RoutingConfiguration: The Routing Configuration indicating the associated and
	//propagated route tables on this connection.
	RoutingConfiguration *RoutingConfiguration_Status `json:"routingConfiguration,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type VirtualHubsHubVirtualNetworkConnectionsSpecApiVersion string

const VirtualHubsHubVirtualNetworkConnectionsSpecApiVersion20200501 = VirtualHubsHubVirtualNetworkConnectionsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/virtualHubs/hubVirtualNetworkConnections"}
type VirtualHubsHubVirtualNetworkConnectionsSpecType string

const VirtualHubsHubVirtualNetworkConnectionsSpecTypeMicrosoftNetworkVirtualHubsHubVirtualNetworkConnections = VirtualHubsHubVirtualNetworkConnectionsSpecType("Microsoft.Network/virtualHubs/hubVirtualNetworkConnections")

func init() {
	SchemeBuilder.Register(&VirtualHubsHubVirtualNetworkConnections{}, &VirtualHubsHubVirtualNetworkConnectionsList{})
}
