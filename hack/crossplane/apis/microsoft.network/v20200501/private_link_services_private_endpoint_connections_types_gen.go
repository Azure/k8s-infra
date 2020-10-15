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
type PrivateLinkServicesPrivateEndpointConnections struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PrivateLinkServicesPrivateEndpointConnections_Spec `json:"spec,omitempty"`
	Status            PrivateEndpointConnection_Status                   `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type PrivateLinkServicesPrivateEndpointConnectionsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PrivateLinkServicesPrivateEndpointConnections `json:"items"`
}

//Generated from:
type PrivateEndpointConnection_Status struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the private end point connection.
	Properties *PrivateEndpointConnectionProperties_Status `json:"properties,omitempty"`

	//Type: The resource type.
	Type *string `json:"type,omitempty"`
}

type PrivateLinkServicesPrivateEndpointConnections_Spec struct {
	ForProvider PrivateLinkServicesPrivateEndpointConnectionsParameters `json:"forProvider"`
}

//Generated from:
type PrivateEndpointConnectionProperties_Status struct {

	//LinkIdentifier: The consumer link id.
	LinkIdentifier *string `json:"linkIdentifier,omitempty"`

	//PrivateEndpoint: The resource of private end point.
	PrivateEndpoint *PrivateEndpoint_Status `json:"privateEndpoint,omitempty"`

	//PrivateLinkServiceConnectionState: A collection of information about the state
	//of the connection between service consumer and provider.
	PrivateLinkServiceConnectionState *PrivateLinkServiceConnectionState_Status `json:"privateLinkServiceConnectionState,omitempty"`

	//ProvisioningState: The provisioning state of the private endpoint connection
	//resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`
}

type PrivateLinkServicesPrivateEndpointConnectionsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion PrivateLinkServicesPrivateEndpointConnectionsSpecApiVersion `json:"apiVersion"`
	Comments   *string                                                     `json:"comments,omitempty"`

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
	//Properties: Properties of the private end point connection.
	Properties PrivateEndpointConnectionProperties `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type PrivateLinkServicesPrivateEndpointConnectionsSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/PrivateEndpointConnectionProperties
type PrivateEndpointConnectionProperties struct {

	//PrivateLinkServiceConnectionState: A collection of information about the state
	//of the connection between service consumer and provider.
	PrivateLinkServiceConnectionState *PrivateLinkServiceConnectionState `json:"privateLinkServiceConnectionState,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type PrivateLinkServicesPrivateEndpointConnectionsSpecApiVersion string

const PrivateLinkServicesPrivateEndpointConnectionsSpecApiVersion20200501 = PrivateLinkServicesPrivateEndpointConnectionsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/privateLinkServices/privateEndpointConnections"}
type PrivateLinkServicesPrivateEndpointConnectionsSpecType string

const PrivateLinkServicesPrivateEndpointConnectionsSpecTypeMicrosoftNetworkPrivateLinkServicesPrivateEndpointConnections = PrivateLinkServicesPrivateEndpointConnectionsSpecType("Microsoft.Network/privateLinkServices/privateEndpointConnections")

func init() {
	SchemeBuilder.Register(&PrivateLinkServicesPrivateEndpointConnections{}, &PrivateLinkServicesPrivateEndpointConnectionsList{})
}
