package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Azure/k8s-infra/hack/generated/pkg/armclient"
	"github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"time"
)

type ControllerResourceProvisioningState string

const (
	SucceededControllerResourceProvisioningState ControllerResourceProvisioningState = "Succeeded"
	FailedControllerResourceProvisioningState    ControllerResourceProvisioningState = "Failed"
	DeletingControllerResourceProvisioningState  ControllerResourceProvisioningState = "Deleting"
	AcceptedControllerResourceProvisioningState  ControllerResourceProvisioningState = "Accepted"
)

type AnnotationKey string

const (
	// PreserveDeploymentAnnotation is the key which tells the applier to keep or delete the deployment
	PreserveDeploymentAnnotation AnnotationKey = "x-preserve-deployment"
)

func IsTerminalProvisioningState(state ControllerResourceProvisioningState) bool {
	return state == SucceededControllerResourceProvisioningState || state == FailedControllerResourceProvisioningState
}

func SpecSignature(metaObject genruntime.MetaObject) (string, error) {
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
		return "", fmt.Errorf("unable to marshal spec of unstructured MetaObject with: %w", err)
	}

	hash := sha256.Sum256(bits)
	return hex.EncodeToString(hash[:]), nil
}

func CreateDeploymentName() (string, error) {
	// no status yet, so start provisioning
	deploymentUUID, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	deploymentName := fmt.Sprintf("%s_%d_%s", "k8s", time.Now().Unix(), deploymentUUID.String())
	return deploymentName, nil
}

func ResourceSpecToArmResourceSpec(gr *GenericReconciler, metaObject genruntime.MetaObject) (string, genruntime.ArmResourceSpec, error) {

	metaObjReflector := reflect.Indirect(reflect.ValueOf(metaObject))
	if !metaObjReflector.IsValid() {
		return "", nil, errors.Errorf("couldn't indirect %T", metaObject)
	}

	specField := metaObjReflector.FieldByName("Spec")
	if !specField.IsValid() {
		return "", nil, errors.Errorf("couldn't find spec field on type %T", metaObject)
	}

	// Spec fields are values, we want a ptr
	specFieldPtr := reflect.New(specField.Type())
	specFieldPtr.Elem().Set(specField)

	// TODO: how to check that this doesn't fail
	spec := specFieldPtr.Interface()

	armTransformer, ok := spec.(genruntime.ArmTransformer)
	if !ok {
		return "", nil, errors.Errorf("spec was of type %T which doesn't implement genruntime.ArmTransformer", spec)
	}

	resourceGroupName, azureName, err := GetFullAzureNameAndResourceGroup(metaObject, gr)
	if err != nil {
		return "", nil, err
	}

	armSpec, err := armTransformer.ConvertToArm(azureName)
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to transform resource %s to ARM", metaObject.GetName())
	}

	typedArmSpec, ok := armSpec.(genruntime.ArmResourceSpec)
	if !ok {
		return "", nil, errors.Errorf("failed to cast armSpec of type %T to genruntime.ArmResourceSpec", armSpec)
	}

	return resourceGroupName, typedArmSpec, nil
}

func ResourceSpecToDeployment(gr *GenericReconciler, metaObject genruntime.MetaObject) (*armclient.Deployment, error) {
	resourceGroupName, typedArmSpec, err := ResourceSpecToArmResourceSpec(gr, metaObject)
	if err != nil {
		return nil, err
	}

	// TODO: get other deployment details from status and avoid creating a new deployment
	deploymentId, deploymentIdOk := metaObject.GetAnnotations()[DeploymentIdAnnotation]
	deploymentName, deploymentNameOk := metaObject.GetAnnotations()[DeploymentNameAnnotation]
	if deploymentIdOk != deploymentNameOk {
		return nil, errors.Errorf(
			"deploymentIdOk: %t, deploymentNameOk: %t expected to match, but didn't",
			deploymentIdOk,
			deploymentNameOk)
	}

	var deployment *armclient.Deployment
	if deploymentIdOk && deploymentNameOk {
		deployment = gr.Applier.NewDeployment(resourceGroupName, deploymentName, typedArmSpec)
		deployment.Id = deploymentId
	} else {
		deploymentName, err := CreateDeploymentName()
		if err != nil {
			return nil, err
		}
		deployment = gr.Applier.NewDeployment(resourceGroupName, deploymentName, typedArmSpec)
	}
	return deployment, nil
}

// TODO: Remove this when we have proper AzureName defaulting on hte way in
func GetAzureName(r genruntime.MetaObject) string {
	if r.AzureName() == "" {
		return r.GetName()
	}

	return r.AzureName()
}

func GetFullAzureNameAndResourceGroup(r genruntime.MetaObject, gr *GenericReconciler) (string, string, error) {
	owner := r.Owner()

	if r.GetObjectKind().GroupVersionKind().Kind == "ResourceGroup" {
		return GetAzureName(r), "", nil
	}

	if owner != nil {
		var ownerGvk schema.GroupVersionKind
		found := false
		for gvk, _ := range gr.Scheme.AllKnownTypes() {
			if gvk.Group == owner.Group && gvk.Kind == owner.Kind {
				if !found {
					ownerGvk = gvk
					found = true
				} else {
					return "", "", errors.Errorf("owner group: %s, kind: %s has multiple possible schemes registered", owner.Group, owner.Kind)
				}
			}
		}

		// TODO: This is a hack for now since we don't have an RG type yet
		if owner.Kind == "ResourceGroup" {
			return owner.Name, GetAzureName(r), nil
		}

		// TODO: We could do this on launch probably since we can check based on the AllKnownTypes() collection
		if !found {
			return "", "", errors.Errorf("couldn't find registered scheme for owner %+v", owner)
		}

		ownerNamespacedName := types.NamespacedName{
			Namespace: r.GetNamespace(), // TODO: Assumption that resource ownership is not cross namespace
			Name: owner.Name,
		}

		ownerObj, err := gr.GetObject(ownerNamespacedName, ownerGvk)
		if err != nil {
			return "", "", errors.Wrapf(err, "couldn't find owner %s of %s", owner.Name, r.GetName())
		}

		ownerMeta, ok := ownerObj.(genruntime.MetaObject)
		if !ok {
			return "", "", errors.Errorf("owner %s (%s) was not of type genruntime.MetaObject", ownerNamespacedName, ownerGvk)
		}

		rgName, ownerName, err := GetFullAzureNameAndResourceGroup(ownerMeta, gr)
		combinedAzureName := GetAzureName(r)
		if ownerName != "" {
			combinedAzureName = genruntime.CombineArmNames(ownerName, r.AzureName())
		}
		return rgName, combinedAzureName, nil
	}

	panic(
		fmt.Sprintf(
			"Can't GetOwnerAndResourceGroupDetails from %s (kind: %s), which has no owner but is not a ResourceGroup",
			r.GetName(),
			r.GetObjectKind().GroupVersionKind()))
}

