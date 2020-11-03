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
type SecurityPartnerProviders struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SecurityPartnerProviders_Spec  `json:"spec,omitempty"`
	Status            SecurityPartnerProvider_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type SecurityPartnerProvidersList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityPartnerProviders `json:"items"`
}

type SecurityPartnerProvider_Status struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              SecurityPartnerProvidersObservation `json:"atProvider"`
}

type SecurityPartnerProviders_Spec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           SecurityPartnerProvidersParameters `json:"forProvider"`
}

type SecurityPartnerProvidersObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Location: Resource location.
	Location *string `json:"location,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the Security Partner Provider.
	Properties *SecurityPartnerProviderPropertiesFormat_Status `json:"properties,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type SecurityPartnerProvidersParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion SecurityPartnerProvidersSpecApiVersion `json:"apiVersion"`
	Comments   *string                                `json:"comments,omitempty"`

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
	//Properties: Properties of the Security Partner Provider.
	Properties                SecurityPartnerProviderPropertiesFormat `json:"properties"`
	ResourceGroupName         string                                  `json:"resourceGroupName"`
	ResourceGroupNameRef      *v1alpha1.Reference                     `json:"resourceGroupNameRef,omitempty"`
	ResourceGroupNameSelector *v1alpha1.Selector                      `json:"resourceGroupNameSelector,omitempty"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type SecurityPartnerProvidersSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/SecurityPartnerProviderPropertiesFormat
type SecurityPartnerProviderPropertiesFormat struct {

	//SecurityProviderName: The security provider name.
	SecurityProviderName *SecurityPartnerProviderPropertiesFormatSecurityProviderName `json:"securityProviderName,omitempty"`

	//VirtualHub: The virtualHub to which the Security Partner Provider belongs.
	VirtualHub *SubResource `json:"virtualHub,omitempty"`
}

//Generated from:
type SecurityPartnerProviderPropertiesFormat_Status struct {

	//ConnectionStatus: The connection status with the Security Partner Provider.
	ConnectionStatus *SecurityPartnerProviderConnectionStatus_Status `json:"connectionStatus,omitempty"`

	//ProvisioningState: The provisioning state of the Security Partner Provider
	//resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//SecurityProviderName: The security provider name.
	SecurityProviderName *SecurityPartnerProvidersecurityProviderName_Status `json:"securityProviderName,omitempty"`

	//VirtualHub: The virtualHub to which the Security Partner Provider belongs.
	VirtualHub *SubResource_Status `json:"virtualHub,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type SecurityPartnerProvidersSpecApiVersion string

const SecurityPartnerProvidersSpecApiVersion20200501 = SecurityPartnerProvidersSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/securityPartnerProviders"}
type SecurityPartnerProvidersSpecType string

const SecurityPartnerProvidersSpecTypeMicrosoftNetworkSecurityPartnerProviders = SecurityPartnerProvidersSpecType("Microsoft.Network/securityPartnerProviders")

//Generated from:
// +kubebuilder:validation:Enum={"Connected","NotConnected","PartiallyConnected","Unknown"}
type SecurityPartnerProviderConnectionStatus_Status string

const (
	SecurityPartnerProviderConnectionStatus_StatusConnected          = SecurityPartnerProviderConnectionStatus_Status("Connected")
	SecurityPartnerProviderConnectionStatus_StatusNotConnected       = SecurityPartnerProviderConnectionStatus_Status("NotConnected")
	SecurityPartnerProviderConnectionStatus_StatusPartiallyConnected = SecurityPartnerProviderConnectionStatus_Status("PartiallyConnected")
	SecurityPartnerProviderConnectionStatus_StatusUnknown            = SecurityPartnerProviderConnectionStatus_Status("Unknown")
)

// +kubebuilder:validation:Enum={"Checkpoint","IBoss","ZScaler"}
type SecurityPartnerProviderPropertiesFormatSecurityProviderName string

const (
	SecurityPartnerProviderPropertiesFormatSecurityProviderNameCheckpoint = SecurityPartnerProviderPropertiesFormatSecurityProviderName("Checkpoint")
	SecurityPartnerProviderPropertiesFormatSecurityProviderNameIBoss      = SecurityPartnerProviderPropertiesFormatSecurityProviderName("IBoss")
	SecurityPartnerProviderPropertiesFormatSecurityProviderNameZScaler    = SecurityPartnerProviderPropertiesFormatSecurityProviderName("ZScaler")
)

//Generated from:
// +kubebuilder:validation:Enum={"Checkpoint","IBoss","ZScaler"}
type SecurityPartnerProvidersecurityProviderName_Status string

const (
	SecurityPartnerProvidersecurityProviderName_StatusCheckpoint = SecurityPartnerProvidersecurityProviderName_Status("Checkpoint")
	SecurityPartnerProvidersecurityProviderName_StatusIBoss      = SecurityPartnerProvidersecurityProviderName_Status("IBoss")
	SecurityPartnerProvidersecurityProviderName_StatusZScaler    = SecurityPartnerProvidersecurityProviderName_Status("ZScaler")
)

func init() {
	SchemeBuilder.Register(&SecurityPartnerProviders{}, &SecurityPartnerProvidersList{})
}
