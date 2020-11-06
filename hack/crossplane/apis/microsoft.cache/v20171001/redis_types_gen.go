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
type Redis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              Redis_Spec           `json:"spec,omitempty"`
	Status            RedisResource_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type RedisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Redis `json:"items"`
}

type RedisResource_Status struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              RedisObservation `json:"atProvider"`
}

type Redis_Spec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           RedisParameters `json:"forProvider"`
}

type RedisObservation struct {

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	// +kubebuilder:validation:Required
	//Location: The geo-location where the resource lives
	Location string `json:"location"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	// +kubebuilder:validation:Required
	//Properties: Redis cache properties.
	Properties RedisProperties_Status `json:"properties"`

	//Tags: Resource tags.
	Tags map[string]string `json:"tags,omitempty"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`

	//Zones: A list of availability zones denoting where the resource needs to come
	//from.
	Zones []string `json:"zones,omitempty"`
}

type RedisParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion RedisSpecApiVersion `json:"apiVersion"`

	//Location: Location to deploy resource to
	Location string `json:"location,omitempty"`

	// +kubebuilder:validation:Required
	//Name: Name of the resource
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	//Properties: Properties supplied to Create Redis operation.
	Properties                RedisCreateProperties `json:"properties"`
	ResourceGroupName         string                `json:"resourceGroupName"`
	ResourceGroupNameRef      *v1alpha1.Reference   `json:"resourceGroupNameRef,omitempty"`
	ResourceGroupNameSelector *v1alpha1.Selector    `json:"resourceGroupNameSelector,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type RedisSpecType `json:"type"`

	//Zones: A list of availability zones denoting where the resource needs to come
	//from.
	Zones []string `json:"zones,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2017-10-01/Microsoft.Cache.json#/definitions/RedisCreateProperties
type RedisCreateProperties struct {

	//EnableNonSslPort: Specifies whether the non-ssl Redis server port (6379) is
	//enabled.
	EnableNonSslPort *bool `json:"enableNonSslPort,omitempty"`

	//RedisConfiguration: All Redis Settings. Few possible keys:
	//rdb-backup-enabled,rdb-storage-connection-string,rdb-backup-frequency,maxmemory-delta,maxmemory-policy,notify-keyspace-events,maxmemory-samples,slowlog-log-slower-than,slowlog-max-len,list-max-ziplist-entries,list-max-ziplist-value,hash-max-ziplist-entries,hash-max-ziplist-value,set-max-intset-entries,zset-max-ziplist-entries,zset-max-ziplist-value
	//etc.
	RedisConfiguration map[string]string `json:"redisConfiguration,omitempty"`

	//ShardCount: The number of shards to be created on a Premium Cluster Cache.
	ShardCount *int `json:"shardCount,omitempty"`

	// +kubebuilder:validation:Required
	//Sku: SKU parameters supplied to the create Redis operation.
	Sku Sku `json:"sku"`

	//StaticIP: Static IP address. Required when deploying a Redis cache inside an
	//existing Azure Virtual Network.
	StaticIP *string `json:"staticIP,omitempty"`

	//SubnetId: The full resource ID of a subnet in a virtual network to deploy the
	//Redis cache in. Example format:
	///subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/Microsoft.{Network|ClassicNetwork}/VirtualNetworks/vnet1/subnets/subnet1
	SubnetId *string `json:"subnetId,omitempty"`

	//TenantSettings: A dictionary of tenant settings
	TenantSettings map[string]string `json:"tenantSettings,omitempty"`
}

//Generated from:
type RedisProperties_Status struct {

	//AccessKeys: The keys of the Redis cache - not set if this object is not the
	//response to Create or Update redis cache
	AccessKeys *RedisAccessKeys_Status `json:"accessKeys,omitempty"`

	//EnableNonSslPort: Specifies whether the non-ssl Redis server port (6379) is
	//enabled.
	EnableNonSslPort *bool `json:"enableNonSslPort,omitempty"`

	//HostName: Redis host name.
	HostName *string `json:"hostName,omitempty"`

	//LinkedServers: List of the linked servers associated with the cache
	LinkedServers []RedisLinkedServer_Status `json:"linkedServers,omitempty"`

	//Port: Redis non-SSL port.
	Port *int `json:"port,omitempty"`

	//ProvisioningState: Redis instance provisioning status.
	ProvisioningState *string `json:"provisioningState,omitempty"`

	//RedisConfiguration: All Redis Settings. Few possible keys:
	//rdb-backup-enabled,rdb-storage-connection-string,rdb-backup-frequency,maxmemory-delta,maxmemory-policy,notify-keyspace-events,maxmemory-samples,slowlog-log-slower-than,slowlog-max-len,list-max-ziplist-entries,list-max-ziplist-value,hash-max-ziplist-entries,hash-max-ziplist-value,set-max-intset-entries,zset-max-ziplist-entries,zset-max-ziplist-value
	//etc.
	RedisConfiguration map[string]string `json:"redisConfiguration,omitempty"`

	//RedisVersion: Redis version.
	RedisVersion *string `json:"redisVersion,omitempty"`

	//ShardCount: The number of shards to be created on a Premium Cluster Cache.
	ShardCount *int `json:"shardCount,omitempty"`

	// +kubebuilder:validation:Required
	//Sku: The SKU of the Redis cache to deploy.
	Sku Sku_Status `json:"sku"`

	//SslPort: Redis SSL port.
	SslPort *int `json:"sslPort,omitempty"`

	//StaticIP: Static IP address. Required when deploying a Redis cache inside an
	//existing Azure Virtual Network.
	StaticIP *string `json:"staticIP,omitempty"`

	//SubnetId: The full resource ID of a subnet in a virtual network to deploy the
	//Redis cache in. Example format:
	///subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/Microsoft.{Network|ClassicNetwork}/VirtualNetworks/vnet1/subnets/subnet1
	SubnetId *string `json:"subnetId,omitempty"`

	//TenantSettings: A dictionary of tenant settings
	TenantSettings map[string]string `json:"tenantSettings,omitempty"`
}

// +kubebuilder:validation:Enum={"2017-10-01"}
type RedisSpecApiVersion string

const RedisSpecApiVersion20171001 = RedisSpecApiVersion("2017-10-01")

// +kubebuilder:validation:Enum={"Microsoft.Cache/Redis"}
type RedisSpecType string

const RedisSpecTypeMicrosoftCacheRedis = RedisSpecType("Microsoft.Cache/Redis")

//Generated from:
type RedisAccessKeys_Status struct {

	//PrimaryKey: The current primary key that clients can use to authenticate with
	//Redis cache.
	PrimaryKey *string `json:"primaryKey,omitempty"`

	//SecondaryKey: The current secondary key that clients can use to authenticate
	//with Redis cache.
	SecondaryKey *string `json:"secondaryKey,omitempty"`
}

//Generated from:
type RedisLinkedServer_Status struct {

	//Id: Linked server Id.
	Id *string `json:"id,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2017-10-01/Microsoft.Cache.json#/definitions/Sku
type Sku struct {

	// +kubebuilder:validation:Required
	//Capacity: The size of the Redis cache to deploy. Valid values: for C
	//(Basic/Standard) family (0, 1, 2, 3, 4, 5, 6), for P (Premium) family (1, 2, 3,
	//4).
	Capacity int `json:"capacity"`

	// +kubebuilder:validation:Required
	//Family: The SKU family to use. Valid values: (C, P). (C = Basic/Standard, P =
	//Premium).
	Family SkuFamily `json:"family"`

	// +kubebuilder:validation:Required
	//Name: The type of Redis cache to deploy. Valid values: (Basic, Standard,
	//Premium).
	Name SkuName `json:"name"`
}

//Generated from:
type Sku_Status struct {

	// +kubebuilder:validation:Required
	//Capacity: The size of the Redis cache to deploy. Valid values: for C
	//(Basic/Standard) family (0, 1, 2, 3, 4, 5, 6), for P (Premium) family (1, 2, 3,
	//4).
	Capacity int `json:"capacity"`

	// +kubebuilder:validation:Required
	//Family: The SKU family to use. Valid values: (C, P). (C = Basic/Standard, P =
	//Premium).
	Family SkuStatusFamily `json:"family"`

	// +kubebuilder:validation:Required
	//Name: The type of Redis cache to deploy. Valid values: (Basic, Standard, Premium)
	Name SkuStatusName `json:"name"`
}

// +kubebuilder:validation:Enum={"C","P"}
type SkuFamily string

const (
	SkuFamilyC = SkuFamily("C")
	SkuFamilyP = SkuFamily("P")
)

// +kubebuilder:validation:Enum={"Basic","Premium","Standard"}
type SkuName string

const (
	SkuNameBasic    = SkuName("Basic")
	SkuNamePremium  = SkuName("Premium")
	SkuNameStandard = SkuName("Standard")
)

// +kubebuilder:validation:Enum={"C","P"}
type SkuStatusFamily string

const (
	SkuStatusFamilyC = SkuStatusFamily("C")
	SkuStatusFamilyP = SkuStatusFamily("P")
)

// +kubebuilder:validation:Enum={"Basic","Premium","Standard"}
type SkuStatusName string

const (
	SkuStatusNameBasic    = SkuStatusName("Basic")
	SkuStatusNamePremium  = SkuStatusName("Premium")
	SkuStatusNameStandard = SkuStatusName("Standard")
)

func init() {
	SchemeBuilder.Register(&Redis{}, &RedisList{})
}