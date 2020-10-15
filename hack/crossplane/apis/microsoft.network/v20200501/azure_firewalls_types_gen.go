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
type AzureFirewalls struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AzureFirewalls_Spec  `json:"spec,omitempty"`
	Status            AzureFirewall_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type AzureFirewallsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureFirewalls `json:"items"`
}

//Generated from:
type AzureFirewall_Status struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Location: Resource location.
	Location *string `json:"location,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the azure firewall.
	Properties *AzureFirewallPropertiesFormat_Status `json:"properties,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`

	//Zones: A list of availability zones denoting where the resource needs to come
	//from.
	Zones []string `json:"zones,omitempty"`
}

type AzureFirewalls_Spec struct {
	ForProvider AzureFirewallsParameters `json:"forProvider"`
}

//Generated from:
type AzureFirewallPropertiesFormat_Status struct {

	//AdditionalProperties: The additional properties used to further config this
	//azure firewall.
	AdditionalProperties map[string]string `json:"additionalProperties,omitempty"`

	//ApplicationRuleCollections: Collection of application rule collections used by
	//Azure Firewall.
	ApplicationRuleCollections []AzureFirewallApplicationRuleCollection_Status `json:"applicationRuleCollections,omitempty"`

	//FirewallPolicy: The firewallPolicy associated with this azure firewall.
	FirewallPolicy *SubResource_Status `json:"firewallPolicy,omitempty"`

	//HubIPAddresses: IP addresses associated with AzureFirewall.
	HubIPAddresses *HubIPAddresses_Status `json:"hubIPAddresses,omitempty"`

	//IpConfigurations: IP configuration of the Azure Firewall resource.
	IpConfigurations []AzureFirewallIPConfiguration_Status `json:"ipConfigurations,omitempty"`

	//IpGroups: IpGroups associated with AzureFirewall.
	IpGroups []AzureFirewallIpGroups_Status `json:"ipGroups,omitempty"`

	//ManagementIpConfiguration: IP configuration of the Azure Firewall used for
	//management traffic.
	ManagementIpConfiguration *AzureFirewallIPConfiguration_Status `json:"managementIpConfiguration,omitempty"`

	//NatRuleCollections: Collection of NAT rule collections used by Azure Firewall.
	NatRuleCollections []AzureFirewallNatRuleCollection_Status `json:"natRuleCollections,omitempty"`

	//NetworkRuleCollections: Collection of network rule collections used by Azure
	//Firewall.
	NetworkRuleCollections []AzureFirewallNetworkRuleCollection_Status `json:"networkRuleCollections,omitempty"`

	//ProvisioningState: The provisioning state of the Azure firewall resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//Sku: The Azure Firewall Resource SKU.
	Sku *AzureFirewallSku_Status `json:"sku,omitempty"`

	//ThreatIntelMode: The operation mode for Threat Intelligence.
	ThreatIntelMode *AzureFirewallThreatIntelMode_Status `json:"threatIntelMode,omitempty"`

	//VirtualHub: The virtualHub to which the firewall belongs.
	VirtualHub *SubResource_Status `json:"virtualHub,omitempty"`
}

type AzureFirewallsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion AzureFirewallsSpecApiVersion `json:"apiVersion"`
	Comments   *string                      `json:"comments,omitempty"`

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
	//Properties: Properties of the azure firewall.
	Properties AzureFirewallPropertiesFormat `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type AzureFirewallsSpecType `json:"type"`

	//Zones: A list of availability zones denoting where the resource needs to come
	//from.
	Zones []string `json:"zones,omitempty"`
}

//Generated from:
type AzureFirewallApplicationRuleCollection_Status struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: The name of the resource that is unique within the Azure firewall. This
	//name can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the azure firewall application rule collection.
	Properties *AzureFirewallApplicationRuleCollectionPropertiesFormat_Status `json:"properties,omitempty"`
}

//Generated from:
type AzureFirewallIPConfiguration_Status struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: Name of the resource that is unique within a resource group. This name can
	//be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the azure firewall IP configuration.
	Properties *AzureFirewallIPConfigurationPropertiesFormat_Status `json:"properties,omitempty"`

	//Type: Type of the resource.
	Type *string `json:"type,omitempty"`
}

//Generated from:
type AzureFirewallIpGroups_Status struct {

	//ChangeNumber: The iteration number.
	ChangeNumber *string `json:"changeNumber,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`
}

//Generated from:
type AzureFirewallNatRuleCollection_Status struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: The name of the resource that is unique within the Azure firewall. This
	//name can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the azure firewall NAT rule collection.
	Properties *AzureFirewallNatRuleCollectionProperties_Status `json:"properties,omitempty"`
}

//Generated from:
type AzureFirewallNetworkRuleCollection_Status struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: The name of the resource that is unique within the Azure firewall. This
	//name can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the azure firewall network rule collection.
	Properties *AzureFirewallNetworkRuleCollectionPropertiesFormat_Status `json:"properties,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AzureFirewallPropertiesFormat
type AzureFirewallPropertiesFormat struct {

	//AdditionalProperties: The additional properties used to further config this
	//azure firewall.
	AdditionalProperties map[string]string `json:"additionalProperties,omitempty"`

	//ApplicationRuleCollections: Collection of application rule collections used by
	//Azure Firewall.
	ApplicationRuleCollections []AzureFirewallApplicationRuleCollection `json:"applicationRuleCollections,omitempty"`

	//FirewallPolicy: The firewallPolicy associated with this azure firewall.
	FirewallPolicy *SubResource `json:"firewallPolicy,omitempty"`

	//HubIPAddresses: IP addresses associated with AzureFirewall.
	HubIPAddresses *HubIPAddresses `json:"hubIPAddresses,omitempty"`

	//IpConfigurations: IP configuration of the Azure Firewall resource.
	IpConfigurations []AzureFirewallIPConfiguration `json:"ipConfigurations,omitempty"`

	//ManagementIpConfiguration: IP configuration of the Azure Firewall used for
	//management traffic.
	ManagementIpConfiguration *AzureFirewallIPConfiguration `json:"managementIpConfiguration,omitempty"`

	//NatRuleCollections: Collection of NAT rule collections used by Azure Firewall.
	NatRuleCollections []AzureFirewallNatRuleCollection `json:"natRuleCollections,omitempty"`

	//NetworkRuleCollections: Collection of network rule collections used by Azure
	//Firewall.
	NetworkRuleCollections []AzureFirewallNetworkRuleCollection `json:"networkRuleCollections,omitempty"`

	//Sku: The Azure Firewall Resource SKU.
	Sku *AzureFirewallSku `json:"sku,omitempty"`

	//ThreatIntelMode: The operation mode for Threat Intelligence.
	ThreatIntelMode *AzureFirewallPropertiesFormatThreatIntelMode `json:"threatIntelMode,omitempty"`

	//VirtualHub: The virtualHub to which the firewall belongs.
	VirtualHub *SubResource `json:"virtualHub,omitempty"`
}

//Generated from:
type AzureFirewallSku_Status struct {

	//Name: Name of an Azure Firewall SKU.
	Name *AzureFirewallSkuStatusName `json:"name,omitempty"`

	//Tier: Tier of an Azure Firewall.
	Tier *AzureFirewallSkuStatusTier `json:"tier,omitempty"`
}

//Generated from:
// +kubebuilder:validation:Enum={"Alert","Deny","Off"}
type AzureFirewallThreatIntelMode_Status string

const (
	AzureFirewallThreatIntelMode_StatusAlert = AzureFirewallThreatIntelMode_Status("Alert")
	AzureFirewallThreatIntelMode_StatusDeny  = AzureFirewallThreatIntelMode_Status("Deny")
	AzureFirewallThreatIntelMode_StatusOff   = AzureFirewallThreatIntelMode_Status("Off")
)

// +kubebuilder:validation:Enum={"2020-05-01"}
type AzureFirewallsSpecApiVersion string

const AzureFirewallsSpecApiVersion20200501 = AzureFirewallsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/azureFirewalls"}
type AzureFirewallsSpecType string

const AzureFirewallsSpecTypeMicrosoftNetworkAzureFirewalls = AzureFirewallsSpecType("Microsoft.Network/azureFirewalls")

//Generated from:
type HubIPAddresses_Status struct {

	//PrivateIPAddress: Private IP Address associated with azure firewall.
	PrivateIPAddress *string `json:"privateIPAddress,omitempty"`

	//PublicIPs: Public IP addresses associated with azure firewall.
	PublicIPs *HubPublicIPAddresses_Status `json:"publicIPs,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AzureFirewallApplicationRuleCollection
type AzureFirewallApplicationRuleCollection struct {

	//Name: The name of the resource that is unique within the Azure firewall. This
	//name can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the azure firewall application rule collection.
	Properties *AzureFirewallApplicationRuleCollectionPropertiesFormat `json:"properties,omitempty"`
}

//Generated from:
type AzureFirewallApplicationRuleCollectionPropertiesFormat_Status struct {

	//Action: The action type of a rule collection.
	Action *AzureFirewallRCAction_Status `json:"action,omitempty"`

	//Priority: Priority of the application rule collection resource.
	Priority *int `json:"priority,omitempty"`

	//ProvisioningState: The provisioning state of the application rule collection
	//resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//Rules: Collection of rules used by a application rule collection.
	Rules []AzureFirewallApplicationRule_Status `json:"rules,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AzureFirewallIPConfiguration
type AzureFirewallIPConfiguration struct {

	//Name: Name of the resource that is unique within a resource group. This name can
	//be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the azure firewall IP configuration.
	Properties *AzureFirewallIPConfigurationPropertiesFormat `json:"properties,omitempty"`
}

//Generated from:
type AzureFirewallIPConfigurationPropertiesFormat_Status struct {

	//PrivateIPAddress: The Firewall Internal Load Balancer IP to be used as the next
	//hop in User Defined Routes.
	PrivateIPAddress *string `json:"privateIPAddress,omitempty"`

	//ProvisioningState: The provisioning state of the Azure firewall IP configuration
	//resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//PublicIPAddress: Reference to the PublicIP resource. This field is a mandatory
	//input if subnet is not null.
	PublicIPAddress *SubResource_Status `json:"publicIPAddress,omitempty"`

	//Subnet: Reference to the subnet resource. This resource must be named
	//'AzureFirewallSubnet' or 'AzureFirewallManagementSubnet'.
	Subnet *SubResource_Status `json:"subnet,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AzureFirewallNatRuleCollection
type AzureFirewallNatRuleCollection struct {

	//Name: The name of the resource that is unique within the Azure firewall. This
	//name can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the azure firewall NAT rule collection.
	Properties *AzureFirewallNatRuleCollectionProperties `json:"properties,omitempty"`
}

//Generated from:
type AzureFirewallNatRuleCollectionProperties_Status struct {

	//Action: The action type of a NAT rule collection.
	Action *AzureFirewallNatRCAction_Status `json:"action,omitempty"`

	//Priority: Priority of the NAT rule collection resource.
	Priority *int `json:"priority,omitempty"`

	//ProvisioningState: The provisioning state of the NAT rule collection resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//Rules: Collection of rules used by a NAT rule collection.
	Rules []AzureFirewallNatRule_Status `json:"rules,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AzureFirewallNetworkRuleCollection
type AzureFirewallNetworkRuleCollection struct {

	//Name: The name of the resource that is unique within the Azure firewall. This
	//name can be used to access the resource.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the azure firewall network rule collection.
	Properties *AzureFirewallNetworkRuleCollectionPropertiesFormat `json:"properties,omitempty"`
}

//Generated from:
type AzureFirewallNetworkRuleCollectionPropertiesFormat_Status struct {

	//Action: The action type of a rule collection.
	Action *AzureFirewallRCAction_Status `json:"action,omitempty"`

	//Priority: Priority of the network rule collection resource.
	Priority *int `json:"priority,omitempty"`

	//ProvisioningState: The provisioning state of the network rule collection
	//resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//Rules: Collection of rules used by a network rule collection.
	Rules []AzureFirewallNetworkRule_Status `json:"rules,omitempty"`
}

// +kubebuilder:validation:Enum={"Alert","Deny","Off"}
type AzureFirewallPropertiesFormatThreatIntelMode string

const (
	AzureFirewallPropertiesFormatThreatIntelModeAlert = AzureFirewallPropertiesFormatThreatIntelMode("Alert")
	AzureFirewallPropertiesFormatThreatIntelModeDeny  = AzureFirewallPropertiesFormatThreatIntelMode("Deny")
	AzureFirewallPropertiesFormatThreatIntelModeOff   = AzureFirewallPropertiesFormatThreatIntelMode("Off")
)

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AzureFirewallSku
type AzureFirewallSku struct {

	//Name: Name of an Azure Firewall SKU.
	Name *AzureFirewallSkuName `json:"name,omitempty"`

	//Tier: Tier of an Azure Firewall.
	Tier *AzureFirewallSkuTier `json:"tier,omitempty"`
}

// +kubebuilder:validation:Enum={"AZFW_Hub","AZFW_VNet"}
type AzureFirewallSkuStatusName string

const (
	AzureFirewallSkuStatusNameAZFWHub  = AzureFirewallSkuStatusName("AZFW_Hub")
	AzureFirewallSkuStatusNameAZFWVNet = AzureFirewallSkuStatusName("AZFW_VNet")
)

// +kubebuilder:validation:Enum={"Premium","Standard"}
type AzureFirewallSkuStatusTier string

const (
	AzureFirewallSkuStatusTierPremium  = AzureFirewallSkuStatusTier("Premium")
	AzureFirewallSkuStatusTierStandard = AzureFirewallSkuStatusTier("Standard")
)

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/HubIPAddresses
type HubIPAddresses struct {

	//PrivateIPAddress: Private IP Address associated with azure firewall.
	PrivateIPAddress *string `json:"privateIPAddress,omitempty"`

	//PublicIPs: Public IP addresses associated with azure firewall.
	PublicIPs *HubPublicIPAddresses `json:"publicIPs,omitempty"`
}

//Generated from:
type HubPublicIPAddresses_Status struct {

	//Addresses: The number of Public IP addresses associated with azure firewall.
	Addresses []AzureFirewallPublicIPAddress_Status `json:"addresses,omitempty"`

	//Count: Private IP Address associated with azure firewall.
	Count *int `json:"count,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AzureFirewallApplicationRuleCollectionPropertiesFormat
type AzureFirewallApplicationRuleCollectionPropertiesFormat struct {

	//Action: The action type of a rule collection.
	Action *AzureFirewallRCAction `json:"action,omitempty"`

	//Priority: Priority of the application rule collection resource.
	Priority *int `json:"priority,omitempty"`

	//Rules: Collection of rules used by a application rule collection.
	Rules []AzureFirewallApplicationRule `json:"rules,omitempty"`
}

//Generated from:
type AzureFirewallApplicationRule_Status struct {

	//Description: Description of the rule.
	Description *string `json:"description,omitempty"`

	//FqdnTags: List of FQDN Tags for this rule.
	FqdnTags []string `json:"fqdnTags,omitempty"`

	//Name: Name of the application rule.
	Name *string `json:"name,omitempty"`

	//Protocols: Array of ApplicationRuleProtocols.
	Protocols []AzureFirewallApplicationRuleProtocol_Status `json:"protocols,omitempty"`

	//SourceAddresses: List of source IP addresses for this rule.
	SourceAddresses []string `json:"sourceAddresses,omitempty"`

	//SourceIpGroups: List of source IpGroups for this rule.
	SourceIpGroups []string `json:"sourceIpGroups,omitempty"`

	//TargetFqdns: List of FQDNs for this rule.
	TargetFqdns []string `json:"targetFqdns,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AzureFirewallIPConfigurationPropertiesFormat
type AzureFirewallIPConfigurationPropertiesFormat struct {

	//PublicIPAddress: Reference to the PublicIP resource. This field is a mandatory
	//input if subnet is not null.
	PublicIPAddress *SubResource `json:"publicIPAddress,omitempty"`

	//Subnet: Reference to the subnet resource. This resource must be named
	//'AzureFirewallSubnet' or 'AzureFirewallManagementSubnet'.
	Subnet *SubResource `json:"subnet,omitempty"`
}

//Generated from:
type AzureFirewallNatRCAction_Status struct {

	//Type: The type of action.
	Type *AzureFirewallNatRCActionType_Status `json:"type,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AzureFirewallNatRuleCollectionProperties
type AzureFirewallNatRuleCollectionProperties struct {

	//Action: The action type of a NAT rule collection.
	Action *AzureFirewallNatRCAction `json:"action,omitempty"`

	//Priority: Priority of the NAT rule collection resource.
	Priority *int `json:"priority,omitempty"`

	//Rules: Collection of rules used by a NAT rule collection.
	Rules []AzureFirewallNatRule `json:"rules,omitempty"`
}

//Generated from:
type AzureFirewallNatRule_Status struct {

	//Description: Description of the rule.
	Description *string `json:"description,omitempty"`

	//DestinationAddresses: List of destination IP addresses for this rule. Supports
	//IP ranges, prefixes, and service tags.
	DestinationAddresses []string `json:"destinationAddresses,omitempty"`

	//DestinationPorts: List of destination ports.
	DestinationPorts []string `json:"destinationPorts,omitempty"`

	//Name: Name of the NAT rule.
	Name *string `json:"name,omitempty"`

	//Protocols: Array of AzureFirewallNetworkRuleProtocols applicable to this NAT
	//rule.
	Protocols []AzureFirewallNetworkRuleProtocol_Status `json:"protocols,omitempty"`

	//SourceAddresses: List of source IP addresses for this rule.
	SourceAddresses []string `json:"sourceAddresses,omitempty"`

	//SourceIpGroups: List of source IpGroups for this rule.
	SourceIpGroups []string `json:"sourceIpGroups,omitempty"`

	//TranslatedAddress: The translated address for this NAT rule.
	TranslatedAddress *string `json:"translatedAddress,omitempty"`

	//TranslatedFqdn: The translated FQDN for this NAT rule.
	TranslatedFqdn *string `json:"translatedFqdn,omitempty"`

	//TranslatedPort: The translated port for this NAT rule.
	TranslatedPort *string `json:"translatedPort,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AzureFirewallNetworkRuleCollectionPropertiesFormat
type AzureFirewallNetworkRuleCollectionPropertiesFormat struct {

	//Action: The action type of a rule collection.
	Action *AzureFirewallRCAction `json:"action,omitempty"`

	//Priority: Priority of the network rule collection resource.
	Priority *int `json:"priority,omitempty"`

	//Rules: Collection of rules used by a network rule collection.
	Rules []AzureFirewallNetworkRule `json:"rules,omitempty"`
}

//Generated from:
type AzureFirewallNetworkRule_Status struct {

	//Description: Description of the rule.
	Description *string `json:"description,omitempty"`

	//DestinationAddresses: List of destination IP addresses.
	DestinationAddresses []string `json:"destinationAddresses,omitempty"`

	//DestinationFqdns: List of destination FQDNs.
	DestinationFqdns []string `json:"destinationFqdns,omitempty"`

	//DestinationIpGroups: List of destination IpGroups for this rule.
	DestinationIpGroups []string `json:"destinationIpGroups,omitempty"`

	//DestinationPorts: List of destination ports.
	DestinationPorts []string `json:"destinationPorts,omitempty"`

	//Name: Name of the network rule.
	Name *string `json:"name,omitempty"`

	//Protocols: Array of AzureFirewallNetworkRuleProtocols.
	Protocols []AzureFirewallNetworkRuleProtocol_Status `json:"protocols,omitempty"`

	//SourceAddresses: List of source IP addresses for this rule.
	SourceAddresses []string `json:"sourceAddresses,omitempty"`

	//SourceIpGroups: List of source IpGroups for this rule.
	SourceIpGroups []string `json:"sourceIpGroups,omitempty"`
}

//Generated from:
type AzureFirewallPublicIPAddress_Status struct {

	//Address: Public IP Address value.
	Address *string `json:"address,omitempty"`
}

//Generated from:
type AzureFirewallRCAction_Status struct {

	//Type: The type of action.
	Type *AzureFirewallRCActionType_Status `json:"type,omitempty"`
}

// +kubebuilder:validation:Enum={"AZFW_Hub","AZFW_VNet"}
type AzureFirewallSkuName string

const (
	AzureFirewallSkuNameAZFWHub  = AzureFirewallSkuName("AZFW_Hub")
	AzureFirewallSkuNameAZFWVNet = AzureFirewallSkuName("AZFW_VNet")
)

// +kubebuilder:validation:Enum={"Premium","Standard"}
type AzureFirewallSkuTier string

const (
	AzureFirewallSkuTierPremium  = AzureFirewallSkuTier("Premium")
	AzureFirewallSkuTierStandard = AzureFirewallSkuTier("Standard")
)

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/HubPublicIPAddresses
type HubPublicIPAddresses struct {

	//Addresses: The number of Public IP addresses associated with azure firewall.
	Addresses []AzureFirewallPublicIPAddress `json:"addresses,omitempty"`

	//Count: Private IP Address associated with azure firewall.
	Count *int `json:"count,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AzureFirewallApplicationRule
type AzureFirewallApplicationRule struct {

	//Description: Description of the rule.
	Description *string `json:"description,omitempty"`

	//FqdnTags: List of FQDN Tags for this rule.
	FqdnTags []string `json:"fqdnTags,omitempty"`

	//Name: Name of the application rule.
	Name *string `json:"name,omitempty"`

	//Protocols: Array of ApplicationRuleProtocols.
	Protocols []AzureFirewallApplicationRuleProtocol `json:"protocols,omitempty"`

	//SourceAddresses: List of source IP addresses for this rule.
	SourceAddresses []string `json:"sourceAddresses,omitempty"`

	//SourceIpGroups: List of source IpGroups for this rule.
	SourceIpGroups []string `json:"sourceIpGroups,omitempty"`

	//TargetFqdns: List of FQDNs for this rule.
	TargetFqdns []string `json:"targetFqdns,omitempty"`
}

//Generated from:
type AzureFirewallApplicationRuleProtocol_Status struct {

	//Port: Port number for the protocol, cannot be greater than 64000. This field is
	//optional.
	Port *int `json:"port,omitempty"`

	//ProtocolType: Protocol type.
	ProtocolType *AzureFirewallApplicationRuleProtocolType_Status `json:"protocolType,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AzureFirewallNatRCAction
type AzureFirewallNatRCAction struct {

	//Type: The type of action.
	Type *AzureFirewallNatRCActionType `json:"type,omitempty"`
}

//Generated from:
// +kubebuilder:validation:Enum={"Dnat","Snat"}
type AzureFirewallNatRCActionType_Status string

const (
	AzureFirewallNatRCActionType_StatusDnat = AzureFirewallNatRCActionType_Status("Dnat")
	AzureFirewallNatRCActionType_StatusSnat = AzureFirewallNatRCActionType_Status("Snat")
)

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AzureFirewallNatRule
type AzureFirewallNatRule struct {

	//Description: Description of the rule.
	Description *string `json:"description,omitempty"`

	//DestinationAddresses: List of destination IP addresses for this rule. Supports
	//IP ranges, prefixes, and service tags.
	DestinationAddresses []string `json:"destinationAddresses,omitempty"`

	//DestinationPorts: List of destination ports.
	DestinationPorts []string `json:"destinationPorts,omitempty"`

	//Name: Name of the NAT rule.
	Name *string `json:"name,omitempty"`

	//Protocols: Array of AzureFirewallNetworkRuleProtocols applicable to this NAT
	//rule.
	Protocols []AzureFirewallNatRuleProtocols `json:"protocols,omitempty"`

	//SourceAddresses: List of source IP addresses for this rule.
	SourceAddresses []string `json:"sourceAddresses,omitempty"`

	//SourceIpGroups: List of source IpGroups for this rule.
	SourceIpGroups []string `json:"sourceIpGroups,omitempty"`

	//TranslatedAddress: The translated address for this NAT rule.
	TranslatedAddress *string `json:"translatedAddress,omitempty"`

	//TranslatedFqdn: The translated FQDN for this NAT rule.
	TranslatedFqdn *string `json:"translatedFqdn,omitempty"`

	//TranslatedPort: The translated port for this NAT rule.
	TranslatedPort *string `json:"translatedPort,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AzureFirewallNetworkRule
type AzureFirewallNetworkRule struct {

	//Description: Description of the rule.
	Description *string `json:"description,omitempty"`

	//DestinationAddresses: List of destination IP addresses.
	DestinationAddresses []string `json:"destinationAddresses,omitempty"`

	//DestinationFqdns: List of destination FQDNs.
	DestinationFqdns []string `json:"destinationFqdns,omitempty"`

	//DestinationIpGroups: List of destination IpGroups for this rule.
	DestinationIpGroups []string `json:"destinationIpGroups,omitempty"`

	//DestinationPorts: List of destination ports.
	DestinationPorts []string `json:"destinationPorts,omitempty"`

	//Name: Name of the network rule.
	Name *string `json:"name,omitempty"`

	//Protocols: Array of AzureFirewallNetworkRuleProtocols.
	Protocols []AzureFirewallNetworkRuleProtocols `json:"protocols,omitempty"`

	//SourceAddresses: List of source IP addresses for this rule.
	SourceAddresses []string `json:"sourceAddresses,omitempty"`

	//SourceIpGroups: List of source IpGroups for this rule.
	SourceIpGroups []string `json:"sourceIpGroups,omitempty"`
}

//Generated from:
// +kubebuilder:validation:Enum={"Any","ICMP","TCP","UDP"}
type AzureFirewallNetworkRuleProtocol_Status string

const (
	AzureFirewallNetworkRuleProtocol_StatusAny  = AzureFirewallNetworkRuleProtocol_Status("Any")
	AzureFirewallNetworkRuleProtocol_StatusICMP = AzureFirewallNetworkRuleProtocol_Status("ICMP")
	AzureFirewallNetworkRuleProtocol_StatusTCP  = AzureFirewallNetworkRuleProtocol_Status("TCP")
	AzureFirewallNetworkRuleProtocol_StatusUDP  = AzureFirewallNetworkRuleProtocol_Status("UDP")
)

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AzureFirewallPublicIPAddress
type AzureFirewallPublicIPAddress struct {

	//Address: Public IP Address value.
	Address *string `json:"address,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AzureFirewallRCAction
type AzureFirewallRCAction struct {

	//Type: The type of action.
	Type *AzureFirewallRCActionType `json:"type,omitempty"`
}

//Generated from:
// +kubebuilder:validation:Enum={"Allow","Deny"}
type AzureFirewallRCActionType_Status string

const (
	AzureFirewallRCActionType_StatusAllow = AzureFirewallRCActionType_Status("Allow")
	AzureFirewallRCActionType_StatusDeny  = AzureFirewallRCActionType_Status("Deny")
)

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AzureFirewallApplicationRuleProtocol
type AzureFirewallApplicationRuleProtocol struct {

	//Port: Port number for the protocol, cannot be greater than 64000. This field is
	//optional.
	Port *int `json:"port,omitempty"`

	//ProtocolType: Protocol type.
	ProtocolType *AzureFirewallApplicationRuleProtocolProtocolType `json:"protocolType,omitempty"`
}

//Generated from:
// +kubebuilder:validation:Enum={"Http","Https","Mssql"}
type AzureFirewallApplicationRuleProtocolType_Status string

const (
	AzureFirewallApplicationRuleProtocolType_StatusHttp  = AzureFirewallApplicationRuleProtocolType_Status("Http")
	AzureFirewallApplicationRuleProtocolType_StatusHttps = AzureFirewallApplicationRuleProtocolType_Status("Https")
	AzureFirewallApplicationRuleProtocolType_StatusMssql = AzureFirewallApplicationRuleProtocolType_Status("Mssql")
)

// +kubebuilder:validation:Enum={"Dnat","Snat"}
type AzureFirewallNatRCActionType string

const (
	AzureFirewallNatRCActionTypeDnat = AzureFirewallNatRCActionType("Dnat")
	AzureFirewallNatRCActionTypeSnat = AzureFirewallNatRCActionType("Snat")
)

// +kubebuilder:validation:Enum={"Any","ICMP","TCP","UDP"}
type AzureFirewallNatRuleProtocols string

const (
	AzureFirewallNatRuleProtocolsAny  = AzureFirewallNatRuleProtocols("Any")
	AzureFirewallNatRuleProtocolsICMP = AzureFirewallNatRuleProtocols("ICMP")
	AzureFirewallNatRuleProtocolsTCP  = AzureFirewallNatRuleProtocols("TCP")
	AzureFirewallNatRuleProtocolsUDP  = AzureFirewallNatRuleProtocols("UDP")
)

// +kubebuilder:validation:Enum={"Any","ICMP","TCP","UDP"}
type AzureFirewallNetworkRuleProtocols string

const (
	AzureFirewallNetworkRuleProtocolsAny  = AzureFirewallNetworkRuleProtocols("Any")
	AzureFirewallNetworkRuleProtocolsICMP = AzureFirewallNetworkRuleProtocols("ICMP")
	AzureFirewallNetworkRuleProtocolsTCP  = AzureFirewallNetworkRuleProtocols("TCP")
	AzureFirewallNetworkRuleProtocolsUDP  = AzureFirewallNetworkRuleProtocols("UDP")
)

// +kubebuilder:validation:Enum={"Allow","Deny"}
type AzureFirewallRCActionType string

const (
	AzureFirewallRCActionTypeAllow = AzureFirewallRCActionType("Allow")
	AzureFirewallRCActionTypeDeny  = AzureFirewallRCActionType("Deny")
)

// +kubebuilder:validation:Enum={"Http","Https","Mssql"}
type AzureFirewallApplicationRuleProtocolProtocolType string

const (
	AzureFirewallApplicationRuleProtocolProtocolTypeHttp  = AzureFirewallApplicationRuleProtocolProtocolType("Http")
	AzureFirewallApplicationRuleProtocolProtocolTypeHttps = AzureFirewallApplicationRuleProtocolProtocolType("Https")
	AzureFirewallApplicationRuleProtocolProtocolTypeMssql = AzureFirewallApplicationRuleProtocolProtocolType("Mssql")
)

func init() {
	SchemeBuilder.Register(&AzureFirewalls{}, &AzureFirewallsList{})
}
