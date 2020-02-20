/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (

	// LoadBalancerSpec defines the desired state of LoadBalancer
	LoadBalancerSpec struct {
	}

	// LoadBalancerStatus defines the observed state of LoadBalancer
	LoadBalancerStatus struct {
		ID                string `json:"id,omitempty"`
		DeploymentID      string `json:"deploymentId,omitempty"`
		ProvisioningState string `json:"provisioningState,omitempty"`
	}

	// +kubebuilder:object:root=true
	// +kubebuilder:subresource:status
	// +kubebuilder:storageversion

	// LoadBalancer is the Schema for the loadbalancers API
	LoadBalancer struct {
		metav1.TypeMeta   `json:",inline"`
		metav1.ObjectMeta `json:"metadata,omitempty"`

		Spec   LoadBalancerSpec   `json:"spec,omitempty"`
		Status LoadBalancerStatus `json:"status,omitempty"`
	}

	// +kubebuilder:object:root=true

	// LoadBalancerList contains a list of LoadBalancer
	LoadBalancerList struct {
		metav1.TypeMeta `json:",inline"`
		metav1.ListMeta `json:"metadata,omitempty"`
		Items           []LoadBalancer `json:"items"`
	}
)

func init() {
	SchemeBuilder.Register(&LoadBalancer{}, &LoadBalancerList{})
}
