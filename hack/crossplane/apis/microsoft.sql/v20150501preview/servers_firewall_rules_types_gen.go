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
type ServersFirewallRules struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServersFirewallRules_Spec `json:"spec,omitempty"`
	Status            FirewallRule_Status       `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ServersFirewallRulesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServersFirewallRules `json:"items"`
}

type FirewallRule_Status struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              ServersFirewallRulesObservation `json:"atProvider"`
}

type ServersFirewallRules_Spec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           ServersFirewallRulesParameters `json:"forProvider"`
}

type ServersFirewallRulesObservation struct {

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Resource properties.
	Properties *ServerFirewallRuleProperties_Status `json:"properties,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type ServersFirewallRulesParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion ServersFirewallRulesSpecApiVersion `json:"apiVersion"`

	//Location: Location to deploy resource to
	Location *string `json:"location,omitempty"`

	// +kubebuilder:validation:Required
	//Name: Name of the resource
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	//Properties: The properties of a server firewall rule.
	Properties                ServerFirewallRuleProperties `json:"properties"`
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
	Type ServersFirewallRulesSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2015-05-01-preview/Microsoft.Sql.json#/definitions/ServerFirewallRuleProperties
type ServerFirewallRuleProperties struct {

	//EndIpAddress: The end IP address of the firewall rule. Must be IPv4 format. Must
	//be greater than or equal to startIpAddress. Use value '0.0.0.0' for all
	//Azure-internal IP addresses.
	EndIpAddress *string `json:"endIpAddress,omitempty"`

	//StartIpAddress: The start IP address of the firewall rule. Must be IPv4 format.
	//Use value '0.0.0.0' for all Azure-internal IP addresses.
	StartIpAddress *string `json:"startIpAddress,omitempty"`
}

//Generated from:
type ServerFirewallRuleProperties_Status struct {

	//EndIpAddress: The end IP address of the firewall rule. Must be IPv4 format. Must
	//be greater than or equal to startIpAddress. Use value '0.0.0.0' for all
	//Azure-internal IP addresses.
	EndIpAddress *string `json:"endIpAddress,omitempty"`

	//StartIpAddress: The start IP address of the firewall rule. Must be IPv4 format.
	//Use value '0.0.0.0' for all Azure-internal IP addresses.
	StartIpAddress *string `json:"startIpAddress,omitempty"`
}

// +kubebuilder:validation:Enum={"2015-05-01-preview"}
type ServersFirewallRulesSpecApiVersion string

const ServersFirewallRulesSpecApiVersion20150501Preview = ServersFirewallRulesSpecApiVersion("2015-05-01-preview")

// +kubebuilder:validation:Enum={"Microsoft.Sql/servers/firewallRules"}
type ServersFirewallRulesSpecType string

const ServersFirewallRulesSpecTypeMicrosoftSqlServersFirewallRules = ServersFirewallRulesSpecType("Microsoft.Sql/servers/firewallRules")

func init() {
	SchemeBuilder.Register(&ServersFirewallRules{}, &ServersFirewallRulesList{})
}