/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package kubeclient

import (
	"context"
	"github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
	"github.com/Azure/k8s-infra/hack/generated/pkg/util/patch"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
	Client      client.Client
	Scheme      *runtime.Scheme
	patchHelper *patch.Helper
}

func NewClient(
	client client.Client,
	scheme *runtime.Scheme) *Client {

	return &Client{
		Client:      client,
		Scheme:      scheme,
		patchHelper: patch.NewHelper(client),
	}
}

func (k *Client) GetObject(ctx context.Context, namespacedName types.NamespacedName, gvk schema.GroupVersionKind) (runtime.Object, error) {
	obj, err := k.Scheme.New(gvk)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create object from gvk %+v with", gvk)
	}

	if err := k.Client.Get(ctx, namespacedName, obj); err != nil {
		// TODO: Would we rather the caller do this?
		if apierrors.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return obj, nil
}

func (k *Client) PatchHelper(
	ctx context.Context,
	obj genruntime.MetaObject,
	mutator func(context.Context, genruntime.MetaObject) error) error {

	before := obj.DeepCopyObject()

	if err := mutator(ctx, obj); err != nil {
		return err
	}

	if err := k.patchHelper.Patch(ctx, before, obj); err != nil {
		return errors.Wrap(err, "patchHelper patch failed")
	}

	// fill resourcer with patched updates since patch will copy resourcer
	return k.Client.Get(ctx, client.ObjectKey{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}, obj)
}
