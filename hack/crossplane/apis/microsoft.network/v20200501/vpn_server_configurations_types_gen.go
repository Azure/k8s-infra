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
type VpnServerConfigurations struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VpnServerConfigurations_Spec  `json:"spec,omitempty"`
	Status            VpnServerConfiguration_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type VpnServerConfigurationsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VpnServerConfigurations `json:"items"`
}

//Generated from:
type VpnServerConfiguration_Status struct {

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Location: Resource location.
	Location *string `json:"location,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the P2SVpnServer configuration.
	Properties *VpnServerConfigurationProperties_Status `json:"properties,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type VpnServerConfigurations_Spec struct {
	ForProvider VpnServerConfigurationsParameters `json:"forProvider"`
}

//Generated from:
type VpnServerConfigurationProperties_Status struct {

	//AadAuthenticationParameters: The set of aad vpn authentication parameters.
	AadAuthenticationParameters *AadAuthenticationParameters_Status `json:"aadAuthenticationParameters,omitempty"`

	//Etag: A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`

	//Name: The name of the VpnServerConfiguration that is unique within a resource
	//group.
	Name *string `json:"name,omitempty"`

	//P2SVpnGateways: List of references to P2SVpnGateways.
	P2SVpnGateways []P2SVpnGateway_Status `json:"p2SVpnGateways,omitempty"`

	//ProvisioningState: The provisioning state of the VpnServerConfiguration
	//resource. Possible values are: 'Updating', 'Deleting', and 'Failed'.
	ProvisioningState *string `json:"provisioningState,omitempty"`

	//RadiusClientRootCertificates: Radius client root certificate of
	//VpnServerConfiguration.
	RadiusClientRootCertificates []VpnServerConfigRadiusClientRootCertificate_Status `json:"radiusClientRootCertificates,omitempty"`

	//RadiusServerAddress: The radius server address property of the
	//VpnServerConfiguration resource for point to site client connection.
	RadiusServerAddress *string `json:"radiusServerAddress,omitempty"`

	//RadiusServerRootCertificates: Radius Server root certificate of
	//VpnServerConfiguration.
	RadiusServerRootCertificates []VpnServerConfigRadiusServerRootCertificate_Status `json:"radiusServerRootCertificates,omitempty"`

	//RadiusServerSecret: The radius secret property of the VpnServerConfiguration
	//resource for point to site client connection.
	RadiusServerSecret *string `json:"radiusServerSecret,omitempty"`

	//RadiusServers: Multiple Radius Server configuration for VpnServerConfiguration.
	RadiusServers []RadiusServer_Status `json:"radiusServers,omitempty"`

	//VpnAuthenticationTypes: VPN authentication types for the VpnServerConfiguration.
	VpnAuthenticationTypes []VpnServerConfigurationPropertiesStatusVpnAuthenticationTypes `json:"vpnAuthenticationTypes,omitempty"`

	//VpnClientIpsecPolicies: VpnClientIpsecPolicies for VpnServerConfiguration.
	VpnClientIpsecPolicies []IpsecPolicy_Status `json:"vpnClientIpsecPolicies,omitempty"`

	//VpnClientRevokedCertificates: VPN client revoked certificate of
	//VpnServerConfiguration.
	VpnClientRevokedCertificates []VpnServerConfigVpnClientRevokedCertificate_Status `json:"vpnClientRevokedCertificates,omitempty"`

	//VpnClientRootCertificates: VPN client root certificate of VpnServerConfiguration.
	VpnClientRootCertificates []VpnServerConfigVpnClientRootCertificate_Status `json:"vpnClientRootCertificates,omitempty"`

	//VpnProtocols: VPN protocols for the VpnServerConfiguration.
	VpnProtocols []VpnServerConfigurationPropertiesStatusVpnProtocols `json:"vpnProtocols,omitempty"`
}

type VpnServerConfigurationsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion VpnServerConfigurationsSpecApiVersion `json:"apiVersion"`
	Comments   *string                               `json:"comments,omitempty"`

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
	//Properties: Properties of the P2SVpnServer configuration.
	Properties VpnServerConfigurationProperties `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type VpnServerConfigurationsSpecType `json:"type"`
}

//Generated from:
type AadAuthenticationParameters_Status struct {

	//AadAudience: AAD Vpn authentication parameter AAD audience.
	AadAudience *string `json:"aadAudience,omitempty"`

	//AadIssuer: AAD Vpn authentication parameter AAD issuer.
	AadIssuer *string `json:"aadIssuer,omitempty"`

	//AadTenant: AAD Vpn authentication parameter AAD tenant.
	AadTenant *string `json:"aadTenant,omitempty"`
}

//Generated from:
type RadiusServer_Status struct {

	// +kubebuilder:validation:Required
	//RadiusServerAddress: The address of this radius server.
	RadiusServerAddress string `json:"radiusServerAddress"`

	//RadiusServerScore: The initial score assigned to this radius server.
	RadiusServerScore *int `json:"radiusServerScore,omitempty"`

	//RadiusServerSecret: The secret used for this radius server.
	RadiusServerSecret *string `json:"radiusServerSecret,omitempty"`
}

//Generated from:
type VpnServerConfigRadiusClientRootCertificate_Status struct {

	//Name: The certificate name.
	Name *string `json:"name,omitempty"`

	//Thumbprint: The Radius client root certificate thumbprint.
	Thumbprint *string `json:"thumbprint,omitempty"`
}

//Generated from:
type VpnServerConfigRadiusServerRootCertificate_Status struct {

	//Name: The certificate name.
	Name *string `json:"name,omitempty"`

	//PublicCertData: The certificate public data.
	PublicCertData *string `json:"publicCertData,omitempty"`
}

//Generated from:
type VpnServerConfigVpnClientRevokedCertificate_Status struct {

	//Name: The certificate name.
	Name *string `json:"name,omitempty"`

	//Thumbprint: The revoked VPN client certificate thumbprint.
	Thumbprint *string `json:"thumbprint,omitempty"`
}

//Generated from:
type VpnServerConfigVpnClientRootCertificate_Status struct {

	//Name: The certificate name.
	Name *string `json:"name,omitempty"`

	//PublicCertData: The certificate public data.
	PublicCertData *string `json:"publicCertData,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/VpnServerConfigurationProperties
type VpnServerConfigurationProperties struct {

	//AadAuthenticationParameters: The set of aad vpn authentication parameters.
	AadAuthenticationParameters *AadAuthenticationParameters `json:"aadAuthenticationParameters,omitempty"`

	//Name: The name of the VpnServerConfiguration that is unique within a resource
	//group.
	Name *string `json:"name,omitempty"`

	//RadiusClientRootCertificates: Radius client root certificate of
	//VpnServerConfiguration.
	RadiusClientRootCertificates []VpnServerConfigRadiusClientRootCertificate `json:"radiusClientRootCertificates,omitempty"`

	//RadiusServerAddress: The radius server address property of the
	//VpnServerConfiguration resource for point to site client connection.
	RadiusServerAddress *string `json:"radiusServerAddress,omitempty"`

	//RadiusServerRootCertificates: Radius Server root certificate of
	//VpnServerConfiguration.
	RadiusServerRootCertificates []VpnServerConfigRadiusServerRootCertificate `json:"radiusServerRootCertificates,omitempty"`

	//RadiusServerSecret: The radius secret property of the VpnServerConfiguration
	//resource for point to site client connection.
	RadiusServerSecret *string `json:"radiusServerSecret,omitempty"`

	//RadiusServers: Multiple Radius Server configuration for VpnServerConfiguration.
	RadiusServers []RadiusServer `json:"radiusServers,omitempty"`

	//VpnAuthenticationTypes: VPN authentication types for the VpnServerConfiguration.
	VpnAuthenticationTypes []VpnServerConfigurationPropertiesVpnAuthenticationTypes `json:"vpnAuthenticationTypes,omitempty"`

	//VpnClientIpsecPolicies: VpnClientIpsecPolicies for VpnServerConfiguration.
	VpnClientIpsecPolicies []IpsecPolicy `json:"vpnClientIpsecPolicies,omitempty"`

	//VpnClientRevokedCertificates: VPN client revoked certificate of
	//VpnServerConfiguration.
	VpnClientRevokedCertificates []VpnServerConfigVpnClientRevokedCertificate `json:"vpnClientRevokedCertificates,omitempty"`

	//VpnClientRootCertificates: VPN client root certificate of VpnServerConfiguration.
	VpnClientRootCertificates []VpnServerConfigVpnClientRootCertificate `json:"vpnClientRootCertificates,omitempty"`

	//VpnProtocols: VPN protocols for the VpnServerConfiguration.
	VpnProtocols []VpnServerConfigurationPropertiesVpnProtocols `json:"vpnProtocols,omitempty"`
}

// +kubebuilder:validation:Enum={"AAD","Certificate","Radius"}
type VpnServerConfigurationPropertiesStatusVpnAuthenticationTypes string

const (
	VpnServerConfigurationPropertiesStatusVpnAuthenticationTypesAAD         = VpnServerConfigurationPropertiesStatusVpnAuthenticationTypes("AAD")
	VpnServerConfigurationPropertiesStatusVpnAuthenticationTypesCertificate = VpnServerConfigurationPropertiesStatusVpnAuthenticationTypes("Certificate")
	VpnServerConfigurationPropertiesStatusVpnAuthenticationTypesRadius      = VpnServerConfigurationPropertiesStatusVpnAuthenticationTypes("Radius")
)

// +kubebuilder:validation:Enum={"IkeV2","OpenVPN"}
type VpnServerConfigurationPropertiesStatusVpnProtocols string

const (
	VpnServerConfigurationPropertiesStatusVpnProtocolsIkeV2   = VpnServerConfigurationPropertiesStatusVpnProtocols("IkeV2")
	VpnServerConfigurationPropertiesStatusVpnProtocolsOpenVPN = VpnServerConfigurationPropertiesStatusVpnProtocols("OpenVPN")
)

// +kubebuilder:validation:Enum={"2020-05-01"}
type VpnServerConfigurationsSpecApiVersion string

const VpnServerConfigurationsSpecApiVersion20200501 = VpnServerConfigurationsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/vpnServerConfigurations"}
type VpnServerConfigurationsSpecType string

const VpnServerConfigurationsSpecTypeMicrosoftNetworkVpnServerConfigurations = VpnServerConfigurationsSpecType("Microsoft.Network/vpnServerConfigurations")

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/AadAuthenticationParameters
type AadAuthenticationParameters struct {

	//AadAudience: AAD Vpn authentication parameter AAD audience.
	AadAudience *string `json:"aadAudience,omitempty"`

	//AadIssuer: AAD Vpn authentication parameter AAD issuer.
	AadIssuer *string `json:"aadIssuer,omitempty"`

	//AadTenant: AAD Vpn authentication parameter AAD tenant.
	AadTenant *string `json:"aadTenant,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/RadiusServer
type RadiusServer struct {

	// +kubebuilder:validation:Required
	//RadiusServerAddress: The address of this radius server.
	RadiusServerAddress string `json:"radiusServerAddress"`

	//RadiusServerScore: The initial score assigned to this radius server.
	RadiusServerScore *int `json:"radiusServerScore,omitempty"`

	//RadiusServerSecret: The secret used for this radius server.
	RadiusServerSecret *string `json:"radiusServerSecret,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/VpnServerConfigRadiusClientRootCertificate
type VpnServerConfigRadiusClientRootCertificate struct {

	//Name: The certificate name.
	Name *string `json:"name,omitempty"`

	//Thumbprint: The Radius client root certificate thumbprint.
	Thumbprint *string `json:"thumbprint,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/VpnServerConfigRadiusServerRootCertificate
type VpnServerConfigRadiusServerRootCertificate struct {

	//Name: The certificate name.
	Name *string `json:"name,omitempty"`

	//PublicCertData: The certificate public data.
	PublicCertData *string `json:"publicCertData,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/VpnServerConfigVpnClientRevokedCertificate
type VpnServerConfigVpnClientRevokedCertificate struct {

	//Name: The certificate name.
	Name *string `json:"name,omitempty"`

	//Thumbprint: The revoked VPN client certificate thumbprint.
	Thumbprint *string `json:"thumbprint,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/VpnServerConfigVpnClientRootCertificate
type VpnServerConfigVpnClientRootCertificate struct {

	//Name: The certificate name.
	Name *string `json:"name,omitempty"`

	//PublicCertData: The certificate public data.
	PublicCertData *string `json:"publicCertData,omitempty"`
}

// +kubebuilder:validation:Enum={"AAD","Certificate","Radius"}
type VpnServerConfigurationPropertiesVpnAuthenticationTypes string

const (
	VpnServerConfigurationPropertiesVpnAuthenticationTypesAAD         = VpnServerConfigurationPropertiesVpnAuthenticationTypes("AAD")
	VpnServerConfigurationPropertiesVpnAuthenticationTypesCertificate = VpnServerConfigurationPropertiesVpnAuthenticationTypes("Certificate")
	VpnServerConfigurationPropertiesVpnAuthenticationTypesRadius      = VpnServerConfigurationPropertiesVpnAuthenticationTypes("Radius")
)

// +kubebuilder:validation:Enum={"IkeV2","OpenVPN"}
type VpnServerConfigurationPropertiesVpnProtocols string

const (
	VpnServerConfigurationPropertiesVpnProtocolsIkeV2   = VpnServerConfigurationPropertiesVpnProtocols("IkeV2")
	VpnServerConfigurationPropertiesVpnProtocolsOpenVPN = VpnServerConfigurationPropertiesVpnProtocols("OpenVPN")
)

func init() {
	SchemeBuilder.Register(&VpnServerConfigurations{}, &VpnServerConfigurationsList{})
}
