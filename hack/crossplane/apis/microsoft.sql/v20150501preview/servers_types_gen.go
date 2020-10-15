// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
// Code generated by k8s-infra-gen. DO NOT EDIT.
package v20150501preview

import (
	"fmt"
	"github.com/Azure/k8s-infra/hack/generated/apis/deploymenttemplate/v20150101"
	"github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
type Servers struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              Servers_Spec  `json:"spec,omitempty"`
	Status            Server_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ServersList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Servers `json:"items"`
}

type Server_Status struct {
	AtProvider ServersObservation `json:"atProvider"`
}

type Servers_Spec struct {
	ForProvider ServersParameters `json:"forProvider"`
}

type ServersObservation struct {

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Identity: The Azure Active Directory identity of the server.
	Identity *ResourceIdentity_Status `json:"identity,omitempty"`

	//Kind: Kind of sql server. This is metadata used for the Azure portal experience.
	Kind *string `json:"kind,omitempty"`

	// +kubebuilder:validation:Required
	//Location: Resource location.
	Location string `json:"location"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Resource properties.
	Properties *ServerProperties_Status `json:"properties,omitempty"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type ServersParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion ServersSpecApiVersion `json:"apiVersion"`
	Comments   *string               `json:"comments,omitempty"`

	//Condition: Condition of the resource
	Condition *bool                   `json:"condition,omitempty"`
	Copy      *v20150101.ResourceCopy `json:"copy,omitempty"`

	//DependsOn: Collection of resources this resource depends on
	DependsOn []string `json:"dependsOn,omitempty"`

	//Identity: Azure Active Directory identity configuration for a resource.
	Identity *ResourceIdentity `json:"identity,omitempty"`

	//Location: Location to deploy resource to
	Location string `json:"location,omitempty"`

	// +kubebuilder:validation:Required
	//Name: Name of the resource
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	//Properties: The properties of a server.
	Properties ServerProperties `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type ServersSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2015-05-01-preview/Microsoft.Sql.json#/definitions/ServerProperties
type ServerProperties struct {

	//AdministratorLogin: Administrator username for the server. Once created it
	//cannot be changed.
	AdministratorLogin *string `json:"administratorLogin,omitempty"`

	//AdministratorLoginPassword: The administrator login password (required for
	//server creation).
	AdministratorLoginPassword *string `json:"administratorLoginPassword,omitempty"`

	//Version: The version of the server.
	Version *string `json:"version,omitempty"`
}

//Generated from:
type ServerProperties_Status struct {

	//AdministratorLogin: Administrator username for the server. Once created it
	//cannot be changed.
	AdministratorLogin *string `json:"administratorLogin,omitempty"`

	//AdministratorLoginPassword: The administrator login password (required for
	//server creation).
	AdministratorLoginPassword *string `json:"administratorLoginPassword,omitempty"`

	//FullyQualifiedDomainName: The fully qualified domain name of the server.
	FullyQualifiedDomainName *string `json:"fullyQualifiedDomainName,omitempty"`

	//State: The state of the server.
	State *string `json:"state,omitempty"`

	//Version: The version of the server.
	Version *string `json:"version,omitempty"`
}

// +kubebuilder:validation:Enum={"2015-05-01-preview"}
type ServersSpecApiVersion string

const ServersSpecApiVersion20150501Preview = ServersSpecApiVersion("2015-05-01-preview")

// +kubebuilder:validation:Enum={"Microsoft.Sql/servers"}
type ServersSpecType string

const ServersSpecTypeMicrosoftSqlServers = ServersSpecType("Microsoft.Sql/servers")

func init() {
	SchemeBuilder.Register(&Servers{}, &ServersList{})
}
