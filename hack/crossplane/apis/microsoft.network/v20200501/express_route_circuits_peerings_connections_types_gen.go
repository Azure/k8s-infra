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
type ExpressRouteCircuitsPeeringsConnections struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ExpressRouteCircuitsPeeringsConnections_Spec `json:"spec,omitempty"`
	Status            ExpressRouteCircuitConnection_Status         `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ExpressRouteCircuitsPeeringsConnectionsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExpressRouteCircuitsPeeringsConnections `json:"items"`
}

type ExpressRouteCircuitConnection_Status struct {
	AtProvider ExpressRouteCircuitsPeeringsConnectionsObservation `json:"atProvider"`
}

type ExpressRouteCircuitsPeeringsConnections_Spec struct {
	ForProvider ExpressRouteCircuitsPeeringsConnectionsParameters `json:"forProvider"`
}

type ExpressRouteCircuitsPeeringsConnectionsObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the express route circuit connection.
	Properties *ExpressRouteCircuitConnectionPropertiesFormat_Status `json:"properties,omitempty"`

	//Type: Type of the resource.
	Type *string `json:"type,omitempty"`
}

type ExpressRouteCircuitsPeeringsConnectionsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion ExpressRouteCircuitsPeeringsConnectionsSpecApiVersion `json:"apiVersion"`
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
	//Properties: Properties of the express route circuit connection.
	Properties ExpressRouteCircuitConnectionPropertiesFormat `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type ExpressRouteCircuitsPeeringsConnectionsSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/ExpressRouteCircuitConnectionPropertiesFormat
type ExpressRouteCircuitConnectionPropertiesFormat struct {

	//AddressPrefix: /29 IP address space to carve out Customer addresses for tunnels.
	AddressPrefix *string `json:"addressPrefix,omitempty"`

	//AuthorizationKey: The authorization key.
	AuthorizationKey *string `json:"authorizationKey,omitempty"`

	//ExpressRouteCircuitPeering: Reference to Express Route Circuit Private Peering
	//Resource of the circuit initiating connection.
	ExpressRouteCircuitPeering *SubResource `json:"expressRouteCircuitPeering,omitempty"`

	//Ipv6CircuitConnectionConfig: IPv6 Address PrefixProperties of the express route
	//circuit connection.
	Ipv6CircuitConnectionConfig *Ipv6CircuitConnectionConfig `json:"ipv6CircuitConnectionConfig,omitempty"`

	//PeerExpressRouteCircuitPeering: Reference to Express Route Circuit Private
	//Peering Resource of the peered circuit.
	PeerExpressRouteCircuitPeering *SubResource `json:"peerExpressRouteCircuitPeering,omitempty"`
}

//Generated from:
type ExpressRouteCircuitConnectionPropertiesFormat_Status struct {

	//AddressPrefix: /29 IP address space to carve out Customer addresses for tunnels.
	AddressPrefix *string `json:"addressPrefix,omitempty"`

	//AuthorizationKey: The authorization key.
	AuthorizationKey *string `json:"authorizationKey,omitempty"`

	//CircuitConnectionStatus: Express Route Circuit connection state.
	CircuitConnectionStatus *CircuitConnectionStatus_Status `json:"circuitConnectionStatus,omitempty"`

	//ExpressRouteCircuitPeering: Reference to Express Route Circuit Private Peering
	//Resource of the circuit initiating connection.
	ExpressRouteCircuitPeering *SubResource_Status `json:"expressRouteCircuitPeering,omitempty"`

	//Ipv6CircuitConnectionConfig: IPv6 Address PrefixProperties of the express route
	//circuit connection.
	Ipv6CircuitConnectionConfig *Ipv6CircuitConnectionConfig_Status `json:"ipv6CircuitConnectionConfig,omitempty"`

	//PeerExpressRouteCircuitPeering: Reference to Express Route Circuit Private
	//Peering Resource of the peered circuit.
	PeerExpressRouteCircuitPeering *SubResource_Status `json:"peerExpressRouteCircuitPeering,omitempty"`

	//ProvisioningState: The provisioning state of the express route circuit
	//connection resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type ExpressRouteCircuitsPeeringsConnectionsSpecApiVersion string

const ExpressRouteCircuitsPeeringsConnectionsSpecApiVersion20200501 = ExpressRouteCircuitsPeeringsConnectionsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/expressRouteCircuits/peerings/connections"}
type ExpressRouteCircuitsPeeringsConnectionsSpecType string

const ExpressRouteCircuitsPeeringsConnectionsSpecTypeMicrosoftNetworkExpressRouteCircuitsPeeringsConnections = ExpressRouteCircuitsPeeringsConnectionsSpecType("Microsoft.Network/expressRouteCircuits/peerings/connections")

//Generated from:
// +kubebuilder:validation:Enum={"Connected","Connecting","Disconnected"}
type CircuitConnectionStatus_Status string

const (
	CircuitConnectionStatus_StatusConnected    = CircuitConnectionStatus_Status("Connected")
	CircuitConnectionStatus_StatusConnecting   = CircuitConnectionStatus_Status("Connecting")
	CircuitConnectionStatus_StatusDisconnected = CircuitConnectionStatus_Status("Disconnected")
)

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/Ipv6CircuitConnectionConfig
type Ipv6CircuitConnectionConfig struct {

	//AddressPrefix: /125 IP address space to carve out customer addresses for global
	//reach.
	AddressPrefix *string `json:"addressPrefix,omitempty"`
}

//Generated from:
type Ipv6CircuitConnectionConfig_Status struct {

	//AddressPrefix: /125 IP address space to carve out customer addresses for global
	//reach.
	AddressPrefix *string `json:"addressPrefix,omitempty"`

	//CircuitConnectionStatus: Express Route Circuit connection state.
	CircuitConnectionStatus *CircuitConnectionStatus_Status `json:"circuitConnectionStatus,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ExpressRouteCircuitsPeeringsConnections{}, &ExpressRouteCircuitsPeeringsConnectionsList{})
}
