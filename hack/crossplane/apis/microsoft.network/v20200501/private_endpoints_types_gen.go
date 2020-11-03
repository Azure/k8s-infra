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
type PrivateEndpoints struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PrivateEndpoints_Spec  `json:"spec,omitempty"`
	Status            PrivateEndpoint_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type PrivateEndpointsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PrivateEndpoints `json:"items"`
}

type PrivateEndpoint_Status struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              PrivateEndpointsObservation `json:"atProvider"`
}

type PrivateEndpoints_Spec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           PrivateEndpointsParameters `json:"forProvider"`
}

type PrivateEndpointsObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Location: Resource location.
	Location *string `json:"location,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the private endpoint.
	Properties *PrivateEndpointProperties_Status `json:"properties,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type PrivateEndpointsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion PrivateEndpointsSpecApiVersion `json:"apiVersion"`
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
	//Properties: Properties of the private endpoint.
	Properties PrivateEndpointProperties `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type PrivateEndpointsSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/PrivateEndpointProperties
type PrivateEndpointProperties struct {

	//CustomDnsConfigs: An array of custom dns configurations.
	CustomDnsConfigs []CustomDnsConfigPropertiesFormat `json:"customDnsConfigs,omitempty"`

	//ManualPrivateLinkServiceConnections: A grouping of information about the
	//connection to the remote resource. Used when the network admin does not have
	//access to approve connections to the remote resource.
	ManualPrivateLinkServiceConnections []PrivateLinkServiceConnection `json:"manualPrivateLinkServiceConnections,omitempty"`

	//PrivateLinkServiceConnections: A grouping of information about the connection to
	//the remote resource.
	PrivateLinkServiceConnections []PrivateLinkServiceConnection `json:"privateLinkServiceConnections,omitempty"`

	//Subnet: The ID of the subnet from which the private IP will be allocated.
	Subnet *SubResource `json:"subnet,omitempty"`
}

//Generated from:
type PrivateEndpointProperties_Status struct {

	//CustomDnsConfigs: An array of custom dns configurations.
	CustomDnsConfigs []CustomDnsConfigPropertiesFormat_Status `json:"customDnsConfigs,omitempty"`

	//ManualPrivateLinkServiceConnections: A grouping of information about the
	//connection to the remote resource. Used when the network admin does not have
	//access to approve connections to the remote resource.
	ManualPrivateLinkServiceConnections []PrivateLinkServiceConnection_Status `json:"manualPrivateLinkServiceConnections,omitempty"`

	//NetworkInterfaces: An array of references to the network interfaces created for
	//this private endpoint.
	NetworkInterfaces []NetworkInterface_Status `json:"networkInterfaces,omitempty"`

	//PrivateLinkServiceConnections: A grouping of information about the connection to
	//the remote resource.
	PrivateLinkServiceConnections []PrivateLinkServiceConnection_Status `json:"privateLinkServiceConnections,omitempty"`

	//ProvisioningState: The provisioning state of the private endpoint resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//Subnet: The ID of the subnet from which the private IP will be allocated.
	Subnet *Subnet_Status `json:"subnet,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type PrivateEndpointsSpecApiVersion string

const PrivateEndpointsSpecApiVersion20200501 = PrivateEndpointsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/privateEndpoints"}
type PrivateEndpointsSpecType string

const PrivateEndpointsSpecTypeMicrosoftNetworkPrivateEndpoints = PrivateEndpointsSpecType("Microsoft.Network/privateEndpoints")

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/CustomDnsConfigPropertiesFormat
type CustomDnsConfigPropertiesFormat struct {

	//Fqdn: Fqdn that resolves to private endpoint ip address.
	Fqdn *string `json:"fqdn,omitempty"`

	//IpAddresses: A list of private ip addresses of the private endpoint.
	IpAddresses []string `json:"ipAddresses,omitempty"`
}

//Generated from:
type CustomDnsConfigPropertiesFormat_Status struct {

	//Fqdn: Fqdn that resolves to private endpoint ip address.
	Fqdn *string `json:"fqdn,omitempty"`

	//IpAddresses: A list of private ip addresses of the private endpoint.
	IpAddresses []string `json:"ipAddresses,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/PrivateLinkServiceConnection
type PrivateLinkServiceConnection struct {

	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the private link service connection.
	Properties *PrivateLinkServiceConnectionProperties `json:"properties,omitempty"`
}

//Generated from:
type PrivateLinkServiceConnection_Status struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the private link service connection.
	Properties *PrivateLinkServiceConnectionProperties_Status `json:"properties,omitempty"`

	//Type: The resource type.
	Type *string `json:"type,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/PrivateLinkServiceConnectionProperties
type PrivateLinkServiceConnectionProperties struct {

	//GroupIds: The ID(s) of the group(s) obtained from the remote resource that this
	//private endpoint should connect to.
	GroupIds []string `json:"groupIds,omitempty"`

	//PrivateLinkServiceConnectionState: A collection of read-only information about
	//the state of the connection to the remote resource.
	PrivateLinkServiceConnectionState *PrivateLinkServiceConnectionState `json:"privateLinkServiceConnectionState,omitempty"`

	//PrivateLinkServiceId: The resource id of private link service.
	PrivateLinkServiceId *string `json:"privateLinkServiceId,omitempty"`

	//RequestMessage: A message passed to the owner of the remote resource with this
	//connection request. Restricted to 140 chars.
	RequestMessage *string `json:"requestMessage,omitempty"`
}

//Generated from:
type PrivateLinkServiceConnectionProperties_Status struct {

	//GroupIds: The ID(s) of the group(s) obtained from the remote resource that this
	//private endpoint should connect to.
	GroupIds []string `json:"groupIds,omitempty"`

	//PrivateLinkServiceConnectionState: A collection of read-only information about
	//the state of the connection to the remote resource.
	PrivateLinkServiceConnectionState *PrivateLinkServiceConnectionState_Status `json:"privateLinkServiceConnectionState,omitempty"`

	//PrivateLinkServiceId: The resource id of private link service.
	PrivateLinkServiceId *string `json:"privateLinkServiceId,omitempty"`

	//ProvisioningState: The provisioning state of the private link service connection
	//resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//RequestMessage: A message passed to the owner of the remote resource with this
	//connection request. Restricted to 140 chars.
	RequestMessage *string `json:"requestMessage,omitempty"`
}

func init() {
	SchemeBuilder.Register(&PrivateEndpoints{}, &PrivateEndpointsList{})
}
