/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package genruntime

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"strconv"
)

const (
	// TODO: Delete these later in favor of something in status?
	DeploymentIdAnnotation   = "deployment-id.infra.azure.com"
	DeploymentNameAnnotation = "deployment-name.infra.azure.com"
	ResourceStateAnnotation  = "resource-state.infra.azure.com"
	ResourceIdAnnotation     = "resource-id.infra.azure.com"
	ResourceErrorAnnotation  = "resource-error.infra.azure.com"
	ResourceSigAnnotationKey = "resource-sig.infra.azure.com"
	// PreserveDeploymentAnnotation is the key which tells the applier to keep or delete the deployment
	PreserveDeploymentAnnotation = "x-preserve-deployment"
)

// MetaObject represents an arbitrary k8s-infra custom resource
type MetaObject interface {
	runtime.Object
	metav1.Object
	KubernetesResource
}

// KubernetesResource is an Azure resource. This interface contains the common set of
// methods that apply to all k8s-infra resources.
type KubernetesResource interface {
	// Owner returns the ResourceReference of the owner, or nil if there is no owner
	Owner() *ResourceReference

	// TODO: I think we need this?
	// KnownOwner() *KnownResourceReference

	// AzureName returns the Azure name of the resource
	AzureName() string

	// TODO: GetAPIVersion here?

	// TODO: I think we need this
	// SetStatus(status interface{})
}

func GetResourceProvisioningStateOrDefault(metaObj MetaObject) string {
	return metaObj.GetAnnotations()[ResourceStateAnnotation]
}

func GetDeploymentId(metaObj MetaObject) (string, bool) {
	id, ok := metaObj.GetAnnotations()[DeploymentIdAnnotation]
	return id, ok
}

func GetDeploymentIdOrDefault(metaObj MetaObject) string {
	id, ok := GetDeploymentId(metaObj)
	if !ok {
		return ""
	}

	return id
}

func SetDeploymentId(metaObj MetaObject, id string) {
	addAnnotation(metaObj, DeploymentIdAnnotation, id)
}

func GetDeploymentName(metaObj MetaObject) (string, bool) {
	id, ok := metaObj.GetAnnotations()[DeploymentNameAnnotation]
	return id, ok
}

func GetDeploymentNameOrDefault(metaObj MetaObject) string {
	id, ok := GetDeploymentName(metaObj)
	if !ok {
		return ""
	}

	return id
}

func SetDeploymentName(metaObj MetaObject, name string) {
	addAnnotation(metaObj, DeploymentNameAnnotation, name)
}

func GetShouldPreserveDeployment(metaObj MetaObject) bool {
	preserveDeploymentString, ok := metaObj.GetAnnotations()[PreserveDeploymentAnnotation]
	if !ok {
		return false
	}

	preserveDeployment, err := strconv.ParseBool(preserveDeploymentString)
	// Anything other than an error is assumed to be false...
	// TODO: Would we rather have any usage of this key imply true (regardless of value?)
	if err != nil {
		// TODO: Log here
		return false
	}

	return preserveDeployment
}

// TODO: Remove this when we have status
func GetResourceId(metaObj MetaObject) string {
	id, ok := metaObj.GetAnnotations()[ResourceIdAnnotation]
	if !ok {
		return ""
	}

	return id
}

func SetResourceId(metaObj MetaObject, id string) {
	addAnnotation(metaObj, ResourceIdAnnotation, id)
}

func SetResourceState(metaObj MetaObject, state string) { // TODO: state should probably be an enum
	addAnnotation(metaObj, ResourceStateAnnotation, state)
}

func SetResourceError(metaObj MetaObject, error string) {
	addAnnotation(metaObj, ResourceErrorAnnotation, error)
}

func SetResourceSignature(metaObj MetaObject, sig string) {
	addAnnotation(metaObj, ResourceSigAnnotationKey, sig)
}

func HasResourceSpecHashChanged(metaObj MetaObject) (bool, error) {
	oldSig, exists := metaObj.GetAnnotations()[ResourceSigAnnotationKey]
	if !exists {
		// signature does not exist, so yes, it has changed
		return true, nil
	}

	newSig, err := SpecSignature(metaObj)
	if err != nil {
		return false, err
	}
	// check if the last signature matches the new signature
	return oldSig != newSig, nil
}

func addAnnotation(metaObj MetaObject, k string, v string) {
	annotations := metaObj.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	// I think this is the behavior we want...
	if v == "" {
		delete(annotations, k)
	} else {
		annotations[k] = v
	}
	metaObj.SetAnnotations(annotations)
}

// SpecSignature calculates the hash of a spec. This can be used to compare specs and determine
// if there has been a change
func SpecSignature(metaObject MetaObject) (string, error) {
	// Convert the resource to unstructured for easier comparison later.
	unObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(metaObject)
	if err != nil {
		return "", err
	}

	spec, ok, err := unstructured.NestedMap(unObj, "spec")
	if err != nil {
		return "", err
	}

	if !ok {
		return "", errors.New("unable to find spec within unstructured MetaObject")
	}

	bits, err := json.Marshal(spec)
	if err != nil {
		return "", errors.Wrap(err, "unable to marshal spec of unstructured MetaObject")
	}

	hash := sha256.Sum256(bits)
	return hex.EncodeToString(hash[:]), nil
}

// ArmResourceSpec is an ARM resource specification. This interface contains
// methods to access properties common to all ARM Resource Specs. An Azure
// Deployment is made of these.
type ArmResourceSpec interface {
	GetApiVersion() string

	GetType() string

	GetName() string
}

// ArmResourceStatus is an ARM resource status
type ArmResourceStatus interface {
	// GetId() string
}

type ArmResource interface {
	Spec() ArmResourceSpec
	Status() ArmResourceStatus

	GetId() string // TODO: ?
}

func NewArmResource(spec ArmResourceSpec, status ArmResourceStatus, id string) ArmResource {
	return &armResourceImpl{
		spec:   spec,
		status: status,
		Id:     id,
	}
}

type armResourceImpl struct {
	spec   ArmResourceSpec
	status ArmResourceStatus
	Id     string
}

var _ ArmResource = &armResourceImpl{}

func (resource *armResourceImpl) Spec() ArmResourceSpec {
	return resource.spec
}

func (resource *armResourceImpl) Status() ArmResourceStatus {
	return resource.status
}

func (resource *armResourceImpl) GetId() string {
	return resource.Id
}
