/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package armresourceresolver

import (
	"context"
	"fmt"
	"github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
	"github.com/Azure/k8s-infra/hack/generated/pkg/util/kubeclient"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

type Resolver struct {
	Client *kubeclient.Client
}

func NewResolver(client *kubeclient.Client) *Resolver {
	return &Resolver{
		Client: client,
	}
}

// GetFullAzureNameAndResourceGroup gets the full name for use in creating a resource. This name includes
// the full "path" to the resource being deployed. For example, a Virtual Network Subnet's name might be:
// "myvnet/mysubnet"
func (r *Resolver) GetFullAzureNameAndResourceGroup(ctx context.Context, obj genruntime.MetaObject) (string, string, error) {
	owner := obj.Owner()

	if obj.GetObjectKind().GroupVersionKind().Kind == "ResourceGroup" {
		return getAzureName(obj), "", nil
	}

	if owner != nil {
		var ownerGvk schema.GroupVersionKind
		found := false
		for gvk := range r.Client.Scheme.AllKnownTypes() {
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
			return owner.Name, getAzureName(obj), nil
		}

		// TODO: We could do this on launch probably since we can check based on the AllKnownTypes() collection
		if !found {
			return "", "", errors.Errorf("couldn't find registered scheme for owner %+v", owner)
		}

		ownerNamespacedName := types.NamespacedName{
			Namespace: obj.GetNamespace(), // TODO: Assumption that resource ownership is not cross namespace
			Name:      owner.Name,
		}

		ownerObj, err := r.Client.GetObject(ctx, ownerNamespacedName, ownerGvk)
		if err != nil {
			return "", "", errors.Wrapf(err, "couldn't find owner %s of %s", owner.Name, obj.GetName())
		}

		ownerMeta, ok := ownerObj.(genruntime.MetaObject)
		if !ok {
			return "", "", errors.Errorf("owner %s (%s) was not of type genruntime.MetaObject", ownerNamespacedName, ownerGvk)
		}

		rgName, ownerName, err := r.GetFullAzureNameAndResourceGroup(ctx, ownerMeta)
		if err != nil {
			return "", "", errors.Wrapf(err, "failed to get full Azure name and resource group for %s", ownerNamespacedName)
		}
		combinedAzureName := getAzureName(obj)
		if ownerName != "" {
			combinedAzureName = genruntime.CombineArmNames(ownerName, combinedAzureName)
		}
		return rgName, combinedAzureName, nil
	}

	// TODO: Do we really want to panic here?
	panic(
		fmt.Sprintf(
			"Can't GetOwnerAndResourceGroupDetails from %s (kind: %s), which has no owner but is not a ResourceGroup",
			obj.GetName(),
			obj.GetObjectKind().GroupVersionKind()))
}

// TODO: Remove this when we have proper AzureName defaulting on the way in
// getAzureName returns the specified AzureName, or else the name of the Kubernetes resource
func getAzureName(r genruntime.MetaObject) string {
	if r.AzureName() == "" {
		return r.GetName()
	}

	return r.AzureName()
}
