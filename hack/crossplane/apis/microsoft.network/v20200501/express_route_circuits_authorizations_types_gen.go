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
type ExpressRouteCircuitsAuthorizations struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ExpressRouteCircuitsAuthorizations_Spec `json:"spec,omitempty"`
	Status            ExpressRouteCircuitAuthorization_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ExpressRouteCircuitsAuthorizationsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExpressRouteCircuitsAuthorizations `json:"items"`
}

type ExpressRouteCircuitAuthorization_Status struct {
	AtProvider ExpressRouteCircuitsAuthorizationsObservation `json:"atProvider"`
}

type ExpressRouteCircuitsAuthorizations_Spec struct {
	ForProvider ExpressRouteCircuitsAuthorizationsParameters `json:"forProvider"`
}

type ExpressRouteCircuitsAuthorizationsObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the express route circuit authorization.
	Properties *AuthorizationPropertiesFormat_Status `json:"properties,omitempty"`

	//Type: Type of the resource.
	Type *string `json:"type,omitempty"`
}

type ExpressRouteCircuitsAuthorizationsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion ExpressRouteCircuitsAuthorizationsSpecApiVersion `json:"apiVersion"`
	Comments   *string                                          `json:"comments,omitempty"`

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
	//Properties: Properties of the express route circuit authorization.
	Properties interface{} `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type ExpressRouteCircuitsAuthorizationsSpecType `json:"type"`
}

//Generated from:
type AuthorizationPropertiesFormat_Status struct {

	//AuthorizationKey: The authorization key.
	AuthorizationKey *string `json:"authorizationKey,omitempty"`

	//AuthorizationUseStatus: The authorization use status.
	AuthorizationUseStatus *AuthorizationPropertiesFormatStatusAuthorizationUseStatus `json:"authorizationUseStatus,omitempty"`

	//ProvisioningState: The provisioning state of the authorization resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type ExpressRouteCircuitsAuthorizationsSpecApiVersion string

const ExpressRouteCircuitsAuthorizationsSpecApiVersion20200501 = ExpressRouteCircuitsAuthorizationsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/expressRouteCircuits/authorizations"}
type ExpressRouteCircuitsAuthorizationsSpecType string

const ExpressRouteCircuitsAuthorizationsSpecTypeMicrosoftNetworkExpressRouteCircuitsAuthorizations = ExpressRouteCircuitsAuthorizationsSpecType("Microsoft.Network/expressRouteCircuits/authorizations")

// +kubebuilder:validation:Enum={"Available","InUse"}
type AuthorizationPropertiesFormatStatusAuthorizationUseStatus string

const (
	AuthorizationPropertiesFormatStatusAuthorizationUseStatusAvailable = AuthorizationPropertiesFormatStatusAuthorizationUseStatus("Available")
	AuthorizationPropertiesFormatStatusAuthorizationUseStatusInUse     = AuthorizationPropertiesFormatStatusAuthorizationUseStatus("InUse")
)

func init() {
	SchemeBuilder.Register(&ExpressRouteCircuitsAuthorizations{}, &ExpressRouteCircuitsAuthorizationsList{})
}
