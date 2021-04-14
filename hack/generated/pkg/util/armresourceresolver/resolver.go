/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package armresourceresolver

import (
	"context"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
	"github.com/Azure/k8s-infra/hack/generated/pkg/util/kubeclient"
)

type Resolver struct {
	client                   *kubeclient.Client
	reconciledResourceLookup map[schema.GroupKind]schema.GroupVersionKind
}

func NewResolver(client *kubeclient.Client, reconciledResourceLookup map[schema.GroupKind]schema.GroupVersionKind) *Resolver {
	return &Resolver{
		client:                   client,
		reconciledResourceLookup: reconciledResourceLookup,
	}
}

// TODO: I'm not sure that owner has to be as special as it's being made here
// GetReferenceARMID gets a references ARM ID. If the reference is just pointing to an ARM resource then the ARMID is returned.
// If the reference is pointing to a Kubernetes resource, that resource is looked up and its ARM ID is computed.
func (r *Resolver) GetReferenceARMID(ctx context.Context, ref genruntime.ResourceReference) (string, error) {
	// TODO: is there a cleaner way to make these checks? Maybe I want to transform the flat type to a hierarchical
	// TODO: for handling internally?
	if ref.IsDirectARMReference() {
		return ref.ARMID, nil
	}

	// TODO: Technically don't need this check as it's handled by ResolveReference
	if !ref.IsKubernetesReference() {
		return "", errors.Errorf("ref %s is neither ARM or Kubernetes reference", ref)
	}

	obj, err := r.ResolveReference(ctx, ref)
	if err != nil {
		return "", err
	}

	hierarchy, err := r.ResolveResourceHierarchy(ctx, obj)
	if err != nil {
		return "", err
	}

	return hierarchy.FullAzureName(), nil
}

// TODO: Possibly can make this "private"
// ResolveResourceHierarchy gets the resource hierarchy for a given resource. The result is a slice of
// resources, with the uppermost parent at position 0 and the resource itself at position len(slice)-1
func (r *Resolver) ResolveResourceHierarchy(ctx context.Context, obj genruntime.MetaObject) (ResourceHierarchy, error) {

	owner := obj.Owner()
	if owner == nil {
		return ResourceHierarchy{obj}, nil
	}

	ownerMeta, err := r.ResolveOwner(ctx, obj)
	if err != nil {
		return nil, err
	}

	owners, err := r.ResolveResourceHierarchy(ctx, ownerMeta)
	if err != nil {
		return nil, errors.Wrapf(err, "getting owners for %s", ownerMeta.GetName())
	}

	return append(owners, obj), nil
}

// ResolveReference resolves a reference, or returns an error if the reference is not pointing to a KubernetesResource
func (r *Resolver) ResolveReference(ctx context.Context, ref genruntime.ResourceReference) (genruntime.MetaObject, error) {
	refGVK, err := r.findGVK(ref)
	if err != nil {
		return nil, err
	}

	refNamespacedName := types.NamespacedName{
		Namespace: ref.Namespace,
		Name:      ref.Name,
	}

	refObj, err := r.client.GetObject(ctx, refNamespacedName, refGVK)
	if err != nil {
		if apierrors.IsNotFound(err) {
			err := NewReferenceNotFoundError(refNamespacedName, err)
			return nil, errors.WithStack(err)
		}

		return nil, errors.Wrapf(err, "couldn't resolve reference %s", ref.String())
	}

	metaObj, ok := refObj.(genruntime.MetaObject)
	if !ok {
		return nil, errors.Errorf("reference %s (%s) was not of type genruntime.MetaObject", refNamespacedName, refGVK)
	}

	return metaObj, nil
}

// ResolveOwner returns the MetaObject for the given resources owner. If the resource is supposed to have
// an owner but doesn't, this returns an ReferenceNotFound error. If the resource is not supposed
// to have an owner (for example, ResourceGroup), returns nil.
func (r *Resolver) ResolveOwner(ctx context.Context, obj genruntime.MetaObject) (genruntime.MetaObject, error) {
	owner := obj.Owner()

	if owner == nil {
		return nil, nil
	}

	ownerMeta, err := r.ResolveReference(ctx, *owner)
	if err != nil {
		return nil, err
	}

	return ownerMeta, nil
}

func (r *Resolver) findGVK(ref genruntime.ResourceReference) (schema.GroupVersionKind, error) {
	var ownerGvk schema.GroupVersionKind

	if !ref.IsKubernetesReference() {
		return ownerGvk, errors.Errorf("reference %s is not pointing to a Kubernetes resource", ref)
	}

	groupKind := schema.GroupKind{Group: ref.Group, Kind: ref.Kind}
	gvk, ok := r.reconciledResourceLookup[groupKind]
	if !ok {
		return ownerGvk, errors.Errorf("group: %q, kind: %q was not in reconciledResourceLookup", ref.Group, ref.Kind)
	}

	return gvk, nil
}
