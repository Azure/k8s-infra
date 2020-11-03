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
type VirtualRoutersPeerings struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VirtualRoutersPeerings_Spec `json:"spec,omitempty"`
	Status            VirtualRouterPeering_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type VirtualRoutersPeeringsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualRoutersPeerings `json:"items"`
}

type VirtualRouterPeering_Status struct {
	AtProvider VirtualRoutersPeeringsObservation `json:"atProvider"`
}

type VirtualRoutersPeerings_Spec struct {
	ForProvider VirtualRoutersPeeringsParameters `json:"forProvider"`
}

type VirtualRoutersPeeringsObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: Name of the virtual router peering that is unique within a virtual router.
	Name *string `json:"name,omitempty"`

	//Properties: The properties of the Virtual Router Peering.
	Properties *VirtualRouterPeeringProperties_Status `json:"properties,omitempty"`

	//Type: Peering type.
	Type *string `json:"type,omitempty"`
}

type VirtualRoutersPeeringsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion VirtualRoutersPeeringsSpecApiVersion `json:"apiVersion"`
	Comments   *string                              `json:"comments,omitempty"`

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
	//Properties: The properties of the Virtual Router Peering.
	Properties VirtualRouterPeeringProperties `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type VirtualRoutersPeeringsSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/VirtualRouterPeeringProperties
type VirtualRouterPeeringProperties struct {

	//PeerAsn: Peer ASN.
	PeerAsn *int `json:"peerAsn,omitempty"`

	//PeerIp: Peer IP.
	PeerIp *string `json:"peerIp,omitempty"`
}

//Generated from:
type VirtualRouterPeeringProperties_Status struct {

	//PeerAsn: Peer ASN.
	PeerAsn *int `json:"peerAsn,omitempty"`

	//PeerIp: Peer IP.
	PeerIp *string `json:"peerIp,omitempty"`

	//ProvisioningState: The provisioning state of the resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type VirtualRoutersPeeringsSpecApiVersion string

const VirtualRoutersPeeringsSpecApiVersion20200501 = VirtualRoutersPeeringsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/virtualRouters/peerings"}
type VirtualRoutersPeeringsSpecType string

const VirtualRoutersPeeringsSpecTypeMicrosoftNetworkVirtualRoutersPeerings = VirtualRoutersPeeringsSpecType("Microsoft.Network/virtualRouters/peerings")

func init() {
	SchemeBuilder.Register(&VirtualRoutersPeerings{}, &VirtualRoutersPeeringsList{})
}
