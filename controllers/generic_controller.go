/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	protov1alpha1 "github.com/Azure/k8s-infra/api/v1alpha1"
)

// GenericReconciler reconciles any generic Azure object
type GenericReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=proto.infra.azure.com,resources=virtualnetwork,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=proto.infra.azure.com,resources=virtualnetwork/status,verbs=get;update;patch
func (gr *GenericReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = gr.Log.WithValues("genericReconciler", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (gr *GenericReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&protov1alpha1.VirtualNetwork{}).
		Complete(gr)
}
