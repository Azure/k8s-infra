/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package v1

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/k8s-infra/pkg/zips"
)

// NetworkSecurityGroupSpec defines the desired state of NetworkSecurityGroup
type (
	SecurityRuleSpecProperties struct {
		// +kubebuilder:validation:Enum=Allow;Deny
		Access string `json:"access,omitempty"`

		// +kubebuiler:validation:MaxLength=140
		Description string `json:"description,omitempty"`

		DestinationAddressPrefix string   `json:"destinationAddressPrefix,omitempty"`
		DestinationPortRange     string   `json:"destinationPortRange,omitempty"`
		DestinationPortRanges    []string `json:"destinationPortRanges,omitempty"`

		// +kubebuilder:validation:Enum=Inbound;Outbound
		Direction string `json:"direction,omitempty"`
		Priority  int    `json:"priority,omitempty"`

		// +kubebuilder:validation:Enum=*;Ah;Esp;Icmp;Tcp;Udp
		Protocol string `json:"protocol,omitempty"`

		ProvisioningState     string   `json:"provisioningState,omitempty"`
		SourceAddressPrefix   string   `json:"sourceAddressPrefix,omitempty"`
		SourceAddressPrefixes string   `json:"sourceAddressPrefixes,omitempty"`
		SourcePortRange       string   `json:"sourcePortRange,omitempty"`
		SourcePortRanges      []string `json:"sourcePortRanges,omitempty"`
	}

	SecurityRuleSpec struct {
		ID         string                      `json:"id,omitempty"`
		Name       string                      `json:"name,omitempty"`
		Properties *SecurityRuleSpecProperties `json:"properties,omitempty"`
	}

	NetworkSecurityGroupSpecProperties struct {
		SecurityRules []SecurityRuleSpec `json:"securityRules,omitempty"`
	}

	NetworkSecurityGroupSpec struct {
		APIVersion string `json:"apiVersion,omitempty"`

		// ResourceGroup is the Azure Resource Group the VirtualNetwork resides within
		// +kubebuilder:validation:Required
		ResourceGroup *corev1.ObjectReference `json:"group"`

		// Location of the VNET in Azure
		// +kubebuilder:validation:Required
		Location string `json:"location"`

		// Tags are user defined key value pairs
		// +optional
		Tags map[string]string `json:"tags,omitempty"`

		// Properties of the Virtual Network
		Properties *NetworkSecurityGroupSpecProperties `json:"properties,omitempty"`
	}

	// NetworkSecurityGroupStatus defines the observed state of NetworkSecurityGroup
	NetworkSecurityGroupStatus struct {
		ID                string `json:"id,omitempty"`
		DeploymentID      string `json:"deploymentId,omitempty"`
		ProvisioningState string `json:"provisioningState,omitempty"`
	}

	// +kubebuilder:object:root=true
	// +kubebuilder:subresource:status
	// +kubebuilder:storageversion

	// NetworkSecurityGroup is the Schema for the networksecuritygroups API
	NetworkSecurityGroup struct {
		metav1.TypeMeta   `json:",inline"`
		metav1.ObjectMeta `json:"metadata,omitempty"`

		Spec   NetworkSecurityGroupSpec   `json:"spec,omitempty"`
		Status NetworkSecurityGroupStatus `json:"status,omitempty"`
	}

	// +kubebuilder:object:root=true

	// NetworkSecurityGroupList contains a list of NetworkSecurityGroup
	NetworkSecurityGroupList struct {
		metav1.TypeMeta `json:",inline"`
		metav1.ListMeta `json:"metadata,omitempty"`
		Items           []NetworkSecurityGroup `json:"items"`
	}
)

func (*NetworkSecurityGroup) Hub() {}

func (nsg *NetworkSecurityGroup) ToResource() (zips.Resource, error) {
	rgName := ""
	if nsg.Spec.ResourceGroup != nil {
		rgName = nsg.Spec.ResourceGroup.Name
	}

	res := zips.Resource{
		ID:                nsg.Status.ID,
		DeploymentID:      nsg.Status.DeploymentID,
		Type:              "Microsoft.Network/networkSecurityGroups",
		ResourceGroup:     rgName,
		Name:              nsg.Name,
		APIVersion:        nsg.Spec.APIVersion,
		Location:          nsg.Spec.Location,
		Tags:              nsg.Spec.Tags,
		ProvisioningState: zips.ProvisioningState(nsg.Status.ProvisioningState),
	}

	bits, err := json.Marshal(nsg.Spec.Properties)
	if err != nil {
		return res, err
	}
	res.Properties = bits

	return *res.SetAnnotations(nsg.Annotations), nil
}

func (nsg *NetworkSecurityGroup) FromResource(res zips.Resource) error {
	nsg.Status.ID = res.ID
	nsg.Status.DeploymentID = res.DeploymentID
	nsg.Status.ProvisioningState = string(res.ProvisioningState)
	nsg.Spec.Tags = res.Tags

	var props NetworkSecurityGroupSpecProperties
	if err := json.Unmarshal(res.Properties, &props); err != nil {
		return err
	}
	nsg.Spec.Properties = &props
	return nil
}

func init() {
	SchemeBuilder.Register(&NetworkSecurityGroup{}, &NetworkSecurityGroupList{})
}
