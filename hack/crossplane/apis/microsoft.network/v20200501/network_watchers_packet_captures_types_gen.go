// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
// Code generated by k8s-infra-gen. DO NOT EDIT.
package v20200501

import (
	"github.com/Azure/k8s-infra/hack/crossplane/apis/deploymenttemplate/v20150101"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/resourceDefinitions/networkWatchers_packetCaptures
type NetworkWatchersPacketCaptures struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              NetworkWatchersPacketCaptures_Spec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/resourceDefinitions/networkWatchers_packetCaptures
type NetworkWatchersPacketCapturesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkWatchersPacketCaptures `json:"items"`
}

type NetworkWatchersPacketCaptures_Spec struct {
	ForProvider NetworkWatchersPacketCapturesParameters `json:"forProvider"`
}

type NetworkWatchersPacketCapturesParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion NetworkWatchersPacketCapturesSpecApiVersion `json:"apiVersion"`
	Comments   *string                                     `json:"comments,omitempty"`

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
	//Properties: Properties of the packet capture.
	Properties PacketCaptureParameters `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type NetworkWatchersPacketCapturesSpecType `json:"type"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type NetworkWatchersPacketCapturesSpecApiVersion string

const NetworkWatchersPacketCapturesSpecApiVersion20200501 = NetworkWatchersPacketCapturesSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/networkWatchers/packetCaptures"}
type NetworkWatchersPacketCapturesSpecType string

const NetworkWatchersPacketCapturesSpecTypeMicrosoftNetworkNetworkWatchersPacketCaptures = NetworkWatchersPacketCapturesSpecType("Microsoft.Network/networkWatchers/packetCaptures")

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/PacketCaptureParameters
type PacketCaptureParameters struct {

	//BytesToCapturePerPacket: Number of bytes captured per packet, the remaining
	//bytes are truncated.
	BytesToCapturePerPacket *int `json:"bytesToCapturePerPacket,omitempty"`

	//Filters: A list of packet capture filters.
	Filters []PacketCaptureFilter `json:"filters,omitempty"`

	// +kubebuilder:validation:Required
	//StorageLocation: The storage location for a packet capture session.
	StorageLocation PacketCaptureStorageLocation `json:"storageLocation"`

	// +kubebuilder:validation:Required
	//Target: The ID of the targeted resource, only VM is currently supported.
	Target string `json:"target"`

	//TimeLimitInSeconds: Maximum duration of the capture session in seconds.
	TimeLimitInSeconds *int `json:"timeLimitInSeconds,omitempty"`

	//TotalBytesPerSession: Maximum size of the capture output.
	TotalBytesPerSession *int `json:"totalBytesPerSession,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/PacketCaptureFilter
type PacketCaptureFilter struct {

	//LocalIPAddress: Local IP Address to be filtered on. Notation: "127.0.0.1" for
	//single address entry. "127.0.0.1-127.0.0.255" for range. "127.0.0.1;127.0.0.5"?
	//for multiple entries. Multiple ranges not currently supported. Mixing ranges
	//with multiple entries not currently supported. Default = null.
	LocalIPAddress *string `json:"localIPAddress,omitempty"`

	//LocalPort: Local port to be filtered on. Notation: "80" for single port
	//entry."80-85" for range. "80;443;" for multiple entries. Multiple ranges not
	//currently supported. Mixing ranges with multiple entries not currently
	//supported. Default = null.
	LocalPort *string `json:"localPort,omitempty"`

	//Protocol: Protocol to be filtered on.
	Protocol *PacketCaptureFilterProtocol `json:"protocol,omitempty"`

	//RemoteIPAddress: Local IP Address to be filtered on. Notation: "127.0.0.1" for
	//single address entry. "127.0.0.1-127.0.0.255" for range. "127.0.0.1;127.0.0.5;"
	//for multiple entries. Multiple ranges not currently supported. Mixing ranges
	//with multiple entries not currently supported. Default = null.
	RemoteIPAddress *string `json:"remoteIPAddress,omitempty"`

	//RemotePort: Remote port to be filtered on. Notation: "80" for single port
	//entry."80-85" for range. "80;443;" for multiple entries. Multiple ranges not
	//currently supported. Mixing ranges with multiple entries not currently
	//supported. Default = null.
	RemotePort *string `json:"remotePort,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/PacketCaptureStorageLocation
type PacketCaptureStorageLocation struct {

	//FilePath: A valid local path on the targeting VM. Must include the name of the
	//capture file (*.cap). For linux virtual machine it must start with
	///var/captures. Required if no storage ID is provided, otherwise optional.
	FilePath *string `json:"filePath,omitempty"`

	//StorageId: The ID of the storage account to save the packet capture session.
	//Required if no local file path is provided.
	StorageId *string `json:"storageId,omitempty"`

	//StoragePath: The URI of the storage path to save the packet capture. Must be a
	//well-formed URI describing the location to save the packet capture.
	StoragePath *string `json:"storagePath,omitempty"`
}

// +kubebuilder:validation:Enum={"Any","TCP","UDP"}
type PacketCaptureFilterProtocol string

const (
	PacketCaptureFilterProtocolAny = PacketCaptureFilterProtocol("Any")
	PacketCaptureFilterProtocolTCP = PacketCaptureFilterProtocol("TCP")
	PacketCaptureFilterProtocolUDP = PacketCaptureFilterProtocol("UDP")
)

func init() {
	SchemeBuilder.Register(&NetworkWatchersPacketCaptures{}, &NetworkWatchersPacketCapturesList{})
}
