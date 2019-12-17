/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	protov1alpha1 "github.com/Azure/k8s-infra/api/v1alpha1"
)

// ResourceGroupReconciler reconciles a ResourceGroup object
type ResourceGroupReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=proto.infra.azure.com,resources=resourcegroups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=proto.infra.azure.com,resources=resourcegroups/status,verbs=get;update;patch
func (r *ResourceGroupReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = r.Log.WithValues("resourcegroup", req.NamespacedName)

	// your logic here
	var rg protov1alpha1.ResourceGroup
	if err := r.Get(ctx, req.NamespacedName, &rg); err != nil {
		return ctrl.Result{}, err
	}

	if rg.DeletionTimestamp.IsZero() {
		// delete me
		return ctrl.Result{}, nil
	}

	if rg.Spec.ID == "" {
		// create me
		return ctrl.Result{}, nil
	}

	// update properties, perhaps only tags

	return ctrl.Result{}, nil
}

func (r *ResourceGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&protov1alpha1.ResourceGroup{}).
		Complete(r)
}
