
/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

// Code generated by k8s-infra-gen. DO NOT EDIT.
// Generator version:  (tree is )

// Package v20200501 contains API Schema definitions for the microsoft.network v20200501 API group
// +kubebuilder:object:generate=true
// All object properties are optional by default, this will be overridden when needed:
// +kubebuilder:validation:Optional
// +groupName=microsoft.network.infra.azure.com
package v20200501

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "microsoft.network.infra.azure.com", Version: "v20200501"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme

	localSchemeBuilder = SchemeBuilder.SchemeBuilder
)
