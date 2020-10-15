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
type RouteFilters struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RouteFilters_Spec  `json:"spec,omitempty"`
	Status            RouteFilter_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type RouteFiltersList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RouteFilters `json:"items"`
}

//Generated from:
type RouteFilter_Status struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Location: Resource location.
	Location *string `json:"location,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the route filter.
	Properties *RouteFilterPropertiesFormat_Status `json:"properties,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type RouteFilters_Spec struct {
	ForProvider RouteFiltersParameters `json:"forProvider"`
}

//Generated from:
type RouteFilterPropertiesFormat_Status struct {

	//Ipv6Peerings: A collection of references to express route circuit ipv6 peerings.
	Ipv6Peerings []ExpressRouteCircuitPeering_Status `json:"ipv6Peerings,omitempty"`

	//Peerings: A collection of references to express route circuit peerings.
	Peerings []ExpressRouteCircuitPeering_Status `json:"peerings,omitempty"`

	//ProvisioningState: The provisioning state of the route filter resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//Rules: Collection of RouteFilterRules contained within a route filter.
	Rules []RouteFilterRule_Status `json:"rules,omitempty"`
}

type RouteFiltersParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion RouteFiltersSpecApiVersion `json:"apiVersion"`
	Comments   *string                    `json:"comments,omitempty"`

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
	//Properties: Properties of the route filter.
	Properties RouteFilterPropertiesFormat `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type RouteFiltersSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/RouteFilterPropertiesFormat
type RouteFilterPropertiesFormat struct {

	//Rules: Collection of RouteFilterRules contained within a route filter.
	Rules []RouteFilterRule `json:"rules,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type RouteFiltersSpecApiVersion string

const RouteFiltersSpecApiVersion20200501 = RouteFiltersSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/routeFilters"}
type RouteFiltersSpecType string

const RouteFiltersSpecTypeMicrosoftNetworkRouteFilters = RouteFiltersSpecType("Microsoft.Network/routeFilters")

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/RouteFilterRule
type RouteFilterRule struct {

	//Location: Resource location.
	Location *string `json:"location,omitempty"`

	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the route filter rule.
	Properties *RouteFilterRulePropertiesFormat `json:"properties,omitempty"`
}

func init() {
	SchemeBuilder.Register(&RouteFilters{}, &RouteFiltersList{})
}
