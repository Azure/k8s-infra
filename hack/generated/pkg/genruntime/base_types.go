package genruntime

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type MetaObject interface {
	runtime.Object
	metav1.Object
	KubernetesResource
}

type KubernetesResource interface {
	// Owner returns the ResourceReference of the owner, or nil if there is no owner
	Owner() *ResourceReference

	// AzureName returns the Azure name of the resource
	AzureName() string
}
