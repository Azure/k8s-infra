// TODO: Can we change this?
/*
Copyright 2017 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package patch

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO: This class is use-once, which seems wrong to me?
// Helper is a utility for ensuring the proper Patching of resources
// and their status
type Helper struct {
	client client.Client
}

// NewHelper returns an initialized Helper
func NewHelper(c client.Client) *Helper {
	return &Helper{
		client: c,
	}
}

// Patch will attempt to patch the given resource and its status
func (h *Helper) Patch(ctx context.Context, before runtime.Object, after runtime.Object) error {
	err := validateResource(before)
	if err != nil {
		return err
	}
	err = validateResource(after)
	if err != nil {
		return err
	}

	resourcePatch := client.MergeFrom(before.DeepCopyObject())
	statusPatch := client.MergeFrom(before.DeepCopyObject())

	// Convert the resource to unstructured to compare against our before copy.
	beforeUnstructured, err := resourceToUnstructured(before)
	if err != nil {
		return err
	}
	afterUnstructured, err := resourceToUnstructured(after)
	if err != nil {
		return err
	}

	beforeStatus, beforeHasStatus, err := extractStatus(beforeUnstructured)
	if err != nil {
		return err
	}
	afterStatus, afterHasStatus, err := extractStatus(afterUnstructured)
	if err != nil {
		return err
	}

	var errs []error

	if !reflect.DeepEqual(beforeUnstructured, afterUnstructured) {
		// only issue a Patch if the before and after resources (minus status) differ
		if err := h.client.Patch(ctx, after.DeepCopyObject(), resourcePatch); err != nil {
			errs = append(errs, err)
		}
	}

	if (beforeHasStatus || afterHasStatus) && !reflect.DeepEqual(beforeStatus, afterStatus) {
		// only issue a Status Patch if the resource has a status and the beforeStatus
		// and afterStatus copies differ
		if err := h.client.Status().Patch(ctx, after.DeepCopyObject(), statusPatch); err != nil {
			errs = append(errs, err)
		}
	}

	return kerrors.NewAggregate(errs)
}

func validateResource(r runtime.Object) error {
	if r == nil {
		return errors.Errorf("expected non-nil resource")
	}

	return nil
}

func resourceToUnstructured(r runtime.Object) (map[string]interface{}, error) {
	// If the object is already unstructured, we need to perform a deepcopy first
	// because the `DefaultUnstructuredConverter.ToUnstructured` function returns
	// the underlying unstructured object map without making a copy.
	if _, ok := r.(runtime.Unstructured); ok {
		r = r.DeepCopyObject()
	}

	// Convert the resource to unstructured for easier comparison later.
	return runtime.DefaultUnstructuredConverter.ToUnstructured(r)
}

// Note that this call modifies u by removing status!
func extractStatus(u map[string]interface{}) (interface{}, bool, error) {
	hasStatus := false
	// attempt to extract the status from the resource for easier comparison later
	status, ok, err := unstructured.NestedFieldCopy(u, "status")
	if err != nil {
		return nil, false, err
	}
	if ok {
		hasStatus = true
		// if the resource contains a status remove it from our unstructured copy
		// to avoid unnecessary patching later
		unstructured.RemoveNestedField(u, "status")
	}

	return status, hasStatus, nil
}
