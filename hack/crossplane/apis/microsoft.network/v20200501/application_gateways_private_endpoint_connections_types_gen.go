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
type ApplicationGatewaysPrivateEndpointConnections struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ApplicationGatewaysPrivateEndpointConnections_Spec `json:"spec,omitempty"`
	Status            ApplicationGatewayPrivateEndpointConnection_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ApplicationGatewaysPrivateEndpointConnectionsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationGatewaysPrivateEndpointConnections `json:"items"`
}

type ApplicationGatewayPrivateEndpointConnection_Status struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              ApplicationGatewaysPrivateEndpointConnectionsObservation `json:"atProvider"`
}

type ApplicationGatewaysPrivateEndpointConnections_Spec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           ApplicationGatewaysPrivateEndpointConnectionsParameters `json:"forProvider"`
}

type ApplicationGatewaysPrivateEndpointConnectionsObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: Name of the private endpoint connection on an application gateway.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the application gateway private endpoint connection.
	Properties *ApplicationGatewayPrivateEndpointConnectionProperties_Status `json:"properties,omitempty"`

	//Type: Type of the resource.
	Type *string `json:"type,omitempty"`
}

type ApplicationGatewaysPrivateEndpointConnectionsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion                      ApplicationGatewaysPrivateEndpointConnectionsSpecApiVersion `json:"apiVersion"`
	ApplicationGatewaysName         string                                                      `json:"applicationGatewaysName"`
	ApplicationGatewaysNameRef      *v1alpha1.Reference                                         `json:"applicationGatewaysNameRef,omitempty"`
	ApplicationGatewaysNameSelector *v1alpha1.Selector                                          `json:"applicationGatewaysNameSelector,omitempty"`
	Comments                        *string                                                     `json:"comments,omitempty"`

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
	//Properties: Properties of the application gateway private endpoint connection.
	Properties                ApplicationGatewayPrivateEndpointConnectionProperties `json:"properties"`
	ResourceGroupName         string                                                `json:"resourceGroupName"`
	ResourceGroupNameRef      *v1alpha1.Reference                                   `json:"resourceGroupNameRef,omitempty"`
	ResourceGroupNameSelector *v1alpha1.Selector                                    `json:"resourceGroupNameSelector,omitempty"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type ApplicationGatewaysPrivateEndpointConnectionsSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/ApplicationGatewayPrivateEndpointConnectionProperties
type ApplicationGatewayPrivateEndpointConnectionProperties struct {

	//PrivateLinkServiceConnectionState: A collection of information about the state
	//of the connection between service consumer and provider.
	PrivateLinkServiceConnectionState *PrivateLinkServiceConnectionState `json:"privateLinkServiceConnectionState,omitempty"`
}

//Generated from:
type ApplicationGatewayPrivateEndpointConnectionProperties_Status struct {

	//LinkIdentifier: The consumer link id.
	LinkIdentifier *string `json:"linkIdentifier,omitempty"`

	//PrivateEndpoint: The resource of private end point.
	PrivateEndpoint *PrivateEndpoint_Status `json:"privateEndpoint,omitempty"`

	//PrivateLinkServiceConnectionState: A collection of information about the state
	//of the connection between service consumer and provider.
	PrivateLinkServiceConnectionState *PrivateLinkServiceConnectionState_Status `json:"privateLinkServiceConnectionState,omitempty"`

	//ProvisioningState: The provisioning state of the application gateway private
	//endpoint connection resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type ApplicationGatewaysPrivateEndpointConnectionsSpecApiVersion string

const ApplicationGatewaysPrivateEndpointConnectionsSpecApiVersion20200501 = ApplicationGatewaysPrivateEndpointConnectionsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/applicationGateways/privateEndpointConnections"}
type ApplicationGatewaysPrivateEndpointConnectionsSpecType string

const ApplicationGatewaysPrivateEndpointConnectionsSpecTypeMicrosoftNetworkApplicationGatewaysPrivateEndpointConnections = ApplicationGatewaysPrivateEndpointConnectionsSpecType("Microsoft.Network/applicationGateways/privateEndpointConnections")

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/PrivateLinkServiceConnectionState
type PrivateLinkServiceConnectionState struct {

	//ActionsRequired: A message indicating if changes on the service provider require
	//any updates on the consumer.
	ActionsRequired *string `json:"actionsRequired,omitempty"`

	//Description: The reason for approval/rejection of the connection.
	Description *string `json:"description,omitempty"`

	//Status: Indicates whether the connection has been Approved/Rejected/Removed by
	//the owner of the service.
	Status *string `json:"status,omitempty"`
}

//Generated from:
type PrivateLinkServiceConnectionState_Status struct {

	//ActionsRequired: A message indicating if changes on the service provider require
	//any updates on the consumer.
	ActionsRequired *string `json:"actionsRequired,omitempty"`

	//Description: The reason for approval/rejection of the connection.
	Description *string `json:"description,omitempty"`

	//Status: Indicates whether the connection has been Approved/Rejected/Removed by
	//the owner of the service.
	Status *string `json:"status,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ApplicationGatewaysPrivateEndpointConnections{}, &ApplicationGatewaysPrivateEndpointConnectionsList{})
}
