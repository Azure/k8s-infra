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
type NetworkWatchersFlowLogs struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              NetworkWatchersFlowLogs_Spec `json:"spec,omitempty"`
	Status            FlowLog_Status               `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type NetworkWatchersFlowLogsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkWatchersFlowLogs `json:"items"`
}

type FlowLog_Status struct {
	AtProvider NetworkWatchersFlowLogsObservation `json:"atProvider"`
}

type NetworkWatchersFlowLogs_Spec struct {
	ForProvider NetworkWatchersFlowLogsParameters `json:"forProvider"`
}

type NetworkWatchersFlowLogsObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Location: Resource location.
	Location *string `json:"location,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the flow log.
	Properties *FlowLogPropertiesFormat_Status `json:"properties,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type NetworkWatchersFlowLogsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion NetworkWatchersFlowLogsSpecApiVersion `json:"apiVersion"`
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
	//Properties: Properties of the flow log.
	Properties FlowLogPropertiesFormat `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type NetworkWatchersFlowLogsSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/FlowLogPropertiesFormat
type FlowLogPropertiesFormat struct {

	//Enabled: Flag to enable/disable flow logging.
	Enabled *bool `json:"enabled,omitempty"`

	//FlowAnalyticsConfiguration: Parameters that define the configuration of traffic
	//analytics.
	FlowAnalyticsConfiguration *TrafficAnalyticsProperties `json:"flowAnalyticsConfiguration,omitempty"`

	//Format: Parameters that define the flow log format.
	Format *FlowLogFormatParameters `json:"format,omitempty"`

	//RetentionPolicy: Parameters that define the retention policy for flow log.
	RetentionPolicy *RetentionPolicyParameters `json:"retentionPolicy,omitempty"`

	// +kubebuilder:validation:Required
	//StorageId: ID of the storage account which is used to store the flow log.
	StorageId string `json:"storageId"`

	// +kubebuilder:validation:Required
	//TargetResourceId: ID of network security group to which flow log will be applied.
	TargetResourceId string `json:"targetResourceId"`
}

//Generated from:
type FlowLogPropertiesFormat_Status struct {

	//Enabled: Flag to enable/disable flow logging.
	Enabled *bool `json:"enabled,omitempty"`

	//FlowAnalyticsConfiguration: Parameters that define the configuration of traffic
	//analytics.
	FlowAnalyticsConfiguration *TrafficAnalyticsProperties_Status `json:"flowAnalyticsConfiguration,omitempty"`

	//Format: Parameters that define the flow log format.
	Format *FlowLogFormatParameters_Status `json:"format,omitempty"`

	//ProvisioningState: The provisioning state of the flow log.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//RetentionPolicy: Parameters that define the retention policy for flow log.
	RetentionPolicy *RetentionPolicyParameters_Status `json:"retentionPolicy,omitempty"`

	// +kubebuilder:validation:Required
	//StorageId: ID of the storage account which is used to store the flow log.
	StorageId string `json:"storageId"`

	//TargetResourceGuid: Guid of network security group to which flow log will be
	//applied.
	TargetResourceGuid *string `json:"targetResourceGuid,omitempty"`

	// +kubebuilder:validation:Required
	//TargetResourceId: ID of network security group to which flow log will be applied.
	TargetResourceId string `json:"targetResourceId"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type NetworkWatchersFlowLogsSpecApiVersion string

const NetworkWatchersFlowLogsSpecApiVersion20200501 = NetworkWatchersFlowLogsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/networkWatchers/flowLogs"}
type NetworkWatchersFlowLogsSpecType string

const NetworkWatchersFlowLogsSpecTypeMicrosoftNetworkNetworkWatchersFlowLogs = NetworkWatchersFlowLogsSpecType("Microsoft.Network/networkWatchers/flowLogs")

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/FlowLogFormatParameters
type FlowLogFormatParameters struct {

	//Type: The file type of flow log.
	Type *FlowLogFormatParametersType `json:"type,omitempty"`

	//Version: The version (revision) of the flow log.
	Version *int `json:"version,omitempty"`
}

//Generated from:
type FlowLogFormatParameters_Status struct {

	//Type: The file type of flow log.
	Type *FlowLogFormatParametersStatusType `json:"type,omitempty"`

	//Version: The version (revision) of the flow log.
	Version *int `json:"version,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/RetentionPolicyParameters
type RetentionPolicyParameters struct {

	//Days: Number of days to retain flow log records.
	Days *int `json:"days,omitempty"`

	//Enabled: Flag to enable/disable retention.
	Enabled *bool `json:"enabled,omitempty"`
}

//Generated from:
type RetentionPolicyParameters_Status struct {

	//Days: Number of days to retain flow log records.
	Days *int `json:"days,omitempty"`

	//Enabled: Flag to enable/disable retention.
	Enabled *bool `json:"enabled,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/TrafficAnalyticsProperties
type TrafficAnalyticsProperties struct {

	//NetworkWatcherFlowAnalyticsConfiguration: Parameters that define the
	//configuration of traffic analytics.
	NetworkWatcherFlowAnalyticsConfiguration *TrafficAnalyticsConfigurationProperties `json:"networkWatcherFlowAnalyticsConfiguration,omitempty"`
}

//Generated from:
type TrafficAnalyticsProperties_Status struct {

	//NetworkWatcherFlowAnalyticsConfiguration: Parameters that define the
	//configuration of traffic analytics.
	NetworkWatcherFlowAnalyticsConfiguration *TrafficAnalyticsConfigurationProperties_Status `json:"networkWatcherFlowAnalyticsConfiguration,omitempty"`
}

// +kubebuilder:validation:Enum={"JSON"}
type FlowLogFormatParametersStatusType string

const FlowLogFormatParametersStatusTypeJSON = FlowLogFormatParametersStatusType("JSON")

// +kubebuilder:validation:Enum={"JSON"}
type FlowLogFormatParametersType string

const FlowLogFormatParametersTypeJSON = FlowLogFormatParametersType("JSON")

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/TrafficAnalyticsConfigurationProperties
type TrafficAnalyticsConfigurationProperties struct {

	//Enabled: Flag to enable/disable traffic analytics.
	Enabled *bool `json:"enabled,omitempty"`

	//TrafficAnalyticsInterval: The interval in minutes which would decide how
	//frequently TA service should do flow analytics.
	TrafficAnalyticsInterval *int `json:"trafficAnalyticsInterval,omitempty"`

	//WorkspaceId: The resource guid of the attached workspace.
	WorkspaceId *string `json:"workspaceId,omitempty"`

	//WorkspaceRegion: The location of the attached workspace.
	WorkspaceRegion *string `json:"workspaceRegion,omitempty"`

	//WorkspaceResourceId: Resource Id of the attached workspace.
	WorkspaceResourceId *string `json:"workspaceResourceId,omitempty"`
}

//Generated from:
type TrafficAnalyticsConfigurationProperties_Status struct {

	//Enabled: Flag to enable/disable traffic analytics.
	Enabled *bool `json:"enabled,omitempty"`

	//TrafficAnalyticsInterval: The interval in minutes which would decide how
	//frequently TA service should do flow analytics.
	TrafficAnalyticsInterval *int `json:"trafficAnalyticsInterval,omitempty"`

	//WorkspaceId: The resource guid of the attached workspace.
	WorkspaceId *string `json:"workspaceId,omitempty"`

	//WorkspaceRegion: The location of the attached workspace.
	WorkspaceRegion *string `json:"workspaceRegion,omitempty"`

	//WorkspaceResourceId: Resource Id of the attached workspace.
	WorkspaceResourceId *string `json:"workspaceResourceId,omitempty"`
}

func init() {
	SchemeBuilder.Register(&NetworkWatchersFlowLogs{}, &NetworkWatchersFlowLogsList{})
}
