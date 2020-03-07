/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package v20191101

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	v1 "github.com/Azure/k8s-infra/apis/microsoft.network/v1"
)

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

	// SecurityRuleSpec defines the desired state of SecurityRule
	SecurityRuleSpec struct {
		Properties *SecurityRuleSpecProperties `json:"properties,omitempty"`
	}

	// SecurityRuleStatus defines the observed state of SecurityRule
	SecurityRuleStatus struct {
		ID                string `json:"id,omitempty"`
		ProvisioningState string `json:"provisioningState,omitempty"`
	}

	// +kubebuilder:object:root=true

	// SecurityRule is the Schema for the securityrules API
	SecurityRule struct {
		metav1.TypeMeta   `json:",inline"`
		metav1.ObjectMeta `json:"metadata,omitempty"`

		Spec   SecurityRuleSpec   `json:"spec,omitempty"`
		Status SecurityRuleStatus `json:"status,omitempty"`
	}

	// +kubebuilder:object:root=true

	// SecurityRuleList contains a list of SecurityRule
	SecurityRuleList struct {
		metav1.TypeMeta `json:",inline"`
		metav1.ListMeta `json:"metadata,omitempty"`
		Items           []SecurityRule `json:"items"`
	}
)

func (sr *SecurityRule) ConvertTo(dstRaw conversion.Hub) error {
	to := dstRaw.(*v1.SecurityRule)
	to.ObjectMeta = sr.ObjectMeta
	to.Spec.APIVersion = "2019-11-01"
	to.Status.ID = sr.Status.ID
	to.Status.ProvisioningState = sr.Status.ProvisioningState
	bits, err := json.Marshal(sr.Spec.Properties)
	if err != nil {
		return err
	}

	var props v1.SecurityRuleSpecProperties
	if err := json.Unmarshal(bits, &props); err != nil {
		return err
	}

	to.Spec.Properties = &props
	return nil
}

func (sr *SecurityRule) ConvertFrom(src conversion.Hub) error {
	from := src.(*v1.SecurityRule)
	sr.ObjectMeta = from.ObjectMeta
	sr.Status.ID = from.Status.ID
	sr.Status.ProvisioningState = from.Status.ProvisioningState

	bits, err := json.Marshal(from.Spec.Properties)
	if err != nil {
		return err
	}

	var props SecurityRuleSpecProperties
	if err := json.Unmarshal(bits, &props); err != nil {
		return err
	}
	sr.Spec.Properties = &props
	return nil
}

func init() {
	SchemeBuilder.Register(&SecurityRule{}, &SecurityRuleList{})
}
