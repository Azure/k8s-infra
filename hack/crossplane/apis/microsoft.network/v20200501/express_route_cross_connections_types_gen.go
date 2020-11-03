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
type ExpressRouteCrossConnections struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ExpressRouteCrossConnections_Spec  `json:"spec,omitempty"`
	Status            ExpressRouteCrossConnection_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ExpressRouteCrossConnectionsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExpressRouteCrossConnections `json:"items"`
}

type ExpressRouteCrossConnection_Status struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              ExpressRouteCrossConnectionsObservation `json:"atProvider"`
}

type ExpressRouteCrossConnections_Spec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           ExpressRouteCrossConnectionsParameters `json:"forProvider"`
}

type ExpressRouteCrossConnectionsObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Location: Resource location.
	Location *string `json:"location,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the express route cross connection.
	Properties *ExpressRouteCrossConnectionProperties_Status `json:"properties,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type ExpressRouteCrossConnectionsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion ExpressRouteCrossConnectionsSpecApiVersion `json:"apiVersion"`
	Comments   *string                                    `json:"comments,omitempty"`

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
	//Properties: Properties of the express route cross connection.
	Properties                ExpressRouteCrossConnectionProperties `json:"properties"`
	ResourceGroupName         string                                `json:"resourceGroupName"`
	ResourceGroupNameRef      *v1alpha1.Reference                   `json:"resourceGroupNameRef,omitempty"`
	ResourceGroupNameSelector *v1alpha1.Selector                    `json:"resourceGroupNameSelector,omitempty"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type ExpressRouteCrossConnectionsSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/ExpressRouteCrossConnectionProperties
type ExpressRouteCrossConnectionProperties struct {

	//BandwidthInMbps: The circuit bandwidth In Mbps.
	BandwidthInMbps *int `json:"bandwidthInMbps,omitempty"`

	//ExpressRouteCircuit: The ExpressRouteCircuit.
	ExpressRouteCircuit *SubResource `json:"expressRouteCircuit,omitempty"`

	//PeeringLocation: The peering location of the ExpressRoute circuit.
	PeeringLocation *string `json:"peeringLocation,omitempty"`

	//Peerings: The list of peerings.
	Peerings []ExpressRouteCrossConnectionPeering `json:"peerings,omitempty"`

	//ServiceProviderNotes: Additional read only notes set by the connectivity
	//provider.
	ServiceProviderNotes *string `json:"serviceProviderNotes,omitempty"`

	//ServiceProviderProvisioningState: The provisioning state of the circuit in the
	//connectivity provider system.
	ServiceProviderProvisioningState *ExpressRouteCrossConnectionPropertiesServiceProviderProvisioningState `json:"serviceProviderProvisioningState,omitempty"`
}

//Generated from:
type ExpressRouteCrossConnectionProperties_Status struct {

	//BandwidthInMbps: The circuit bandwidth In Mbps.
	BandwidthInMbps *int `json:"bandwidthInMbps,omitempty"`

	//ExpressRouteCircuit: The ExpressRouteCircuit.
	ExpressRouteCircuit *ExpressRouteCircuitReference_Status `json:"expressRouteCircuit,omitempty"`

	//PeeringLocation: The peering location of the ExpressRoute circuit.
	PeeringLocation *string `json:"peeringLocation,omitempty"`

	//Peerings: The list of peerings.
	Peerings []ExpressRouteCrossConnectionPeering_Status `json:"peerings,omitempty"`

	//PrimaryAzurePort: The name of the primary port.
	PrimaryAzurePort *string `json:"primaryAzurePort,omitempty"`

	//ProvisioningState: The provisioning state of the express route cross connection
	//resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//STag: The identifier of the circuit traffic.
	STag *int `json:"sTag,omitempty"`

	//SecondaryAzurePort: The name of the secondary port.
	SecondaryAzurePort *string `json:"secondaryAzurePort,omitempty"`

	//ServiceProviderNotes: Additional read only notes set by the connectivity
	//provider.
	ServiceProviderNotes *string `json:"serviceProviderNotes,omitempty"`

	//ServiceProviderProvisioningState: The provisioning state of the circuit in the
	//connectivity provider system.
	ServiceProviderProvisioningState *ServiceProviderProvisioningState_Status `json:"serviceProviderProvisioningState,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type ExpressRouteCrossConnectionsSpecApiVersion string

const ExpressRouteCrossConnectionsSpecApiVersion20200501 = ExpressRouteCrossConnectionsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/expressRouteCrossConnections"}
type ExpressRouteCrossConnectionsSpecType string

const ExpressRouteCrossConnectionsSpecTypeMicrosoftNetworkExpressRouteCrossConnections = ExpressRouteCrossConnectionsSpecType("Microsoft.Network/expressRouteCrossConnections")

//Generated from:
type ExpressRouteCircuitReference_Status struct {

	//Id: Corresponding Express Route Circuit Id.
	Id *string `json:"id,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/ExpressRouteCrossConnectionPeering
type ExpressRouteCrossConnectionPeering struct {

	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the express route cross connection peering.
	Properties *ExpressRouteCrossConnectionPeeringProperties `json:"properties,omitempty"`
}

// +kubebuilder:validation:Enum={"Deprovisioning","NotProvisioned","Provisioned","Provisioning"}
type ExpressRouteCrossConnectionPropertiesServiceProviderProvisioningState string

const (
	ExpressRouteCrossConnectionPropertiesServiceProviderProvisioningStateDeprovisioning = ExpressRouteCrossConnectionPropertiesServiceProviderProvisioningState("Deprovisioning")
	ExpressRouteCrossConnectionPropertiesServiceProviderProvisioningStateNotProvisioned = ExpressRouteCrossConnectionPropertiesServiceProviderProvisioningState("NotProvisioned")
	ExpressRouteCrossConnectionPropertiesServiceProviderProvisioningStateProvisioned    = ExpressRouteCrossConnectionPropertiesServiceProviderProvisioningState("Provisioned")
	ExpressRouteCrossConnectionPropertiesServiceProviderProvisioningStateProvisioning   = ExpressRouteCrossConnectionPropertiesServiceProviderProvisioningState("Provisioning")
)

func init() {
	SchemeBuilder.Register(&ExpressRouteCrossConnections{}, &ExpressRouteCrossConnectionsList{})
}
