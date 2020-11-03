// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
// Code generated by k8s-infra-gen. DO NOT EDIT.
package v20150501preview

import (
	"github.com/Azure/k8s-infra/hack/crossplane/apis/deploymenttemplate/v20150101"
	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
type ServersFailoverGroups struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServersFailoverGroups_Spec `json:"spec,omitempty"`
	Status            FailoverGroup_Status       `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ServersFailoverGroupsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServersFailoverGroups `json:"items"`
}

type FailoverGroup_Status struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              ServersFailoverGroupsObservation `json:"atProvider"`
}

type ServersFailoverGroups_Spec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           ServersFailoverGroupsParameters `json:"forProvider"`
}

type ServersFailoverGroupsObservation struct {

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Location: Resource location.
	Location *string `json:"location,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Resource properties.
	Properties *FailoverGroupProperties_Status `json:"properties,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type ServersFailoverGroupsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion ServersFailoverGroupsSpecApiVersion `json:"apiVersion"`
	Comments   *string                             `json:"comments,omitempty"`

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
	//Properties: Properties of a failover group.
	Properties                FailoverGroupProperties `json:"properties"`
	ResourceGroupName         string                  `json:"resourceGroupName"`
	ResourceGroupNameRef      *v1alpha1.Reference     `json:"resourceGroupNameRef,omitempty"`
	ResourceGroupNameSelector *v1alpha1.Selector      `json:"resourceGroupNameSelector,omitempty"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope               *string             `json:"scope,omitempty"`
	ServersName         string              `json:"serversName"`
	ServersNameRef      *v1alpha1.Reference `json:"serversNameRef,omitempty"`
	ServersNameSelector *v1alpha1.Selector  `json:"serversNameSelector,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type ServersFailoverGroupsSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2015-05-01-preview/Microsoft.Sql.json#/definitions/FailoverGroupProperties
type FailoverGroupProperties struct {

	//Databases: List of databases in the failover group.
	Databases []string `json:"databases,omitempty"`

	// +kubebuilder:validation:Required
	//PartnerServers: List of partner server information for the failover group.
	PartnerServers []PartnerInfo `json:"partnerServers"`

	//ReadOnlyEndpoint: Read-only endpoint of the failover group instance.
	ReadOnlyEndpoint *FailoverGroupReadOnlyEndpoint `json:"readOnlyEndpoint,omitempty"`

	// +kubebuilder:validation:Required
	//ReadWriteEndpoint: Read-write endpoint of the failover group instance.
	ReadWriteEndpoint FailoverGroupReadWriteEndpoint `json:"readWriteEndpoint"`
}

//Generated from:
type FailoverGroupProperties_Status struct {

	//Databases: List of databases in the failover group.
	Databases []string `json:"databases,omitempty"`

	// +kubebuilder:validation:Required
	//PartnerServers: List of partner server information for the failover group.
	PartnerServers []PartnerInfo_Status `json:"partnerServers"`

	//ReadOnlyEndpoint: Read-only endpoint of the failover group instance.
	ReadOnlyEndpoint *FailoverGroupReadOnlyEndpoint_Status `json:"readOnlyEndpoint,omitempty"`

	// +kubebuilder:validation:Required
	//ReadWriteEndpoint: Read-write endpoint of the failover group instance.
	ReadWriteEndpoint FailoverGroupReadWriteEndpoint_Status `json:"readWriteEndpoint"`

	//ReplicationRole: Local replication role of the failover group instance.
	ReplicationRole *FailoverGroupPropertiesStatusReplicationRole `json:"replicationRole,omitempty"`

	//ReplicationState: Replication state of the failover group instance.
	ReplicationState *string `json:"replicationState,omitempty"`
}

// +kubebuilder:validation:Enum={"2015-05-01-preview"}
type ServersFailoverGroupsSpecApiVersion string

const ServersFailoverGroupsSpecApiVersion20150501Preview = ServersFailoverGroupsSpecApiVersion("2015-05-01-preview")

// +kubebuilder:validation:Enum={"Microsoft.Sql/servers/failoverGroups"}
type ServersFailoverGroupsSpecType string

const ServersFailoverGroupsSpecTypeMicrosoftSqlServersFailoverGroups = ServersFailoverGroupsSpecType("Microsoft.Sql/servers/failoverGroups")

// +kubebuilder:validation:Enum={"Primary","Secondary"}
type FailoverGroupPropertiesStatusReplicationRole string

const (
	FailoverGroupPropertiesStatusReplicationRolePrimary   = FailoverGroupPropertiesStatusReplicationRole("Primary")
	FailoverGroupPropertiesStatusReplicationRoleSecondary = FailoverGroupPropertiesStatusReplicationRole("Secondary")
)

//Generated from: https://schema.management.azure.com/schemas/2015-05-01-preview/Microsoft.Sql.json#/definitions/FailoverGroupReadOnlyEndpoint
type FailoverGroupReadOnlyEndpoint struct {

	//FailoverPolicy: Failover policy of the read-only endpoint for the failover group.
	FailoverPolicy *FailoverGroupReadOnlyEndpointFailoverPolicy `json:"failoverPolicy,omitempty"`
}

//Generated from:
type FailoverGroupReadOnlyEndpoint_Status struct {

	//FailoverPolicy: Failover policy of the read-only endpoint for the failover group.
	FailoverPolicy *FailoverGroupReadOnlyEndpointStatusFailoverPolicy `json:"failoverPolicy,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2015-05-01-preview/Microsoft.Sql.json#/definitions/FailoverGroupReadWriteEndpoint
type FailoverGroupReadWriteEndpoint struct {

	// +kubebuilder:validation:Required
	//FailoverPolicy: Failover policy of the read-write endpoint for the failover
	//group. If failoverPolicy is Automatic then
	//failoverWithDataLossGracePeriodMinutes is required.
	FailoverPolicy FailoverGroupReadWriteEndpointFailoverPolicy `json:"failoverPolicy"`

	//FailoverWithDataLossGracePeriodMinutes: Grace period before failover with data
	//loss is attempted for the read-write endpoint. If failoverPolicy is Automatic
	//then failoverWithDataLossGracePeriodMinutes is required.
	FailoverWithDataLossGracePeriodMinutes *int `json:"failoverWithDataLossGracePeriodMinutes,omitempty"`
}

//Generated from:
type FailoverGroupReadWriteEndpoint_Status struct {

	// +kubebuilder:validation:Required
	//FailoverPolicy: Failover policy of the read-write endpoint for the failover
	//group. If failoverPolicy is Automatic then
	//failoverWithDataLossGracePeriodMinutes is required.
	FailoverPolicy FailoverGroupReadWriteEndpointStatusFailoverPolicy `json:"failoverPolicy"`

	//FailoverWithDataLossGracePeriodMinutes: Grace period before failover with data
	//loss is attempted for the read-write endpoint. If failoverPolicy is Automatic
	//then failoverWithDataLossGracePeriodMinutes is required.
	FailoverWithDataLossGracePeriodMinutes *int `json:"failoverWithDataLossGracePeriodMinutes,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2015-05-01-preview/Microsoft.Sql.json#/definitions/PartnerInfo
type PartnerInfo struct {

	// +kubebuilder:validation:Required
	//Id: Resource identifier of the partner server.
	Id string `json:"id"`
}

//Generated from:
type PartnerInfo_Status struct {

	// +kubebuilder:validation:Required
	//Id: Resource identifier of the partner server.
	Id string `json:"id"`

	//Location: Geo location of the partner server.
	Location *string `json:"location,omitempty"`

	//ReplicationRole: Replication role of the partner server.
	ReplicationRole *PartnerInfoStatusReplicationRole `json:"replicationRole,omitempty"`
}

// +kubebuilder:validation:Enum={"Disabled","Enabled"}
type FailoverGroupReadOnlyEndpointFailoverPolicy string

const (
	FailoverGroupReadOnlyEndpointFailoverPolicyDisabled = FailoverGroupReadOnlyEndpointFailoverPolicy("Disabled")
	FailoverGroupReadOnlyEndpointFailoverPolicyEnabled  = FailoverGroupReadOnlyEndpointFailoverPolicy("Enabled")
)

// +kubebuilder:validation:Enum={"Disabled","Enabled"}
type FailoverGroupReadOnlyEndpointStatusFailoverPolicy string

const (
	FailoverGroupReadOnlyEndpointStatusFailoverPolicyDisabled = FailoverGroupReadOnlyEndpointStatusFailoverPolicy("Disabled")
	FailoverGroupReadOnlyEndpointStatusFailoverPolicyEnabled  = FailoverGroupReadOnlyEndpointStatusFailoverPolicy("Enabled")
)

// +kubebuilder:validation:Enum={"Automatic","Manual"}
type FailoverGroupReadWriteEndpointFailoverPolicy string

const (
	FailoverGroupReadWriteEndpointFailoverPolicyAutomatic = FailoverGroupReadWriteEndpointFailoverPolicy("Automatic")
	FailoverGroupReadWriteEndpointFailoverPolicyManual    = FailoverGroupReadWriteEndpointFailoverPolicy("Manual")
)

// +kubebuilder:validation:Enum={"Automatic","Manual"}
type FailoverGroupReadWriteEndpointStatusFailoverPolicy string

const (
	FailoverGroupReadWriteEndpointStatusFailoverPolicyAutomatic = FailoverGroupReadWriteEndpointStatusFailoverPolicy("Automatic")
	FailoverGroupReadWriteEndpointStatusFailoverPolicyManual    = FailoverGroupReadWriteEndpointStatusFailoverPolicy("Manual")
)

// +kubebuilder:validation:Enum={"Primary","Secondary"}
type PartnerInfoStatusReplicationRole string

const (
	PartnerInfoStatusReplicationRolePrimary   = PartnerInfoStatusReplicationRole("Primary")
	PartnerInfoStatusReplicationRoleSecondary = PartnerInfoStatusReplicationRole("Secondary")
)

func init() {
	SchemeBuilder.Register(&ServersFailoverGroups{}, &ServersFailoverGroupsList{})
}
