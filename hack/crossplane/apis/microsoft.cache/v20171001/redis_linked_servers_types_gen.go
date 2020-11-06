// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
// Code generated by k8s-infra-gen. DO NOT EDIT.
package v20171001

import (
	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type RedisLinkedServers struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RedisLinkedServers_Spec                `json:"spec,omitempty"`
	Status            RedisLinkedServerWithProperties_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type RedisLinkedServersList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisLinkedServers `json:"items"`
}

type RedisLinkedServerWithProperties_Status struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              RedisLinkedServersObservation `json:"atProvider"`
}

type RedisLinkedServers_Spec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           RedisLinkedServersParameters `json:"forProvider"`
}

type RedisLinkedServersObservation struct {

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	//Properties: Properties of the linked server.
	Properties *RedisLinkedServerProperties_Status `json:"properties,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type RedisLinkedServersParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion RedisLinkedServersSpecApiVersion `json:"apiVersion"`

	//Location: Location to deploy resource to
	Location *string `json:"location,omitempty"`

	// +kubebuilder:validation:Required
	//Name: Name of the resource
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	//Properties: Create properties for a linked server
	Properties                RedisLinkedServerCreateProperties `json:"properties"`
	RedisName                 string                            `json:"redisName"`
	RedisNameRef              *v1alpha1.Reference               `json:"redisNameRef,omitempty"`
	RedisNameSelector         *v1alpha1.Selector                `json:"redisNameSelector,omitempty"`
	ResourceGroupName         string                            `json:"resourceGroupName"`
	ResourceGroupNameRef      *v1alpha1.Reference               `json:"resourceGroupNameRef,omitempty"`
	ResourceGroupNameSelector *v1alpha1.Selector                `json:"resourceGroupNameSelector,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type RedisLinkedServersSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2017-10-01/Microsoft.Cache.json#/definitions/RedisLinkedServerCreateProperties
type RedisLinkedServerCreateProperties struct {

	// +kubebuilder:validation:Required
	//LinkedRedisCacheId: Fully qualified resourceId of the linked redis cache.
	LinkedRedisCacheId string `json:"linkedRedisCacheId"`

	// +kubebuilder:validation:Required
	//LinkedRedisCacheLocation: Location of the linked redis cache.
	LinkedRedisCacheLocation string `json:"linkedRedisCacheLocation"`

	// +kubebuilder:validation:Required
	//ServerRole: Role of the linked server.
	ServerRole RedisLinkedServerCreatePropertiesServerRole `json:"serverRole"`
}

//Generated from:
type RedisLinkedServerProperties_Status struct {

	// +kubebuilder:validation:Required
	//LinkedRedisCacheId: Fully qualified resourceId of the linked redis cache.
	LinkedRedisCacheId string `json:"linkedRedisCacheId"`

	// +kubebuilder:validation:Required
	//LinkedRedisCacheLocation: Location of the linked redis cache.
	LinkedRedisCacheLocation string `json:"linkedRedisCacheLocation"`

	//ProvisioningState: Terminal state of the link between primary and secondary
	//redis cache.
	ProvisioningState *string `json:"provisioningState,omitempty"`

	// +kubebuilder:validation:Required
	//ServerRole: Role of the linked server.
	ServerRole RedisLinkedServerPropertiesStatusServerRole `json:"serverRole"`
}

// +kubebuilder:validation:Enum={"2017-10-01"}
type RedisLinkedServersSpecApiVersion string

const RedisLinkedServersSpecApiVersion20171001 = RedisLinkedServersSpecApiVersion("2017-10-01")

// +kubebuilder:validation:Enum={"Microsoft.Cache/Redis/linkedServers"}
type RedisLinkedServersSpecType string

const RedisLinkedServersSpecTypeMicrosoftCacheRedisLinkedServers = RedisLinkedServersSpecType("Microsoft.Cache/Redis/linkedServers")

// +kubebuilder:validation:Enum={"Primary","Secondary"}
type RedisLinkedServerCreatePropertiesServerRole string

const (
	RedisLinkedServerCreatePropertiesServerRolePrimary   = RedisLinkedServerCreatePropertiesServerRole("Primary")
	RedisLinkedServerCreatePropertiesServerRoleSecondary = RedisLinkedServerCreatePropertiesServerRole("Secondary")
)

// +kubebuilder:validation:Enum={"Primary","Secondary"}
type RedisLinkedServerPropertiesStatusServerRole string

const (
	RedisLinkedServerPropertiesStatusServerRolePrimary   = RedisLinkedServerPropertiesStatusServerRole("Primary")
	RedisLinkedServerPropertiesStatusServerRoleSecondary = RedisLinkedServerPropertiesStatusServerRole("Secondary")
)

func init() {
	SchemeBuilder.Register(&RedisLinkedServers{}, &RedisLinkedServersList{})
}