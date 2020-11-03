// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
// Code generated by k8s-infra-gen. DO NOT EDIT.
package v20171001

import (
	"github.com/Azure/k8s-infra/hack/crossplane/apis/deploymenttemplate/v20150101"
	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
type RedisPatchSchedules struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RedisPatchSchedules_Spec  `json:"spec,omitempty"`
	Status            RedisPatchSchedule_Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type RedisPatchSchedulesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisPatchSchedules `json:"items"`
}

type RedisPatchSchedule_Status struct {
	v1alpha1.ResourceStatus `json:",inline"`
	AtProvider              RedisPatchSchedulesObservation `json:"atProvider"`
}

type RedisPatchSchedules_Spec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           RedisPatchSchedulesParameters `json:"forProvider"`
}

type RedisPatchSchedulesObservation struct {

	//Id: Resource ID.
	Id *string `json:"id,omitempty"`

	//Name: Resource name.
	Name *string `json:"name,omitempty"`

	// +kubebuilder:validation:Required
	//Properties: List of patch schedules for a Redis cache.
	Properties ScheduleEntries_Status `json:"properties"`

	//Type: Resource type.
	Type *string `json:"type,omitempty"`
}

type RedisPatchSchedulesParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion RedisPatchSchedulesSpecApiVersion `json:"apiVersion"`
	Comments   *string                           `json:"comments,omitempty"`

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
	//Properties: List of patch schedules for a Redis cache.
	Properties ScheduleEntries `json:"properties"`

	//Scope: Scope for the resource or deployment. Today, this works for two cases: 1)
	//setting the scope for extension resources 2) deploying resources to the tenant
	//scope in non-tenant scope deployments
	Scope *string `json:"scope,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type RedisPatchSchedulesSpecType `json:"type"`
}

// +kubebuilder:validation:Enum={"2017-10-01"}
type RedisPatchSchedulesSpecApiVersion string

const RedisPatchSchedulesSpecApiVersion20171001 = RedisPatchSchedulesSpecApiVersion("2017-10-01")

// +kubebuilder:validation:Enum={"Microsoft.Cache/Redis/patchSchedules"}
type RedisPatchSchedulesSpecType string

const RedisPatchSchedulesSpecTypeMicrosoftCacheRedisPatchSchedules = RedisPatchSchedulesSpecType("Microsoft.Cache/Redis/patchSchedules")

//Generated from: https://schema.management.azure.com/schemas/2017-10-01/Microsoft.Cache.json#/definitions/ScheduleEntries
type ScheduleEntries struct {

	// +kubebuilder:validation:Required
	//ScheduleEntries: List of patch schedules for a Redis cache.
	ScheduleEntries []ScheduleEntry `json:"scheduleEntries"`
}

//Generated from:
type ScheduleEntries_Status struct {

	// +kubebuilder:validation:Required
	//ScheduleEntries: List of patch schedules for a Redis cache.
	ScheduleEntries []ScheduleEntry_Status `json:"scheduleEntries"`
}

//Generated from: https://schema.management.azure.com/schemas/2017-10-01/Microsoft.Cache.json#/definitions/ScheduleEntry
type ScheduleEntry struct {

	// +kubebuilder:validation:Required
	//DayOfWeek: Day of the week when a cache can be patched.
	DayOfWeek ScheduleEntryDayOfWeek `json:"dayOfWeek"`

	//MaintenanceWindow: ISO8601 timespan specifying how much time cache patching can
	//take.
	MaintenanceWindow *string `json:"maintenanceWindow,omitempty"`

	// +kubebuilder:validation:Required
	//StartHourUtc: Start hour after which cache patching can start.
	StartHourUtc int `json:"startHourUtc"`
}

//Generated from:
type ScheduleEntry_Status struct {

	// +kubebuilder:validation:Required
	//DayOfWeek: Day of the week when a cache can be patched.
	DayOfWeek ScheduleEntryStatusDayOfWeek `json:"dayOfWeek"`

	//MaintenanceWindow: ISO8601 timespan specifying how much time cache patching can
	//take.
	MaintenanceWindow *string `json:"maintenanceWindow,omitempty"`

	// +kubebuilder:validation:Required
	//StartHourUtc: Start hour after which cache patching can start.
	StartHourUtc int `json:"startHourUtc"`
}

// +kubebuilder:validation:Enum={"Everyday","Friday","Monday","Saturday","Sunday","Thursday","Tuesday","Wednesday","Weekend"}
type ScheduleEntryDayOfWeek string

const (
	ScheduleEntryDayOfWeekEveryday  = ScheduleEntryDayOfWeek("Everyday")
	ScheduleEntryDayOfWeekFriday    = ScheduleEntryDayOfWeek("Friday")
	ScheduleEntryDayOfWeekMonday    = ScheduleEntryDayOfWeek("Monday")
	ScheduleEntryDayOfWeekSaturday  = ScheduleEntryDayOfWeek("Saturday")
	ScheduleEntryDayOfWeekSunday    = ScheduleEntryDayOfWeek("Sunday")
	ScheduleEntryDayOfWeekThursday  = ScheduleEntryDayOfWeek("Thursday")
	ScheduleEntryDayOfWeekTuesday   = ScheduleEntryDayOfWeek("Tuesday")
	ScheduleEntryDayOfWeekWednesday = ScheduleEntryDayOfWeek("Wednesday")
	ScheduleEntryDayOfWeekWeekend   = ScheduleEntryDayOfWeek("Weekend")
)

// +kubebuilder:validation:Enum={"Everyday","Friday","Monday","Saturday","Sunday","Thursday","Tuesday","Wednesday","Weekend"}
type ScheduleEntryStatusDayOfWeek string

const (
	ScheduleEntryStatusDayOfWeekEveryday  = ScheduleEntryStatusDayOfWeek("Everyday")
	ScheduleEntryStatusDayOfWeekFriday    = ScheduleEntryStatusDayOfWeek("Friday")
	ScheduleEntryStatusDayOfWeekMonday    = ScheduleEntryStatusDayOfWeek("Monday")
	ScheduleEntryStatusDayOfWeekSaturday  = ScheduleEntryStatusDayOfWeek("Saturday")
	ScheduleEntryStatusDayOfWeekSunday    = ScheduleEntryStatusDayOfWeek("Sunday")
	ScheduleEntryStatusDayOfWeekThursday  = ScheduleEntryStatusDayOfWeek("Thursday")
	ScheduleEntryStatusDayOfWeekTuesday   = ScheduleEntryStatusDayOfWeek("Tuesday")
	ScheduleEntryStatusDayOfWeekWednesday = ScheduleEntryStatusDayOfWeek("Wednesday")
	ScheduleEntryStatusDayOfWeekWeekend   = ScheduleEntryStatusDayOfWeek("Weekend")
)

func init() {
	SchemeBuilder.Register(&RedisPatchSchedules{}, &RedisPatchSchedulesList{})
}
