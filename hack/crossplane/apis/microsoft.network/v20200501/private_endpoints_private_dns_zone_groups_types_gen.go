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
type PrivateEndpointsPrivateDnsZoneGroups struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PrivateEndpointsPrivateDnsZoneGroups_Spec `json:"spec,omitempty"`
	Status            PrivateDnsZoneGroup_Status                `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type PrivateEndpointsPrivateDnsZoneGroupsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PrivateEndpointsPrivateDnsZoneGroups `json:"items"`
}

type PrivateDnsZoneGroup_Status struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              PrivateEndpointsPrivateDnsZoneGroupsObservation `json:"atProvider"`
}

type PrivateEndpointsPrivateDnsZoneGroups_Spec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           PrivateEndpointsPrivateDnsZoneGroupsParameters `json:"forProvider"`
}

type PrivateEndpointsPrivateDnsZoneGroupsObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: Name of the resource that is unique within a resource group. This name can
	//be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the private dns zone group.
	Properties *PrivateDnsZoneGroupPropertiesFormat_Status `json:"properties,omitempty"`
}

type PrivateEndpointsPrivateDnsZoneGroupsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion PrivateEndpointsPrivateDnsZoneGroupsSpecApiVersion `json:"apiVersion"`
	Comments   *string                                            `json:"comments,omitempty"`

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
	//Properties: Properties of the private dns zone group.
	Properties PrivateDnsZoneGroupPropertiesFormat `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type PrivateEndpointsPrivateDnsZoneGroupsSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/PrivateDnsZoneGroupPropertiesFormat
type PrivateDnsZoneGroupPropertiesFormat struct {

	//PrivateDnsZoneConfigs: A collection of private dns zone configurations of the
	//private dns zone group.
	PrivateDnsZoneConfigs []PrivateDnsZoneConfig `json:"privateDnsZoneConfigs,omitempty"`
}

//Generated from:
type PrivateDnsZoneGroupPropertiesFormat_Status struct {

	//PrivateDnsZoneConfigs: A collection of private dns zone configurations of the
	//private dns zone group.
	PrivateDnsZoneConfigs []PrivateDnsZoneConfig_Status `json:"privateDnsZoneConfigs,omitempty"`

	//ProvisioningState: The provisioning state of the private dns zone group resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type PrivateEndpointsPrivateDnsZoneGroupsSpecApiVersion string

const PrivateEndpointsPrivateDnsZoneGroupsSpecApiVersion20200501 = PrivateEndpointsPrivateDnsZoneGroupsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/privateEndpoints/privateDnsZoneGroups"}
type PrivateEndpointsPrivateDnsZoneGroupsSpecType string

const PrivateEndpointsPrivateDnsZoneGroupsSpecTypeMicrosoftNetworkPrivateEndpointsPrivateDnsZoneGroups = PrivateEndpointsPrivateDnsZoneGroupsSpecType("Microsoft.Network/privateEndpoints/privateDnsZoneGroups")

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/PrivateDnsZoneConfig
type PrivateDnsZoneConfig struct {

	//Name: Name of the resource that is unique within a resource group. This name can
	//be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the private dns zone configuration.
	Properties *PrivateDnsZonePropertiesFormat `json:"properties,omitempty"`
}

//Generated from:
type PrivateDnsZoneConfig_Status struct {

	//Name: Name of the resource that is unique within a resource group. This name can
	//be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the private dns zone configuration.
	Properties *PrivateDnsZonePropertiesFormat_Status `json:"properties,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/PrivateDnsZonePropertiesFormat
type PrivateDnsZonePropertiesFormat struct {

	//PrivateDnsZoneId: The resource id of the private dns zone.
	PrivateDnsZoneId *string `json:"privateDnsZoneId,omitempty"`
}

//Generated from:
type PrivateDnsZonePropertiesFormat_Status struct {

	//PrivateDnsZoneId: The resource id of the private dns zone.
	PrivateDnsZoneId *string `json:"privateDnsZoneId,omitempty"`

	//RecordSets: A collection of information regarding a recordSet, holding
	//information to identify private resources.
	RecordSets []RecordSet_Status `json:"recordSets,omitempty"`
}

//Generated from:
type RecordSet_Status struct {

	//Fqdn: Fqdn that resolves to private endpoint ip address.
	Fqdn *string `json:"fqdn,omitempty"`

	//IpAddresses: The private ip address of the private endpoint.
	IpAddresses []string `json:"ipAddresses,omitempty"`

	//ProvisioningState: The provisioning state of the recordset.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//RecordSetName: Recordset name.
	RecordSetName *string `json:"recordSetName,omitempty"`

	//RecordType: Resource record type.
	RecordType *string `json:"recordType,omitempty"`

	//Ttl: Recordset time to live.
	Ttl *int `json:"ttl,omitempty"`
}

func init() {
	SchemeBuilder.Register(&PrivateEndpointsPrivateDnsZoneGroups{}, &PrivateEndpointsPrivateDnsZoneGroupsList{})
}
