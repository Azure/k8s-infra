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
type ExpressRouteGatewaysExpressRouteConnections struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ExpressRouteGatewaysExpressRouteConnections_Spec `json:"spec,omitempty"`
	Status            ExpressRouteConnection_Status                    `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ExpressRouteGatewaysExpressRouteConnectionsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExpressRouteGatewaysExpressRouteConnections `json:"items"`
}

//Generated from:
type ExpressRouteConnection_Status struct {

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	// +kubebuilder:validation:Required
	//Name: The name of the resource.
	Name string `json:"name"`

	//Properties: Properties of the express route connection.
	Properties *ExpressRouteConnectionProperties_Status `json:"properties,omitempty"`
}

type ExpressRouteGatewaysExpressRouteConnections_Spec struct {
	ForProvider ExpressRouteGatewaysExpressRouteConnectionsParameters `json:"forProvider"`
}

//Generated from:
type ExpressRouteConnectionProperties_Status struct {

	//AuthorizationKey: Authorization key to establish the connection.
	AuthorizationKey *string `json:"authorizationKey,omitempty"`

	//EnableInternetSecurity: Enable internet security.
	EnableInternetSecurity *bool `json:"enableInternetSecurity,omitempty"`

	// +kubebuilder:validation:Required
	//ExpressRouteCircuitPeering: The ExpressRoute circuit peering.
	ExpressRouteCircuitPeering ExpressRouteCircuitPeeringId_Status `json:"expressRouteCircuitPeering"`

	//ProvisioningState: The provisioning state of the express route connection
	//resource.
	ProvisioningState *ProvisioningState_Status `json:"provisioningState,omitempty"`

	//RoutingConfiguration: The Routing Configuration indicating the associated and
	//propagated route tables on this connection.
	RoutingConfiguration *RoutingConfiguration_Status `json:"routingConfiguration,omitempty"`

	//RoutingWeight: The routing weight associated to the connection.
	RoutingWeight *int `json:"routingWeight,omitempty"`
}

type ExpressRouteGatewaysExpressRouteConnectionsParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion ExpressRouteGatewaysExpressRouteConnectionsSpecApiVersion `json:"apiVersion"`
	Comments   *string                                                   `json:"comments,omitempty"`

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
	//Properties: Properties of the express route connection.
	Properties ExpressRouteConnectionProperties `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type ExpressRouteGatewaysExpressRouteConnectionsSpecType `json:"type"`
}

//Generated from:
type ExpressRouteCircuitPeeringId_Status struct {

	//Id: The ID of the ExpressRoute circuit peering.
	Id *string `json:"id,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/ExpressRouteConnectionProperties
type ExpressRouteConnectionProperties struct {

	//AuthorizationKey: Authorization key to establish the connection.
	AuthorizationKey *string `json:"authorizationKey,omitempty"`

	//EnableInternetSecurity: Enable internet security.
	EnableInternetSecurity *bool `json:"enableInternetSecurity,omitempty"`

	// +kubebuilder:validation:Required
	//ExpressRouteCircuitPeering: The ExpressRoute circuit peering.
	ExpressRouteCircuitPeering SubResource `json:"expressRouteCircuitPeering"`

	//RoutingConfiguration: The Routing Configuration indicating the associated and
	//propagated route tables on this connection.
	RoutingConfiguration *RoutingConfiguration `json:"routingConfiguration,omitempty"`

	//RoutingWeight: The routing weight associated to the connection.
	RoutingWeight *int `json:"routingWeight,omitempty"`
}

// +kubebuilder:validation:Enum={"2020-05-01"}
type ExpressRouteGatewaysExpressRouteConnectionsSpecApiVersion string

const ExpressRouteGatewaysExpressRouteConnectionsSpecApiVersion20200501 = ExpressRouteGatewaysExpressRouteConnectionsSpecApiVersion("2020-05-01")

// +kubebuilder:validation:Enum={"Microsoft.Network/expressRouteGateways/expressRouteConnections"}
type ExpressRouteGatewaysExpressRouteConnectionsSpecType string

const ExpressRouteGatewaysExpressRouteConnectionsSpecTypeMicrosoftNetworkExpressRouteGatewaysExpressRouteConnections = ExpressRouteGatewaysExpressRouteConnectionsSpecType("Microsoft.Network/expressRouteGateways/expressRouteConnections")

//Generated from:
type RoutingConfiguration_Status struct {

	//AssociatedRouteTable: The resource id RouteTable associated with this
	//RoutingConfiguration.
	AssociatedRouteTable *SubResource_Status `json:"associatedRouteTable,omitempty"`

	//PropagatedRouteTables: The list of RouteTables to advertise the routes to.
	PropagatedRouteTables *PropagatedRouteTable_Status `json:"propagatedRouteTables,omitempty"`

	//VnetRoutes: List of routes that control routing from VirtualHub into a virtual
	//network connection.
	VnetRoutes *VnetRoute_Status `json:"vnetRoutes,omitempty"`
}

//Generated from:
type PropagatedRouteTable_Status struct {

	//Ids: The list of resource ids of all the RouteTables.
	Ids []SubResource_Status `json:"ids,omitempty"`

	//Labels: The list of labels.
	Labels []string `json:"labels,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/RoutingConfiguration
type RoutingConfiguration struct {

	//AssociatedRouteTable: The resource id RouteTable associated with this
	//RoutingConfiguration.
	AssociatedRouteTable *SubResource `json:"associatedRouteTable,omitempty"`

	//PropagatedRouteTables: The list of RouteTables to advertise the routes to.
	PropagatedRouteTables *PropagatedRouteTable `json:"propagatedRouteTables,omitempty"`

	//VnetRoutes: List of routes that control routing from VirtualHub into a virtual
	//network connection.
	VnetRoutes *VnetRoute `json:"vnetRoutes,omitempty"`
}

//Generated from:
type VnetRoute_Status struct {

	//StaticRoutes: List of all Static Routes.
	StaticRoutes []StaticRoute_Status `json:"staticRoutes,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/PropagatedRouteTable
type PropagatedRouteTable struct {

	//Ids: The list of resource ids of all the RouteTables.
	Ids []SubResource `json:"ids,omitempty"`

	//Labels: The list of labels.
	Labels []string `json:"labels,omitempty"`
}

//Generated from:
type StaticRoute_Status struct {

	//AddressPrefixes: List of all address prefixes.
	AddressPrefixes []string `json:"addressPrefixes,omitempty"`

	//Name: The name of the StaticRoute that is unique within a VnetRoute.
	Name *string `json:"name,omitempty"`

	//NextHopIpAddress: The ip address of the next hop.
	NextHopIpAddress *string `json:"nextHopIpAddress,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/VnetRoute
type VnetRoute struct {

	//StaticRoutes: List of all Static Routes.
	StaticRoutes []StaticRoute `json:"staticRoutes,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2020-05-01/Microsoft.Network.json#/definitions/StaticRoute
type StaticRoute struct {

	//AddressPrefixes: List of all address prefixes.
	AddressPrefixes []string `json:"addressPrefixes,omitempty"`

	//Name: The name of the StaticRoute that is unique within a VnetRoute.
	Name *string `json:"name,omitempty"`

	//NextHopIpAddress: The ip address of the next hop.
	NextHopIpAddress *string `json:"nextHopIpAddress,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ExpressRouteGatewaysExpressRouteConnections{}, &ExpressRouteGatewaysExpressRouteConnectionsList{})
}
