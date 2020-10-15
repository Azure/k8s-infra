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
type ApplicationGatewayWebApplicationFirewallPolicies struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ApplicationGatewayWebApplicationFirewallPolicies_Spec `json:"spec,omitempty"`
	Status            WebApplicationFirewallPolicy_Status                   `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ApplicationGatewayWebApplicationFirewallPoliciesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationGatewayWebApplicationFirewallPolicies `json:"items"`
}

type ApplicationGatewayWebApplicationFirewallPolicies_Spec struct {
	ForProvider ApplicationGatewayWebApplicationFirewallPoliciesParameters `json:"forProvider"`
}

//Generated from:
type WebApplicationFirewallPolicy_Status struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Location: Resource location.
	Location *string `json:"location,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the web application firewall policy.
	Properties *WebApplicationFirewallPolicyPropertiesFormat_Status `json:"properties,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type ApplicationGatewayWebApplicationFirewallPoliciesParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion ApplicationGatewayWebApplicationFirewallPoliciesSpecApiVersion `json:"apiVersion"`
	Comments   *string                                                        `json:"comments,omitempty"`

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
	//Properties: Properties of the web application firewall policy.
	Properties WebApplicationFirewallPolicyPropertiesFormat `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type ApplicationGatewayWebApplicationFirewallPoliciesSpecType `json:"type"`
}

//Generated from:
type WebApplicationFirewallPolicyPropertiesFormat_Status struct {

	//ApplicationGateways: A collection of references to application gateways.
	ApplicationGateways []ApplicationGateway_Status `json:"applicationGateways,omitempty"`

	//CustomRules: The custom rules inside the policy.
	CustomRules []WebApplicationFirewallCustomRule_Status `json:"customRules,omitempty"`

	//HttpListeners: A collection of references to application gateway http listeners.
	HttpListeners []SubResource_Status `json:"httpListeners,omitempty"`

	// +kubebuilder:validation:Required
	//ManagedRules: Describes the managedRules structure.
	ManagedRules ManagedRulesDefinition_Status `json:"managedRules"`

	//PathBasedRules: A collection of references to application gateway path rules.
	PathBasedRules []SubResource_Status `json:"pathBasedRules,omitempty"`

	//PolicySettings: The PolicySettings for policy.
	PolicySettings *PolicySettings_Status `json:"policySettings,omitempty"`

	//ProvisioningState: The provisioning state of the web application firewall policy
	//resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//ResourceState: Resource status of the policy.
	ResourceState *WebApplicationFirewallPolicyPropertiesFormatStatusResourceState `json:"resourceState,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type ApplicationGatewayWebApplicationFirewallPoliciesSpecApiVersion string

const ApplicationGatewayWebApplicationFirewallPoliciesSpecApiVersion20200501 = ApplicationGatewayWebApplicationFirewallPoliciesSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies"}
type ApplicationGatewayWebApplicationFirewallPoliciesSpecType string

const ApplicationGatewayWebApplicationFirewallPoliciesSpecTypeMicrosoftNetworkApplicationGatewayWebApplicationFirewallPolicies = ApplicationGatewayWebApplicationFirewallPoliciesSpecType("Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies")

//Generated from:
type ManagedRulesDefinition_Status struct {

	//Exclusions: The Exclusions that are applied on the policy.
	Exclusions []OwaspCrsExclusionEntry_Status `json:"exclusions,omitempty"`

	// +kubebuilder:validation:Required
	//ManagedRuleSets: The managed rule sets that are associated with the policy.
	ManagedRuleSets []ManagedRuleSet_Status `json:"managedRuleSets"`
}

//Generated from:
type PolicySettings_Status struct {

	//FileUploadLimitInMb: Maximum file upload size in Mb for WAF.
	FileUploadLimitInMb *int `json:"fileUploadLimitInMb,omitempty"`

	//MaxRequestBodySizeInKb: Maximum request body size in Kb for WAF.
	MaxRequestBodySizeInKb *int `json:"maxRequestBodySizeInKb,omitempty"`

	//Mode: The mode of the policy.
	Mode *PolicySettingsStatusMode `json:"mode,omitempty"`

	//RequestBodyCheck: Whether to allow WAF to check request Body.
	RequestBodyCheck *bool `json:"requestBodyCheck,omitempty"`

	//State: The state of the policy.
	State *PolicySettingsStatusState `json:"state,omitempty"`
}

//Generated from:
// +kubebuilder:validation:Enum={"Deleting","Failed","Succeeded","Updating"}
type ProvisioningState_Status string

const (
	ProvisioningState_StatusDeleting  = ProvisioningState_Status("Deleting")
	ProvisioningState_StatusFailed    = ProvisioningState_Status("Failed")
	ProvisioningState_StatusSucceeded = ProvisioningState_Status("Succeeded")
	ProvisioningState_StatusUpdating  = ProvisioningState_Status("Updating")
)

//Generated from:
type SubResource_Status struct {

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`
}

//Generated from:
type WebApplicationFirewallCustomRule_Status struct {

	// +kubebuilder:validation:Required
	//Action: Type of Actions.
	Action WebApplicationFirewallCustomRuleStatusAction `json:"action"`

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	// +kubebuilder:validation:Required
	//MatchConditions: List of match conditions.
	MatchConditions []MatchCondition_Status `json:"matchConditions"`

	//Name: The name of the resource that is unique within a policy. This name can be
	//used to access the resource.
	Name *string `json:"name,omitempty"`

	// +kubebuilder:validation:Required
	//Priority: Priority of the rule. Rules with a lower value will be evaluated
	//before rules with a higher value.
	Priority int `json:"priority"`

	// +kubebuilder:validation:Required
	//RuleType: The rule type.
	RuleType WebApplicationFirewallCustomRuleStatusRuleType `json:"ruleType"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/WebApplicationFirewallPolicyPropertiesFormat
type WebApplicationFirewallPolicyPropertiesFormat struct {

	//CustomRules: The custom rules inside the policy.
	CustomRules []WebApplicationFirewallCustomRule `json:"customRules,omitempty"`

	// +kubebuilder:validation:Required
	//ManagedRules: Describes the managedRules structure.
	ManagedRules ManagedRulesDefinition `json:"managedRules"`

	//PolicySettings: The PolicySettings for policy.
	PolicySettings *PolicySettings `json:"policySettings,omitempty"`
}

// +kubebuilder:validation:Enum={"Creating","Deleting","Disabled","Disabling","Enabled","Enabling"}
type WebApplicationFirewallPolicyPropertiesFormatStatusResourceState string

const (
	WebApplicationFirewallPolicyPropertiesFormatStatusResourceStateCreating  = WebApplicationFirewallPolicyPropertiesFormatStatusResourceState("Creating")
	WebApplicationFirewallPolicyPropertiesFormatStatusResourceStateDeleting  = WebApplicationFirewallPolicyPropertiesFormatStatusResourceState("Deleting")
	WebApplicationFirewallPolicyPropertiesFormatStatusResourceStateDisabled  = WebApplicationFirewallPolicyPropertiesFormatStatusResourceState("Disabled")
	WebApplicationFirewallPolicyPropertiesFormatStatusResourceStateDisabling = WebApplicationFirewallPolicyPropertiesFormatStatusResourceState("Disabling")
	WebApplicationFirewallPolicyPropertiesFormatStatusResourceStateEnabled   = WebApplicationFirewallPolicyPropertiesFormatStatusResourceState("Enabled")
	WebApplicationFirewallPolicyPropertiesFormatStatusResourceStateEnabling  = WebApplicationFirewallPolicyPropertiesFormatStatusResourceState("Enabling")
)

//Generated from:
type ManagedRuleSet_Status struct {

	//RuleGroupOverrides: Defines the rule group overrides to apply to the rule set.
	RuleGroupOverrides []ManagedRuleGroupOverride_Status `json:"ruleGroupOverrides,omitempty"`

	// +kubebuilder:validation:Required
	//RuleSetType: Defines the rule set type to use.
	RuleSetType string `json:"ruleSetType"`

	// +kubebuilder:validation:Required
	//RuleSetVersion: Defines the version of the rule set to use.
	RuleSetVersion string `json:"ruleSetVersion"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/ManagedRulesDefinition
type ManagedRulesDefinition struct {

	//Exclusions: The Exclusions that are applied on the policy.
	Exclusions []OwaspCrsExclusionEntry `json:"exclusions,omitempty"`

	// +kubebuilder:validation:Required
	//ManagedRuleSets: The managed rule sets that are associated with the policy.
	ManagedRuleSets []ManagedRuleSet `json:"managedRuleSets"`
}

//Generated from:
type MatchCondition_Status struct {

	// +kubebuilder:validation:Required
	//MatchValues: Match value.
	MatchValues []string `json:"matchValues"`

	// +kubebuilder:validation:Required
	//MatchVariables: List of match variables.
	MatchVariables []MatchVariable_Status `json:"matchVariables"`

	//NegationConditon: Whether this is negate condition or not.
	NegationConditon *bool `json:"negationConditon,omitempty"`

	// +kubebuilder:validation:Required
	//Operator: The operator to be matched.
	Operator MatchConditionStatusOperator `json:"operator"`

	//Transforms: List of transforms.
	Transforms []Transform_Status `json:"transforms,omitempty"`
}

//Generated from:
type OwaspCrsExclusionEntry_Status struct {

	// +kubebuilder:validation:Required
	//MatchVariable: The variable to be excluded.
	MatchVariable OwaspCrsExclusionEntryStatusMatchVariable `json:"matchVariable"`

	// +kubebuilder:validation:Required
	//Selector: When matchVariable is a collection, operator used to specify which
	//elements in the collection this exclusion applies to.
	Selector string `json:"selector"`

	// +kubebuilder:validation:Required
	//SelectorMatchOperator: When matchVariable is a collection, operate on the
	//selector to specify which elements in the collection this exclusion applies to.
	SelectorMatchOperator OwaspCrsExclusionEntryStatusSelectorMatchOperator `json:"selectorMatchOperator"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/PolicySettings
type PolicySettings struct {

	//FileUploadLimitInMb: Maximum file upload size in Mb for WAF.
	FileUploadLimitInMb *int `json:"fileUploadLimitInMb,omitempty"`

	//MaxRequestBodySizeInKb: Maximum request body size in Kb for WAF.
	MaxRequestBodySizeInKb *int `json:"maxRequestBodySizeInKb,omitempty"`

	//Mode: The mode of the policy.
	Mode *PolicySettingsMode `json:"mode,omitempty"`

	//RequestBodyCheck: Whether to allow WAF to check request Body.
	RequestBodyCheck *bool `json:"requestBodyCheck,omitempty"`

	//State: The state of the policy.
	State *PolicySettingsState `json:"state,omitempty"`
}

// +kubebuilder:validation:Enum={"Detection","Prevention"}
type PolicySettingsStatusMode string

const (
	PolicySettingsStatusModeDetection  = PolicySettingsStatusMode("Detection")
	PolicySettingsStatusModePrevention = PolicySettingsStatusMode("Prevention")
)

// +kubebuilder:validation:Enum={"Disabled","Enabled"}
type PolicySettingsStatusState string

const (
	PolicySettingsStatusStateDisabled = PolicySettingsStatusState("Disabled")
	PolicySettingsStatusStateEnabled  = PolicySettingsStatusState("Enabled")
)

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/WebApplicationFirewallCustomRule
type WebApplicationFirewallCustomRule struct {

	// +kubebuilder:validation:Required
	//Action: Type of Actions.
	Action WebApplicationFirewallCustomRuleAction `json:"action"`

	// +kubebuilder:validation:Required
	//MatchConditions: List of match conditions.
	MatchConditions []MatchCondition `json:"matchConditions"`

	//Name: The name of the resource that is unique within a policy. This name can be
	//used to access the resource.
	Name *string `json:"name,omitempty"`

	// +kubebuilder:validation:Required
	//Priority: Priority of the rule. Rules with a lower value will be evaluated
	//before rules with a higher value.
	Priority int `json:"priority"`

	// +kubebuilder:validation:Required
	//RuleType: The rule type.
	RuleType WebApplicationFirewallCustomRuleRuleType `json:"ruleType"`
}

// +kubebuilder:validation:Enum={"Allow","Block","Log"}
type WebApplicationFirewallCustomRuleStatusAction string

const (
	WebApplicationFirewallCustomRuleStatusActionAllow = WebApplicationFirewallCustomRuleStatusAction("Allow")
	WebApplicationFirewallCustomRuleStatusActionBlock = WebApplicationFirewallCustomRuleStatusAction("Block")
	WebApplicationFirewallCustomRuleStatusActionLog   = WebApplicationFirewallCustomRuleStatusAction("Log")
)

// +kubebuilder:validation:Enum={"Invalid","MatchRule"}
type WebApplicationFirewallCustomRuleStatusRuleType string

const (
	WebApplicationFirewallCustomRuleStatusRuleTypeInvalid   = WebApplicationFirewallCustomRuleStatusRuleType("Invalid")
	WebApplicationFirewallCustomRuleStatusRuleTypeMatchRule = WebApplicationFirewallCustomRuleStatusRuleType("MatchRule")
)

//Generated from:
type ManagedRuleGroupOverride_Status struct {

	// +kubebuilder:validation:Required
	//RuleGroupName: The managed rule group to override.
	RuleGroupName string `json:"ruleGroupName"`

	//Rules: List of rules that will be disabled. If none specified, all rules in the
	//group will be disabled.
	Rules []ManagedRuleOverride_Status `json:"rules,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/ManagedRuleSet
type ManagedRuleSet struct {

	//RuleGroupOverrides: Defines the rule group overrides to apply to the rule set.
	RuleGroupOverrides []ManagedRuleGroupOverride `json:"ruleGroupOverrides,omitempty"`

	// +kubebuilder:validation:Required
	//RuleSetType: Defines the rule set type to use.
	RuleSetType string `json:"ruleSetType"`

	// +kubebuilder:validation:Required
	//RuleSetVersion: Defines the version of the rule set to use.
	RuleSetVersion string `json:"ruleSetVersion"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/MatchCondition
type MatchCondition struct {

	// +kubebuilder:validation:Required
	//MatchValues: Match value.
	MatchValues []string `json:"matchValues"`

	// +kubebuilder:validation:Required
	//MatchVariables: List of match variables.
	MatchVariables []MatchVariable `json:"matchVariables"`

	//NegationConditon: Whether this is negate condition or not.
	NegationConditon *bool `json:"negationConditon,omitempty"`

	// +kubebuilder:validation:Required
	//Operator: The operator to be matched.
	Operator MatchConditionOperator `json:"operator"`

	//Transforms: List of transforms.
	Transforms []MatchConditionTransforms `json:"transforms,omitempty"`
}

// +kubebuilder:validation:Enum={"BeginsWith","Contains","EndsWith","Equal","GeoMatch","GreaterThan","GreaterThanOrEqual","IPMatch","LessThan","LessThanOrEqual","Regex"}
type MatchConditionStatusOperator string

const (
	MatchConditionStatusOperatorBeginsWith         = MatchConditionStatusOperator("BeginsWith")
	MatchConditionStatusOperatorContains           = MatchConditionStatusOperator("Contains")
	MatchConditionStatusOperatorEndsWith           = MatchConditionStatusOperator("EndsWith")
	MatchConditionStatusOperatorEqual              = MatchConditionStatusOperator("Equal")
	MatchConditionStatusOperatorGeoMatch           = MatchConditionStatusOperator("GeoMatch")
	MatchConditionStatusOperatorGreaterThan        = MatchConditionStatusOperator("GreaterThan")
	MatchConditionStatusOperatorGreaterThanOrEqual = MatchConditionStatusOperator("GreaterThanOrEqual")
	MatchConditionStatusOperatorIPMatch            = MatchConditionStatusOperator("IPMatch")
	MatchConditionStatusOperatorLessThan           = MatchConditionStatusOperator("LessThan")
	MatchConditionStatusOperatorLessThanOrEqual    = MatchConditionStatusOperator("LessThanOrEqual")
	MatchConditionStatusOperatorRegex              = MatchConditionStatusOperator("Regex")
)

//Generated from:
type MatchVariable_Status struct {

	//Selector: The selector of match variable.
	Selector *string `json:"selector,omitempty"`

	// +kubebuilder:validation:Required
	//VariableName: Match Variable.
	VariableName MatchVariableStatusVariableName `json:"variableName"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/OwaspCrsExclusionEntry
type OwaspCrsExclusionEntry struct {

	// +kubebuilder:validation:Required
	//MatchVariable: The variable to be excluded.
	MatchVariable OwaspCrsExclusionEntryMatchVariable `json:"matchVariable"`

	// +kubebuilder:validation:Required
	//Selector: When matchVariable is a collection, operator used to specify which
	//elements in the collection this exclusion applies to.
	Selector string `json:"selector"`

	// +kubebuilder:validation:Required
	//SelectorMatchOperator: When matchVariable is a collection, operate on the
	//selector to specify which elements in the collection this exclusion applies to.
	SelectorMatchOperator OwaspCrsExclusionEntrySelectorMatchOperator `json:"selectorMatchOperator"`
}

// +kubebuilder:validation:Enum={"RequestArgNames","RequestCookieNames","RequestHeaderNames"}
type OwaspCrsExclusionEntryStatusMatchVariable string

const (
	OwaspCrsExclusionEntryStatusMatchVariableRequestArgNames    = OwaspCrsExclusionEntryStatusMatchVariable("RequestArgNames")
	OwaspCrsExclusionEntryStatusMatchVariableRequestCookieNames = OwaspCrsExclusionEntryStatusMatchVariable("RequestCookieNames")
	OwaspCrsExclusionEntryStatusMatchVariableRequestHeaderNames = OwaspCrsExclusionEntryStatusMatchVariable("RequestHeaderNames")
)

// +kubebuilder:validation:Enum={"Contains","EndsWith","Equals","EqualsAny","StartsWith"}
type OwaspCrsExclusionEntryStatusSelectorMatchOperator string

const (
	OwaspCrsExclusionEntryStatusSelectorMatchOperatorContains   = OwaspCrsExclusionEntryStatusSelectorMatchOperator("Contains")
	OwaspCrsExclusionEntryStatusSelectorMatchOperatorEndsWith   = OwaspCrsExclusionEntryStatusSelectorMatchOperator("EndsWith")
	OwaspCrsExclusionEntryStatusSelectorMatchOperatorEquals     = OwaspCrsExclusionEntryStatusSelectorMatchOperator("Equals")
	OwaspCrsExclusionEntryStatusSelectorMatchOperatorEqualsAny  = OwaspCrsExclusionEntryStatusSelectorMatchOperator("EqualsAny")
	OwaspCrsExclusionEntryStatusSelectorMatchOperatorStartsWith = OwaspCrsExclusionEntryStatusSelectorMatchOperator("StartsWith")
)

// +kubebuilder:validation:Enum={"Detection","Prevention"}
type PolicySettingsMode string

const (
	PolicySettingsModeDetection  = PolicySettingsMode("Detection")
	PolicySettingsModePrevention = PolicySettingsMode("Prevention")
)

// +kubebuilder:validation:Enum={"Disabled","Enabled"}
type PolicySettingsState string

const (
	PolicySettingsStateDisabled = PolicySettingsState("Disabled")
	PolicySettingsStateEnabled  = PolicySettingsState("Enabled")
)

//Generated from:
// +kubebuilder:validation:Enum={"HtmlEntityDecode","Lowercase","RemoveNulls","Trim","UrlDecode","UrlEncode"}
type Transform_Status string

const (
	Transform_StatusHtmlEntityDecode = Transform_Status("HtmlEntityDecode")
	Transform_StatusLowercase        = Transform_Status("Lowercase")
	Transform_StatusRemoveNulls      = Transform_Status("RemoveNulls")
	Transform_StatusTrim             = Transform_Status("Trim")
	Transform_StatusUrlDecode        = Transform_Status("UrlDecode")
	Transform_StatusUrlEncode        = Transform_Status("UrlEncode")
)

// +kubebuilder:validation:Enum={"Allow","Block","Log"}
type WebApplicationFirewallCustomRuleAction string

const (
	WebApplicationFirewallCustomRuleActionAllow = WebApplicationFirewallCustomRuleAction("Allow")
	WebApplicationFirewallCustomRuleActionBlock = WebApplicationFirewallCustomRuleAction("Block")
	WebApplicationFirewallCustomRuleActionLog   = WebApplicationFirewallCustomRuleAction("Log")
)

// +kubebuilder:validation:Enum={"Invalid","MatchRule"}
type WebApplicationFirewallCustomRuleRuleType string

const (
	WebApplicationFirewallCustomRuleRuleTypeInvalid   = WebApplicationFirewallCustomRuleRuleType("Invalid")
	WebApplicationFirewallCustomRuleRuleTypeMatchRule = WebApplicationFirewallCustomRuleRuleType("MatchRule")
)

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/ManagedRuleGroupOverride
type ManagedRuleGroupOverride struct {

	// +kubebuilder:validation:Required
	//RuleGroupName: The managed rule group to override.
	RuleGroupName string `json:"ruleGroupName"`

	//Rules: List of rules that will be disabled. If none specified, all rules in the
	//group will be disabled.
	Rules []ManagedRuleOverride `json:"rules,omitempty"`
}

//Generated from:
type ManagedRuleOverride_Status struct {

	// +kubebuilder:validation:Required
	//RuleId: Identifier for the managed rule.
	RuleId string `json:"ruleId"`

	//State: The state of the managed rule. Defaults to Disabled if not specified.
	State *ManagedRuleOverrideStatusState `json:"state,omitempty"`
}

// +kubebuilder:validation:Enum={"BeginsWith","Contains","EndsWith","Equal","GeoMatch","GreaterThan","GreaterThanOrEqual","IPMatch","LessThan","LessThanOrEqual","Regex"}
type MatchConditionOperator string

const (
	MatchConditionOperatorBeginsWith         = MatchConditionOperator("BeginsWith")
	MatchConditionOperatorContains           = MatchConditionOperator("Contains")
	MatchConditionOperatorEndsWith           = MatchConditionOperator("EndsWith")
	MatchConditionOperatorEqual              = MatchConditionOperator("Equal")
	MatchConditionOperatorGeoMatch           = MatchConditionOperator("GeoMatch")
	MatchConditionOperatorGreaterThan        = MatchConditionOperator("GreaterThan")
	MatchConditionOperatorGreaterThanOrEqual = MatchConditionOperator("GreaterThanOrEqual")
	MatchConditionOperatorIPMatch            = MatchConditionOperator("IPMatch")
	MatchConditionOperatorLessThan           = MatchConditionOperator("LessThan")
	MatchConditionOperatorLessThanOrEqual    = MatchConditionOperator("LessThanOrEqual")
	MatchConditionOperatorRegex              = MatchConditionOperator("Regex")
)

// +kubebuilder:validation:Enum={"HtmlEntityDecode","Lowercase","RemoveNulls","Trim","UrlDecode","UrlEncode"}
type MatchConditionTransforms string

const (
	MatchConditionTransformsHtmlEntityDecode = MatchConditionTransforms("HtmlEntityDecode")
	MatchConditionTransformsLowercase        = MatchConditionTransforms("Lowercase")
	MatchConditionTransformsRemoveNulls      = MatchConditionTransforms("RemoveNulls")
	MatchConditionTransformsTrim             = MatchConditionTransforms("Trim")
	MatchConditionTransformsUrlDecode        = MatchConditionTransforms("UrlDecode")
	MatchConditionTransformsUrlEncode        = MatchConditionTransforms("UrlEncode")
)

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/MatchVariable
type MatchVariable struct {

	//Selector: The selector of match variable.
	Selector *string `json:"selector,omitempty"`

	// +kubebuilder:validation:Required
	//VariableName: Match Variable.
	VariableName MatchVariableVariableName `json:"variableName"`
}

// +kubebuilder:validation:Enum={"PostArgs","QueryString","RemoteAddr","RequestBody","RequestCookies","RequestHeaders","RequestMethod","RequestUri"}
type MatchVariableStatusVariableName string

const (
	MatchVariableStatusVariableNamePostArgs       = MatchVariableStatusVariableName("PostArgs")
	MatchVariableStatusVariableNameQueryString    = MatchVariableStatusVariableName("QueryString")
	MatchVariableStatusVariableNameRemoteAddr     = MatchVariableStatusVariableName("RemoteAddr")
	MatchVariableStatusVariableNameRequestBody    = MatchVariableStatusVariableName("RequestBody")
	MatchVariableStatusVariableNameRequestCookies = MatchVariableStatusVariableName("RequestCookies")
	MatchVariableStatusVariableNameRequestHeaders = MatchVariableStatusVariableName("RequestHeaders")
	MatchVariableStatusVariableNameRequestMethod  = MatchVariableStatusVariableName("RequestMethod")
	MatchVariableStatusVariableNameRequestUri     = MatchVariableStatusVariableName("RequestUri")
)

// +kubebuilder:validation:Enum={"RequestArgNames","RequestCookieNames","RequestHeaderNames"}
type OwaspCrsExclusionEntryMatchVariable string

const (
	OwaspCrsExclusionEntryMatchVariableRequestArgNames    = OwaspCrsExclusionEntryMatchVariable("RequestArgNames")
	OwaspCrsExclusionEntryMatchVariableRequestCookieNames = OwaspCrsExclusionEntryMatchVariable("RequestCookieNames")
	OwaspCrsExclusionEntryMatchVariableRequestHeaderNames = OwaspCrsExclusionEntryMatchVariable("RequestHeaderNames")
)

// +kubebuilder:validation:Enum={"Contains","EndsWith","Equals","EqualsAny","StartsWith"}
type OwaspCrsExclusionEntrySelectorMatchOperator string

const (
	OwaspCrsExclusionEntrySelectorMatchOperatorContains   = OwaspCrsExclusionEntrySelectorMatchOperator("Contains")
	OwaspCrsExclusionEntrySelectorMatchOperatorEndsWith   = OwaspCrsExclusionEntrySelectorMatchOperator("EndsWith")
	OwaspCrsExclusionEntrySelectorMatchOperatorEquals     = OwaspCrsExclusionEntrySelectorMatchOperator("Equals")
	OwaspCrsExclusionEntrySelectorMatchOperatorEqualsAny  = OwaspCrsExclusionEntrySelectorMatchOperator("EqualsAny")
	OwaspCrsExclusionEntrySelectorMatchOperatorStartsWith = OwaspCrsExclusionEntrySelectorMatchOperator("StartsWith")
)

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/ManagedRuleOverride
type ManagedRuleOverride struct {

	// +kubebuilder:validation:Required
	//RuleId: Identifier for the managed rule.
	RuleId string `json:"ruleId"`

	//State: The state of the managed rule. Defaults to Disabled if not specified.
	State *ManagedRuleOverrideState `json:"state,omitempty"`
}

// +kubebuilder:validation:Enum={"Disabled"}
type ManagedRuleOverrideStatusState string

const ManagedRuleOverrideStatusStateDisabled = ManagedRuleOverrideStatusState("Disabled")

// +kubebuilder:validation:Enum={"PostArgs","QueryString","RemoteAddr","RequestBody","RequestCookies","RequestHeaders","RequestMethod","RequestUri"}
type MatchVariableVariableName string

const (
	MatchVariableVariableNamePostArgs       = MatchVariableVariableName("PostArgs")
	MatchVariableVariableNameQueryString    = MatchVariableVariableName("QueryString")
	MatchVariableVariableNameRemoteAddr     = MatchVariableVariableName("RemoteAddr")
	MatchVariableVariableNameRequestBody    = MatchVariableVariableName("RequestBody")
	MatchVariableVariableNameRequestCookies = MatchVariableVariableName("RequestCookies")
	MatchVariableVariableNameRequestHeaders = MatchVariableVariableName("RequestHeaders")
	MatchVariableVariableNameRequestMethod  = MatchVariableVariableName("RequestMethod")
	MatchVariableVariableNameRequestUri     = MatchVariableVariableName("RequestUri")
)

// +kubebuilder:validation:Enum={"Disabled"}
type ManagedRuleOverrideState string

const ManagedRuleOverrideStateDisabled = ManagedRuleOverrideState("Disabled")

func init() {
	SchemeBuilder.Register(&ApplicationGatewayWebApplicationFirewallPolicies{}, &ApplicationGatewayWebApplicationFirewallPoliciesList{})
}
