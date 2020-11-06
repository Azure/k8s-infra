// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
// Code generated by k8s-infra-gen. DO NOT EDIT.
package v20150501preview

import (
	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type ServersVirtualNetworkRules struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServersVirtualNetworkRules_Spec `json:"spec,omitempty"`
	Status            VirtualNetworkRule_Status       `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ServersVirtualNetworkRulesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServersVirtualNetworkRules `json:"items"`
}

type ServersVirtualNetworkRules_Spec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           ServersVirtualNetworkRulesParameters `json:"forProvider"`
}

type VirtualNetworkRule_Status struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              ServersVirtualNetworkRulesObservation `json:"atProvider"`
}

type ServersVirtualNetworkRulesObservation struct {

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Resource properties.
	Properties *VirtualNetworkRuleProperties_Status `json:"properties,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type ServersVirtualNetworkRulesParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion ServersVirtualNetworkRulesSpecApiVersion `json:"apiVersion"`

	//Location: Location to deploy resource to
	Location *string `json:"location,omitempty"`

	// +kubebuilder:validation:Required
	//Name: Name of the resource
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	//Properties: Properties of a virtual network rule.
	Properties                VirtualNetworkRuleProperties `json:"properties"`
	ResourceGroupName         string                       `json:"resourceGroupName"`
	ResourceGroupNameRef      *v1alpha1.Reference          `json:"resourceGroupNameRef,omitempty"`
	ResourceGroupNameSelector *v1alpha1.Selector           `json:"resourceGroupNameSelector,omitempty"`
	ServersName               string                       `json:"serversName"`
	ServersNameRef            *v1alpha1.Reference          `json:"serversNameRef,omitempty"`
	ServersNameSelector       *v1alpha1.Selector           `json:"serversNameSelector,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type ServersVirtualNetworkRulesSpecType `json:"type"`
}

// +kubebuilder:validation:Enum={"2015-05-01-preview"}
type ServersVirtualNetworkRulesSpecApiVersion string

const ServersVirtualNetworkRulesSpecApiVersion20150501Preview = ServersVirtualNetworkRulesSpecApiVersion("2015-05-01-preview")

// +kubebuilder:validation:Enum={"Microsoft.Sql/servers/virtualNetworkRules"}
type ServersVirtualNetworkRulesSpecType string

const ServersVirtualNetworkRulesSpecTypeMicrosoftSqlServersVirtualNetworkRules = ServersVirtualNetworkRulesSpecType("Microsoft.Sql/servers/virtualNetworkRules")

//Generated from: https://schema.management.azure.com/schemas/2015-05-01-preview/Microsoft.Sql.json#/definitions/VirtualNetworkRuleProperties
type VirtualNetworkRuleProperties struct {

	//IgnoreMissingVnetServiceEndpoint: Create firewall rule before the virtual
	//network has vnet service endpoint enabled.
	IgnoreMissingVnetServiceEndpoint *bool `json:"ignoreMissingVnetServiceEndpoint,omitempty"`

	// +kubebuilder:validation:Required
	//VirtualNetworkSubnetId: The ARM resource id of the virtual network subnet.
	VirtualNetworkSubnetId string `json:"virtualNetworkSubnetId"`
}

//Generated from:
type VirtualNetworkRuleProperties_Status struct {

	//IgnoreMissingVnetServiceEndpoint: Create firewall rule before the virtual
	//network has vnet service endpoint enabled.
	IgnoreMissingVnetServiceEndpoint *bool `json:"ignoreMissingVnetServiceEndpoint,omitempty"`

	//State: Virtual Network Rule State
	State *VirtualNetworkRulePropertiesStatusState `json:"state,omitempty"`

	// +kubebuilder:validation:Required
	//VirtualNetworkSubnetId: The ARM resource id of the virtual network subnet.
	VirtualNetworkSubnetId string `json:"virtualNetworkSubnetId"`
}

// +kubebuilder:validation:Enum={"Deleting","InProgress","Initializing","Ready","Unknown"}
type VirtualNetworkRulePropertiesStatusState string

const (
	VirtualNetworkRulePropertiesStatusStateDeleting     = VirtualNetworkRulePropertiesStatusState("Deleting")
	VirtualNetworkRulePropertiesStatusStateInProgress   = VirtualNetworkRulePropertiesStatusState("InProgress")
	VirtualNetworkRulePropertiesStatusStateInitializing = VirtualNetworkRulePropertiesStatusState("Initializing")
	VirtualNetworkRulePropertiesStatusStateReady        = VirtualNetworkRulePropertiesStatusState("Ready")
	VirtualNetworkRulePropertiesStatusStateUnknown      = VirtualNetworkRulePropertiesStatusState("Unknown")
)

func init() {
	SchemeBuilder.Register(&ServersVirtualNetworkRules{}, &ServersVirtualNetworkRulesList{})
}