/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package v1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	azcorev1 "github.com/Azure/k8s-infra/apis/core/v1"
	"github.com/Azure/k8s-infra/pkg/zips"
)

// NetworkSecurityGroupSpec defines the desired state of NetworkSecurityGroup
type (
	NetworkSecurityGroupSpecProperties struct {
		SecurityRuleRefs []azcorev1.KnownTypeReference `json:"securityRules,omitempty" group:"microsoft.network.infra.azure.com" kind:"SecurityRule"`
	}

	NetworkSecurityGroupSpec struct {
		// +k8s:conversion-gen=false
		APIVersion string `json:"apiVersion,omitempty"`

		// ResourceGroupRef is the Azure Resource Group the VirtualNetwork resides within
		// +kubebuilder:validation:Required
		ResourceGroupRef *azcorev1.KnownTypeReference `json:"groupRef" group:"microsoft.resources.infra.azure.com" kind:"ResourceGroup"`

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
		ID string `json:"id,omitempty"`
		// +k8s:conversion-gen=false
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

func (nsg *NetworkSecurityGroup) GetResourceGroupObjectRef() *azcorev1.KnownTypeReference {
	return nsg.Spec.ResourceGroupRef
}

func (nsg *NetworkSecurityGroup) ToResource() (zips.Resource, error) {
	rgName := ""
	if nsg.Spec.ResourceGroupRef != nil {
		rgName = nsg.Spec.ResourceGroupRef.Name
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
