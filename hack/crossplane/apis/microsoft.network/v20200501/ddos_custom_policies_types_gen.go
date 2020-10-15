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
type DdosCustomPolicies struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DdosCustomPolicies_Spec `json:"spec,omitempty"`
	Status            DdosCustomPolicy_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type DdosCustomPoliciesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DdosCustomPolicies `json:"items"`
}

type DdosCustomPolicies_Spec struct {
	ForProvider DdosCustomPoliciesParameters `json:"forProvider"`
}

type DdosCustomPolicy_Status struct {
	AtProvider DdosCustomPoliciesObservation `json:"atProvider"`
}

type DdosCustomPoliciesObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Location: Resource location.
	Location *string `json:"location,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the DDoS custom policy.
	Properties *DdosCustomPolicyPropertiesFormat_Status `json:"properties,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type DdosCustomPoliciesParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion DdosCustomPoliciesSpecApiVersion `json:"apiVersion"`
	Comments   *string                          `json:"comments,omitempty"`

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
	//Properties: Properties of the DDoS custom policy.
	Properties DdosCustomPolicyPropertiesFormat `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type DdosCustomPoliciesSpecType `json:"type"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type DdosCustomPoliciesSpecApiVersion string

const DdosCustomPoliciesSpecApiVersion20200501 = DdosCustomPoliciesSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/ddosCustomPolicies"}
type DdosCustomPoliciesSpecType string

const DdosCustomPoliciesSpecTypeMicrosoftNetworkDdosCustomPolicies = DdosCustomPoliciesSpecType("Microsoft.Network/ddosCustomPolicies")

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/DdosCustomPolicyPropertiesFormat
type DdosCustomPolicyPropertiesFormat struct {

	//ProtocolCustomSettings: The protocol-specific DDoS policy customization
	//parameters.
	ProtocolCustomSettings []ProtocolCustomSettingsFormat `json:"protocolCustomSettings,omitempty"`
}

//Generated from:
type DdosCustomPolicyPropertiesFormat_Status struct {

	//ProtocolCustomSettings: The protocol-specific DDoS policy customization
	//parameters.
	ProtocolCustomSettings []ProtocolCustomSettingsFormat_Status `json:"protocolCustomSettings,omitempty"`

	//ProvisioningState: The provisioning state of the DDoS custom policy resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//PublicIPAddresses: The list of public IPs associated with the DDoS custom policy
	//resource. This list is read-only.
	PublicIPAddresses []SubResource_Status `json:"publicIPAddresses,omitempty"`

	//ResourceGuid: The resource GUID property of the DDoS custom policy resource. It
	//uniquely identifies the resource, even if the user changes its name or migrate
	//the resource across subscriptions or resource groups.
	ResourceGuid *string `json:"resourceGuid,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/ProtocolCustomSettingsFormat
type ProtocolCustomSettingsFormat struct {

	//Protocol: The protocol for which the DDoS protection policy is being customized.
	Protocol *ProtocolCustomSettingsFormatProtocol `json:"protocol,omitempty"`

	//SourceRateOverride: The customized DDoS protection source rate.
	SourceRateOverride *string `json:"sourceRateOverride,omitempty"`

	//TriggerRateOverride: The customized DDoS protection trigger rate.
	TriggerRateOverride *string `json:"triggerRateOverride,omitempty"`

	//TriggerSensitivityOverride: The customized DDoS protection trigger rate
	//sensitivity degrees. High: Trigger rate set with most sensitivity w.r.t. normal
	//traffic. Default: Trigger rate set with moderate sensitivity w.r.t. normal
	//traffic. Low: Trigger rate set with less sensitivity w.r.t. normal traffic.
	//Relaxed: Trigger rate set with least sensitivity w.r.t. normal traffic.
	TriggerSensitivityOverride *ProtocolCustomSettingsFormatTriggerSensitivityOverride `json:"triggerSensitivityOverride,omitempty"`
}

//Generated from:
type ProtocolCustomSettingsFormat_Status struct {

	//Protocol: The protocol for which the DDoS protection policy is being customized.
	Protocol *ProtocolCustomSettingsFormatStatusProtocol `json:"protocol,omitempty"`

	//SourceRateOverride: The customized DDoS protection source rate.
	SourceRateOverride *string `json:"sourceRateOverride,omitempty"`

	//TriggerRateOverride: The customized DDoS protection trigger rate.
	TriggerRateOverride *string `json:"triggerRateOverride,omitempty"`

	//TriggerSensitivityOverride: The customized DDoS protection trigger rate
	//sensitivity degrees. High: Trigger rate set with most sensitivity w.r.t. normal
	//traffic. Default: Trigger rate set with moderate sensitivity w.r.t. normal
	//traffic. Low: Trigger rate set with less sensitivity w.r.t. normal traffic.
	//Relaxed: Trigger rate set with least sensitivity w.r.t. normal traffic.
	TriggerSensitivityOverride *ProtocolCustomSettingsFormatStatusTriggerSensitivityOverride `json:"triggerSensitivityOverride,omitempty"`
}

// +kubebuilder:validation:Enum={"Syn","Tcp","Udp"}
type ProtocolCustomSettingsFormatProtocol string

const (
	ProtocolCustomSettingsFormatProtocolSyn = ProtocolCustomSettingsFormatProtocol("Syn")
	ProtocolCustomSettingsFormatProtocolTcp = ProtocolCustomSettingsFormatProtocol("Tcp")
	ProtocolCustomSettingsFormatProtocolUdp = ProtocolCustomSettingsFormatProtocol("Udp")
)

// +kubebuilder:validation:Enum={"Syn","Tcp","Udp"}
type ProtocolCustomSettingsFormatStatusProtocol string

const (
	ProtocolCustomSettingsFormatStatusProtocolSyn = ProtocolCustomSettingsFormatStatusProtocol("Syn")
	ProtocolCustomSettingsFormatStatusProtocolTcp = ProtocolCustomSettingsFormatStatusProtocol("Tcp")
	ProtocolCustomSettingsFormatStatusProtocolUdp = ProtocolCustomSettingsFormatStatusProtocol("Udp")
)

// +kubebuilder:validation:Enum={"Default","High","Low","Relaxed"}
type ProtocolCustomSettingsFormatStatusTriggerSensitivityOverride string

const (
	ProtocolCustomSettingsFormatStatusTriggerSensitivityOverrideDefault = ProtocolCustomSettingsFormatStatusTriggerSensitivityOverride("Default")
	ProtocolCustomSettingsFormatStatusTriggerSensitivityOverrideHigh    = ProtocolCustomSettingsFormatStatusTriggerSensitivityOverride("High")
	ProtocolCustomSettingsFormatStatusTriggerSensitivityOverrideLow     = ProtocolCustomSettingsFormatStatusTriggerSensitivityOverride("Low")
	ProtocolCustomSettingsFormatStatusTriggerSensitivityOverrideRelaxed = ProtocolCustomSettingsFormatStatusTriggerSensitivityOverride("Relaxed")
)

// +kubebuilder:validation:Enum={"Default","High","Low","Relaxed"}
type ProtocolCustomSettingsFormatTriggerSensitivityOverride string

const (
	ProtocolCustomSettingsFormatTriggerSensitivityOverrideDefault = ProtocolCustomSettingsFormatTriggerSensitivityOverride("Default")
	ProtocolCustomSettingsFormatTriggerSensitivityOverrideHigh    = ProtocolCustomSettingsFormatTriggerSensitivityOverride("High")
	ProtocolCustomSettingsFormatTriggerSensitivityOverrideLow     = ProtocolCustomSettingsFormatTriggerSensitivityOverride("Low")
	ProtocolCustomSettingsFormatTriggerSensitivityOverrideRelaxed = ProtocolCustomSettingsFormatTriggerSensitivityOverride("Relaxed")
)

func init() {
	SchemeBuilder.Register(&DdosCustomPolicies{}, &DdosCustomPoliciesList{})
}
