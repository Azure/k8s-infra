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
type ServiceEndpointPoliciesServiceEndpointPolicyDefinitions struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServiceEndpointPoliciesServiceEndpointPolicyDefinitions_Spec `json:"spec,omitempty"`
	Status            ServiceEndpointPolicyDefinition_Status                       `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ServiceEndpointPoliciesServiceEndpointPolicyDefinitionsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceEndpointPoliciesServiceEndpointPolicyDefinitions `json:"items"`
}

type ServiceEndpointPoliciesServiceEndpointPolicyDefinitions_Spec struct {
	ForProvider ServiceEndpointPoliciesServiceEndpointPolicyDefinitionsParameters `json:"forProvider"`
}

//Generated from:
type ServiceEndpointPolicyDefinition_Status struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: The name of the resource that is unique within a resource group. This name
	//can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the service endpoint policy definition.
	Properties *ServiceEndpointPolicyDefinitionPropertiesFormat_Status `json:"properties,omitempty"`
}

type ServiceEndpointPoliciesServiceEndpointPolicyDefinitionsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion ServiceEndpointPoliciesServiceEndpointPolicyDefinitionsSpecApiVersion `json:"apiVersion"`
	Comments   *string                                                               `json:"comments,omitempty"`

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
	//Properties: Properties of the service endpoint policy definition.
	Properties ServiceEndpointPolicyDefinitionPropertiesFormat `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type ServiceEndpointPoliciesServiceEndpointPolicyDefinitionsSpecType `json:"type"`
}

//Generated from:
type ServiceEndpointPolicyDefinitionPropertiesFormat_Status struct {

	//Description: A description for this rule. Restricted to 140 chars.
	Description *string `json:"description,omitempty"`

	//ProvisioningState: The provisioning state of the service endpoint policy
	//definition resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//Service: Service endpoint name.
	Service *string `json:"service,omitempty"`

	//ServiceResources: A list of service resources.
	ServiceResources []string `json:"serviceResources,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type ServiceEndpointPoliciesServiceEndpointPolicyDefinitionsSpecApiVersion string

const ServiceEndpointPoliciesServiceEndpointPolicyDefinitionsSpecApiVersion20200501 = ServiceEndpointPoliciesServiceEndpointPolicyDefinitionsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/serviceEndpointPolicies/serviceEndpointPolicyDefinitions"}
type ServiceEndpointPoliciesServiceEndpointPolicyDefinitionsSpecType string

const ServiceEndpointPoliciesServiceEndpointPolicyDefinitionsSpecTypeMicrosoftNetworkServiceEndpointPoliciesServiceEndpointPolicyDefinitions = ServiceEndpointPoliciesServiceEndpointPolicyDefinitionsSpecType("Microsoft.Network/serviceEndpointPolicies/serviceEndpointPolicyDefinitions")

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/ServiceEndpointPolicyDefinitionPropertiesFormat
type ServiceEndpointPolicyDefinitionPropertiesFormat struct {

	//Description: A description for this rule. Restricted to 140 chars.
	Description *string `json:"description,omitempty"`

	//Service: Service endpoint name.
	Service *string `json:"service,omitempty"`

	//ServiceResources: A list of service resources.
	ServiceResources []string `json:"serviceResources,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ServiceEndpointPoliciesServiceEndpointPolicyDefinitions{}, &ServiceEndpointPoliciesServiceEndpointPolicyDefinitionsList{})
}
