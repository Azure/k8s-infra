/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package v20191101

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	azcorev1 "github.com/Azure/k8s-infra/apis/core/v1"
)

type (
	// DHCPOptionsSpec contains an array of DNS servers available to VMs deployed in the virtual network
	DHCPOptionsSpec struct {
		// DNSServers a list of DNS servers IP addresses
		DNSServers []string `json:"dnsServers,omitempty"`
	}

	// AddressSpaceSpec contains an array of IP address ranges that can be used by subnets
	AddressSpaceSpec struct {
		// AddressPrefixes are a list of address blocks reserved for this virtual network in CIDR notation
		AddressPrefixes []string `json:"addressPrefixes,omitempty"`
	}

	// BGPCommunitiesSpec are BGP communities sent over ExpressRoute with each route corresponding to a prefix in this VNET
	BGPCommunitiesSpec struct {
		// RegionalCommunity is a BGP community associated with the region of the virtual network
		RegionalCommunity string `json:"regionalCommunity,omitempty"`

		// VirtualNetworkCommunity is the BGP community associated with the virtual network
		VirtualNetworkCommunity string `json:"virtualNetworkCommunity,omitempty"`
	}

	// SubnetProperties are the properties of the subnet
	SubnetProperties struct {
		// AddressPrefix for the subnet, eg. 10.0.0.0/24
		AddressPrefix string `json:"addressPrefix,omitempty"`

		// AddressPrefixes are a list of address prefixes for a subnet
		AddressPrefixes []string `json:"addressPrefixes,omitempty"`
	}

	// SubnetSpec is a subnet in a Virtual Network
	// TODO: (dj) I think this should probably be a slice of corev1.ObjectReference
	SubnetSpec struct {
		// ID of the subnet resource
		ID string `json:"id,omitempty"`

		// Name of the subnet
		Name string `json:"name,omitempty"`

		// Properties of the subnet
		Properties SubnetProperties `json:"properties,omitempty"`
	}

	// VirtualNetworkSpecProperties are the property bodies to be applied
	VirtualNetworkSpecProperties struct {

		// AddressSpace contains an array of IP address ranges that can be used by subnets
		AddressSpace *AddressSpaceSpec `json:"addressSpace,omitempty"`

		// BGPCommunities are BGP communities sent over ExpressRoute with each route corresponding to a prefix in this VNET
		// +optional
		BGPCommunities *BGPCommunitiesSpec `json:"bgpCommunities,omitempty"`

		// DHCPOptions contains an array of DNS servers available to VMs deployed in the virtual network
		// +optional
		DHCPOptions *DHCPOptionsSpec `json:"dhcpOptions,omitempty"`

		// Subnets is a list of subnets in the VNET
		// +optional
		Subnets []SubnetSpec `json:"subnets,omitempty"`

		// EnableVMProtection indicates if VM protection is enabled for all the subnets in the virtual network
		// +optional
		EnableVMProtection bool `json:"enableVMProtection,omitempty"`
	}

	// VirtualNetworkSpec defines the desired state of VirtualNetwork
	VirtualNetworkSpec struct {
		// ResourceGroupRef is the Azure Resource Group the VirtualNetwork resides within
		ResourceGroupRef *azcorev1.KnownTypeReference `json:"resourceGroupRef" group:"microsoft.resources.infra.azure.com" kind:"ResourceGroup"`

		// Location of the VNET in Azure
		Location string `json:"location"`

		// Tags are user defined key value pairs
		// +optional
		Tags map[string]string `json:"tags,omitempty"`

		// Properties of the Virtual Network
		Properties *VirtualNetworkSpecProperties `json:"properties,omitempty"`
	}

	// VirtualNetworkStatus defines the observed state of VirtualNetwork
	VirtualNetworkStatus struct {
		ID                string `json:"id,omitempty"`
		ProvisioningState string `json:"provisioningState,omitempty"`
	}

	// +kubebuilder:object:root=true

	// VirtualNetwork is the Schema for the virtualnetworks API
	VirtualNetwork struct {
		metav1.TypeMeta   `json:",inline"`
		metav1.ObjectMeta `json:"metadata,omitempty"`

		Spec   VirtualNetworkSpec   `json:"spec,omitempty"`
		Status VirtualNetworkStatus `json:"status,omitempty"`
	}

	// +kubebuilder:object:root=true

	// VirtualNetworkList contains a list of VirtualNetwork
	VirtualNetworkList struct {
		metav1.TypeMeta `json:",inline"`
		metav1.ListMeta `json:"metadata,omitempty"`
		Items           []VirtualNetwork `json:"items"`
	}
)

func init() {
	SchemeBuilder.Register(&VirtualNetwork{}, &VirtualNetworkList{})
}
