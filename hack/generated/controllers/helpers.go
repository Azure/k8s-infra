/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

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

// TODO: Remove this when we have proper AzureName defaulting on the way in
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

// TODO: is there some helper that does this?
func StringPtrToString(s *string) string {
	if s == nil {
		return "nil"
	}

	return *s
}