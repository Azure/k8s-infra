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
type LoadBalancersInboundNatRules struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              LoadBalancersInboundNatRules_Spec `json:"spec,omitempty"`
	Status            InboundNatRule_Status             `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type LoadBalancersInboundNatRulesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoadBalancersInboundNatRules `json:"items"`
}

type InboundNatRule_Status struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              LoadBalancersInboundNatRulesObservation `json:"atProvider"`
}

type LoadBalancersInboundNatRules_Spec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           LoadBalancersInboundNatRulesParameters `json:"forProvider"`
}

type LoadBalancersInboundNatRulesObservation struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: The name of the resource that is unique within the set of inbound NAT
	//rules used by the load balancer. This name can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of load balancer inbound nat rule.
	Properties *InboundNatRulePropertiesFormat_Status `json:"properties,omitempty"`

	//Type: Type of the resource.
	Type *string `json:"type,omitempty"`
}

type LoadBalancersInboundNatRulesParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion LoadBalancersInboundNatRulesSpecApiVersion `json:"apiVersion"`
	Comments   *string                                    `json:"comments,omitempty"`

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
	//Properties: Properties of load balancer inbound nat rule.
	Properties InboundNatRulePropertiesFormat `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type LoadBalancersInboundNatRulesSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/InboundNatRulePropertiesFormat
type InboundNatRulePropertiesFormat struct {

	// +kubebuilder:validation:Required
	//BackendPort: The port used for the internal endpoint. Acceptable values range
	//from 1 to 65535.
	BackendPort int `json:"backendPort"`

	//EnableFloatingIP: Configures a virtual machine's endpoint for the floating IP
	//capability required to configure a SQL AlwaysOn Availability Group. This setting
	//is required when using the SQL AlwaysOn Availability Groups in SQL server. This
	//setting can't be changed after you create the endpoint.
	EnableFloatingIP *bool `json:"enableFloatingIP,omitempty"`

	//EnableTcpReset: Receive bidirectional TCP Reset on TCP flow idle timeout or
	//unexpected connection termination. This element is only used when the protocol
	//is set to TCP.
	EnableTcpReset *bool `json:"enableTcpReset,omitempty"`

	// +kubebuilder:validation:Required
	//FrontendIPConfiguration: A reference to frontend IP addresses.
	FrontendIPConfiguration SubResource `json:"frontendIPConfiguration"`

	// +kubebuilder:validation:Required
	//FrontendPort: The port for the external endpoint. Port numbers for each rule
	//must be unique within the Load Balancer. Acceptable values range from 1 to 65534.
	FrontendPort int `json:"frontendPort"`

	//IdleTimeoutInMinutes: The timeout for the TCP idle connection. The value can be
	//set between 4 and 30 minutes. The default value is 4 minutes. This element is
	//only used when the protocol is set to TCP.
	IdleTimeoutInMinutes *int `json:"idleTimeoutInMinutes,omitempty"`

	// +kubebuilder:validation:Required
	//Protocol: The reference to the transport protocol used by the load balancing
	//rule.
	Protocol InboundNatRulePropertiesFormatProtocol `json:"protocol"`
}

//Generated from:
type InboundNatRulePropertiesFormat_Status struct {

	//BackendIPConfiguration: A reference to a private IP address defined on a network
	//interface of a VM. Traffic sent to the frontend port of each of the frontend IP
	//configurations is forwarded to the backend IP.
	BackendIPConfiguration *NetworkInterfaceIPConfiguration_Status `json:"backendIPConfiguration,omitempty"`

	//BackendPort: The port used for the internal endpoint. Acceptable values range
	//from 1 to 65535.
	BackendPort *int `json:"backendPort,omitempty"`

	//EnableFloatingIP: Configures a virtual machine's endpoint for the floating IP
	//capability required to configure a SQL AlwaysOn Availability Group. This setting
	//is required when using the SQL AlwaysOn Availability Groups in SQL server. This
	//setting can't be changed after you create the endpoint.
	EnableFloatingIP *bool `json:"enableFloatingIP,omitempty"`

	//EnableTcpReset: Receive bidirectional TCP Reset on TCP flow idle timeout or
	//unexpected connection termination. This element is only used when the protocol
	//is set to TCP.
	EnableTcpReset *bool `json:"enableTcpReset,omitempty"`

	//FrontendIPConfiguration: A reference to frontend IP addresses.
	FrontendIPConfiguration *SubResource_Status `json:"frontendIPConfiguration,omitempty"`

	//FrontendPort: The port for the external endpoint. Port numbers for each rule
	//must be unique within the Load Balancer. Acceptable values range from 1 to 65534.
	FrontendPort *int `json:"frontendPort,omitempty"`

	//IdleTimeoutInMinutes: The timeout for the TCP idle connection. The value can be
	//set between 4 and 30 minutes. The default value is 4 minutes. This element is
	//only used when the protocol is set to TCP.
	IdleTimeoutInMinutes *int `json:"idleTimeoutInMinutes,omitempty"`

	//Protocol: The reference to the transport protocol used by the load balancing
	//rule.
	Protocol *TransportProtocol_Status `json:"protocol,omitempty"`

	//ProvisioningState: The provisioning state of the inbound NAT rule resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type LoadBalancersInboundNatRulesSpecApiVersion string

const LoadBalancersInboundNatRulesSpecApiVersion20200501 = LoadBalancersInboundNatRulesSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/loadBalancers/inboundNatRules"}
type LoadBalancersInboundNatRulesSpecType string

const LoadBalancersInboundNatRulesSpecTypeMicrosoftNetworkLoadBalancersInboundNatRules = LoadBalancersInboundNatRulesSpecType("Microsoft.Network/loadBalancers/inboundNatRules")

// +kubebuilder:validation:Enum={"All","Tcp","Udp"}
type InboundNatRulePropertiesFormatProtocol string

const (
	InboundNatRulePropertiesFormatProtocolAll = InboundNatRulePropertiesFormatProtocol("All")
	InboundNatRulePropertiesFormatProtocolTcp = InboundNatRulePropertiesFormatProtocol("Tcp")
	InboundNatRulePropertiesFormatProtocolUdp = InboundNatRulePropertiesFormatProtocol("Udp")
)

//Generated from:
// +kubebuilder:validation:Enum={"All","Tcp","Udp"}
type TransportProtocol_Status string

const (
	TransportProtocol_StatusAll = TransportProtocol_Status("All")
	TransportProtocol_StatusTcp = TransportProtocol_Status("Tcp")
	TransportProtocol_StatusUdp = TransportProtocol_Status("Udp")
)

func init() {
	SchemeBuilder.Register(&LoadBalancersInboundNatRules{}, &LoadBalancersInboundNatRulesList{})
}
