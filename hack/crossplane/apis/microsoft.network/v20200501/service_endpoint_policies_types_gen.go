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
type ServiceEndpointPolicies struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServiceEndpointPolicies_Spec `json:"spec,omitempty"`
	Status            ServiceEndpointPolicy_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ServiceEndpointPoliciesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceEndpointPolicies `json:"items"`
}

type ServiceEndpointPolicies_Spec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           ServiceEndpointPoliciesParameters `json:"forProvider"`
}

type ServiceEndpointPolicy_Status struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              ServiceEndpointPoliciesObservation `json:"atProvider"`
}

type ServiceEndpointPoliciesObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Location: Resource location.
	Location *string `json:"location,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the service end point policy.
	Properties *ServiceEndpointPolicyPropertiesFormat_Status `json:"properties,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type ServiceEndpointPoliciesParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion ServiceEndpointPoliciesSpecApiVersion `json:"apiVersion"`
	Comments   *string                               `json:"comments,omitempty"`

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
	//Properties: Properties of the service end point policy.
	Properties                ServiceEndpointPolicyPropertiesFormat `json:"properties"`
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
	Type ServiceEndpointPoliciesSpecType `json:"type"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type ServiceEndpointPoliciesSpecApiVersion string

const ServiceEndpointPoliciesSpecApiVersion20200501 = ServiceEndpointPoliciesSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/serviceEndpointPolicies"}
type ServiceEndpointPoliciesSpecType string

const ServiceEndpointPoliciesSpecTypeMicrosoftNetworkServiceEndpointPolicies = ServiceEndpointPoliciesSpecType("Microsoft.Network/serviceEndpointPolicies")

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/ServiceEndpointPolicyPropertiesFormat
type ServiceEndpointPolicyPropertiesFormat struct {

	//ServiceEndpointPolicyDefinitions: A collection of service endpoint policy
	//definitions of the service endpoint policy.
	ServiceEndpointPolicyDefinitions []ServiceEndpointPolicyDefinition `json:"serviceEndpointPolicyDefinitions,omitempty"`
}

//Generated from:
type ServiceEndpointPolicyPropertiesFormat_Status struct {

	//ProvisioningState: The provisioning state of the service endpoint policy
	//resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//ResourceGuid: The resource GUID property of the service endpoint policy resource.
	ResourceGuid *string `json:"resourceGuid,omitempty"`

	//ServiceEndpointPolicyDefinitions: A collection of service endpoint policy
	//definitions of the service endpoint policy.
	ServiceEndpointPolicyDefinitions []ServiceEndpointPolicyDefinition_Status `json:"serviceEndpointPolicyDefinitions,omitempty"`

	//Subnets: A collection of references to subnets.
	Subnets []Subnet_Status `json:"subnets,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/ServiceEndpointPolicyDefinition
type ServiceEndpointPolicyDefinition struct {

	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the service endpoint policy definition.
	Properties *ServiceEndpointPolicyDefinitionPropertiesFormat `json:"properties,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ServiceEndpointPolicies{}, &ServiceEndpointPoliciesList{})
}
