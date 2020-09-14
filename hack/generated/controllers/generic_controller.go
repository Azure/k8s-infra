/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/k8s-infra/hack/generated/pkg/armclient"
	"github.com/Azure/k8s-infra/hack/generated/pkg/util/armresourceresolver"
	"github.com/Azure/k8s-infra/hack/generated/pkg/util/kubeclient"
	"github.com/pkg/errors"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
)

const (
	// ResourceSigAnnotationKey is an annotation key which holds the value of the hash of the spec
	ResourceSigAnnotationKey = "resource-sig.infra.azure.com"

	// TODO: Delete these later in favor of something in status?
	DeploymentIdAnnotation   = "deployment-id.infra.azure.com"
	DeploymentNameAnnotation = "deployment-name.infra.azure.com"
	ResourceStateAnnotation  = "resource-state.infra.azure.com"
	ResourceIdAnnotation     = "resource-id.infra.azure.com"
	ResourceErrorAnnotation  = "resource-error.infra.azure.com"
	// PreserveDeploymentAnnotation is the key which tells the applier to keep or delete the deployment
	PreserveDeploymentAnnotation = "x-preserve-deployment"

	GenericControllerFinalizer = "generated.infra.azure.com/finalizer"
)

// GenericReconciler reconciles resources
type GenericReconciler struct {
	Log              logr.Logger
	ARMClient        armclient.Applier
	KubeClient       *kubeclient.Client
	ResourceResolver *armresourceresolver.Resolver
	Recorder         record.EventRecorder
	Name             string
	GVK              schema.GroupVersionKind
	Controller       controller.Controller
}

func RegisterAll(mgr ctrl.Manager, applier armclient.Applier, objs []runtime.Object, log logr.Logger, options controller.Options) []error {
	var errs []error
	for _, obj := range objs {
		// TODO: What were these for?
		// mgr := mgr
		// obj := obj
		if err := register(mgr, applier, obj, log, options); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func register(mgr ctrl.Manager, applier armclient.Applier, obj runtime.Object, log logr.Logger, options controller.Options) error {
	v, err := conversion.EnforcePtr(obj)
	if err != nil {
		return errors.Wrap(err, "obj was expected to be ptr but was not")
	}

	t := v.Type()
	controllerName := fmt.Sprintf("%sController", t.Name())

	// Use the provided GVK to construct a new runtime object of the desired concrete type.
	gvk, err := apiutil.GVKForObject(obj, mgr.GetScheme())
	if err != nil {
		return errors.Wrapf(err, "couldn't create GVK for obj %T", obj)
	}
	log.V(4).Info("Registering", "GVK", gvk)

	// TODO: Do we need to do this?
	//if err := mgr.GetFieldIndexer().IndexField(obj, "status.id", func(obj runtime.Object) []string {
	//	unObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	//	if err != nil {
	//		return []string{}
	//	}
	//
	//	id, ok, err := unstructured.NestedString(unObj, "status", "id")
	//	if err != nil || !ok {
	//		return []string{}
	//	}
	//
	//	return []string{id}
	//}); err != nil {
	//	return fmt.Errorf("unable to setup field indexer for status.id of %v with: %w", gvk, err)
	//}

	kubeClient := kubeclient.NewClient(mgr.GetClient(), mgr.GetScheme())

	reconciler := &GenericReconciler{
		ARMClient:        applier,
		KubeClient:       kubeClient,
		ResourceResolver: armresourceresolver.NewResolver(kubeClient),
		Name:             t.Name(),
		Log:              log.WithName(controllerName),
		Recorder:         mgr.GetEventRecorderFor(controllerName),
		GVK:              gvk,
	}

	ctrlBuilder := ctrl.NewControllerManagedBy(mgr).
		For(obj).
		WithOptions(options)

	c, err := ctrlBuilder.Build(reconciler)
	if err != nil {
		return errors.Wrap(err, "unable to build controllers / reconciler")
	}

	reconciler.Controller = c

	return ctrl.NewWebhookManagedBy(mgr).
		For(obj).
		Complete()
}

// Reconcile will take state in K8s and apply it to Azure
func (gr *GenericReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := gr.Log.WithValues("name", req.Name, "namespace", req.Namespace)

	obj, err := gr.KubeClient.GetObject(ctx, req.NamespacedName, gr.GVK)
	if err != nil {
		return ctrl.Result{}, err
	}

	if obj == nil {
		// This means that the resource doesn't exist
		return ctrl.Result{}, nil
	}

	// The Go type for the Kubernetes object must understand how to
	// convert itself to/from the corresponding Azure types.
	metaObj, ok := obj.(genruntime.MetaObject)
	if !ok {
		return ctrl.Result{}, errors.Errorf("object is not a genruntime.MetaObject: %+v - type: %T", obj, obj)
	}

	reconcileData := NewReconcileMetadata(metaObj, log)
	action, err := reconcileData.DetermineReconcileAction()
	if err != nil {
		log.Error(err, "error determining reconcile action")
		gr.Recorder.Event(metaObj, v1.EventTypeWarning, "DetermineReconcileActionError", err.Error())
	}

	var result ctrl.Result
	var eventMsg string

	// TODO: Remaining work:
	// TODO:   1. Check that the resource owner is ready (see xform AreOwnersReady)
	// TODO:   2. Apply Kubernetes resource ownership (see xform ApplyOwnership).

	// TODO: Action could be a function here if we wanted?
	switch action {
	case ReconcileActionNoAction:
		result = ctrl.Result{} // No action
	case ReconcileActionBeginDeployment:
		result, err = gr.createDeployment(ctx, reconcileData)
		// TODO: Hmm this doesn't get the ID?
		eventMsg = fmt.Sprintf("Starting new deployment to Azure with ID: %q", getDeploymentId(metaObj))
	case ReconcileActionWatchDeployment:
		result, err = gr.watchDeployment(ctx, reconcileData)

		currentState := getResourceProvisioningState(metaObj)
		// TODO: Is there a better way to do this?
		var currentStateStr = "<nil>"
		if currentState != nil {
			currentStateStr = string(*currentState)
		}
		eventMsg = fmt.Sprintf(
			"Ongoing deployment with ID %q in state: %q",
			getDeploymentId(metaObj),
			currentStateStr)
	// For delete, there are 2 possible state transitions.
	// *  ReconcileActionBeginDelete --> Start deleting in Azure and mark state as "Deleting"
	// *  ReconcileActionWatchDelete --> http HEAD to see if resource still exists in Azure. If so, requeue, else, remove finalizer.
	case ReconcileActionBeginDelete:
		eventMsg = "start deleting resource"
		result, err = gr.reconcileDelete(ctx, log, metaObj, gr.startDeleteOfResource)
	case ReconcileActionWatchDelete:
		eventMsg = "deleting... checking for updated state"
		result, err = gr.reconcileDelete(ctx, log, metaObj, gr.updateFromNonTerminalDeleteState)
	}

	if eventMsg != "" {
		log.Info(eventMsg, "action", string(action))
		gr.Recorder.Event(metaObj, v1.EventTypeNormal, string(action), eventMsg)
	}

	if err != nil {
		log.Error(err, "Error during reconcile", "action", action)
		return result, err
	}

	return result, err
}

func (gr *GenericReconciler) reconcileDelete(
	ctx context.Context,
	log logr.Logger,
	metaObj genruntime.MetaObject,
	f func(context.Context, logr.Logger, genruntime.ArmResource, genruntime.MetaObject) (ctrl.Result, error)) (ctrl.Result, error) {

	log.Info("reconcile delete start")

	_, armResourceSpec, err := ResourceSpecToArmResourceSpec(ctx, gr, metaObj)
	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "couldn't convert to armResourceSpec")
	}
	// TODO: Not setting status here
	resource := genruntime.NewArmResource(armResourceSpec, nil, getResourceId(metaObj))

	return f(ctx, log, resource, metaObj)
}

// startDeleteOfResource will begin the delete of a resource by telling Azure to start deleting it. The resource will be
// marked with the provisioning state of "Deleting".
func (gr *GenericReconciler) startDeleteOfResource(
	ctx context.Context,
	log logr.Logger,
	resource genruntime.ArmResource,
	metaObj genruntime.MetaObject) (ctrl.Result, error) {

	if err := gr.KubeClient.PatchHelper(ctx, metaObj, func(ctx context.Context, mutMetaObject genruntime.MetaObject) error {
		if resource.GetId() != "" {
			emptyStatus, err := NewEmptyArmResourceStatus(metaObj)
			if err != nil {
				return errors.Wrapf(err, "failed trying to create empty status for %s", resource.GetId())
			}

			err = gr.ARMClient.BeginDeleteResource(ctx, resource.GetId(), resource.Spec().GetApiVersion(), emptyStatus)
			if err != nil {
				return errors.Wrapf(err, "failed trying to delete resource %s", resource.Spec().GetType())
			}

			annotations := make(map[string]string)
			annotations[ResourceStateAnnotation] = string(armclient.DeletingProvisioningState)
			addAnnotations(metaObj, annotations)
		} else {
			controllerutil.RemoveFinalizer(mutMetaObject, GenericControllerFinalizer)
		}

		// TODO: Not exactly sure what this was doing
		//if err := gr.Converter.FromResource(resource, mutMetaObject); err != nil {
		//	return errors.Wrapf(err, "error gr.Converter.FromResource")
		//}

		return nil
	}); err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, errors.Wrap(err, "failed to patch after starting delete")
	}

	// delete has started, check back to seen when the finalizer can be removed
	return ctrl.Result{
		RequeueAfter: 5 * time.Second,
	}, nil
}

// updateFromNonTerminalDeleteState will call Azure to check if the resource still exists. If so, it will requeue, else,
// the finalizer will be removed.
func (gr *GenericReconciler) updateFromNonTerminalDeleteState(
	ctx context.Context,
	log logr.Logger,
	resource genruntime.ArmResource,
	metaObj genruntime.MetaObject) (ctrl.Result, error) {

	// already deleting, just check to see if it still exists and if it's gone, remove finalizer
	found, err := gr.ARMClient.HeadResource(ctx, resource.GetId(), resource.Spec().GetApiVersion())
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "failed to head resource")
	}

	if found {
		log.V(0).Info("Found resource: continuing to wait for deletion...")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	err = gr.KubeClient.PatchHelper(ctx, metaObj, func(ctx context.Context, mutMetaObject genruntime.MetaObject) error {
		controllerutil.RemoveFinalizer(mutMetaObject, GenericControllerFinalizer)
		return nil
	})

	// patcher will try to fetch the object after patching, so ignore not found errors
	if apierrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, err
}

func (gr *GenericReconciler) getStatus(ctx context.Context, id string, data *ReconcileMetadata) (genruntime.ArmTransformer, error) {
	_, typedArmSpec, err := ResourceSpecToArmResourceSpec(ctx, gr, data.metaObj)
	if err != nil {
		return nil, err
	}

	// TODO: do we tolerate not exists here?
	armStatus, err := NewEmptyArmResourceStatus(data.metaObj)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to construct empty ARM status object for resource with ID: %q", id)
	}

	// Get the resource
	err = gr.ARMClient.GetResource(ctx, id, typedArmSpec.GetApiVersion(), armStatus)
	if data.log.V(0).Enabled() { // TODO: Change this level
		statusBytes, err := json.Marshal(armStatus)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to serialize ARM status to JSON for debugging")
		}
		data.log.V(0).Info("Got ARM status", "status", string(statusBytes))
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to get resource with ID: %q", id)
	}

	// Convert the ARM shape to the Kube shape
	status, err := NewEmptyStatus(data.metaObj)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to construct empty status object for resource with ID: %q", id)
	}

	// TODO: need to use KnownOwner?
	owner := data.metaObj.Owner()
	correctOwner := genruntime.KnownResourceReference{
		Name: owner.Name,
	}

	// Fill the kube status with the results from the arm status
	err = status.PopulateFromArm(correctOwner, ValueOfPtr(armStatus)) // TODO: PopulateFromArm expects a value... ick
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't convert ARM status to Kubernetes status")
	}

	return status, nil
}

func (gr *GenericReconciler) updateMetaObject(
	metaObj genruntime.MetaObject,
	deployment *armclient.Deployment,
	status genruntime.ArmTransformer) error {

	controllerutil.AddFinalizer(metaObj, GenericControllerFinalizer)

	sig, err := SpecSignature(metaObj)
	if err != nil {
		return errors.Wrap(err, "failed to compute resource spec hash")
	}

	annotations := map[string]string{
		ResourceSigAnnotationKey: sig,
		DeploymentIdAnnotation:   deployment.Id,
		DeploymentNameAnnotation: deployment.Name,
		// TODO: Do we want to just use Azure's annotations here? I bet we don't? We probably want to map
		// TODO: them onto something more robust? For now just use Azure's though.
		ResourceStateAnnotation: string(deployment.Properties.ProvisioningState),
	}

	if deployment.IsTerminalProvisioningState() {
		if deployment.Properties.ProvisioningState == armclient.FailedProvisioningState {
			annotations[ResourceErrorAnnotation] = deployment.Properties.Error.String()
		} else if len(deployment.Properties.OutputResources) > 0 {
			resourceId := deployment.Properties.OutputResources[0].ID
			annotations[ResourceIdAnnotation] = resourceId

			if status != nil {
				err = SetStatus(metaObj, status)
				if err != nil {
					return err
				}
			}
		} else {
			return errors.Errorf("template deployment didn't have any output resources")
		}
	}

	addAnnotations(metaObj, annotations)
	return nil
}

func (gr *GenericReconciler) createDeployment(ctx context.Context, data *ReconcileMetadata) (ctrl.Result, error) {
	deployment, err := gr.resourceSpecToDeployment(ctx, data)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := gr.KubeClient.PatchHelper(ctx, data.metaObj, func(ctx context.Context, mutObj genruntime.MetaObject) error {
		deployment, err = gr.ARMClient.CreateDeployment(ctx, deployment)
		if err != nil {
			return err
		}

		err = gr.updateMetaObject(mutObj, deployment, nil) // Status is always nil here
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		// This is a superfluous error as per https://github.com/kubernetes-sigs/controller-runtime/issues/377
		// The correct handling is just to ignore it and we will get an event shortly with the updated version to patch
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, errors.Wrap(err, "failed to patch")
	}

	result := ctrl.Result{}
	// TODO: This is going to be common... need a wrapper/helper somehow?
	if !deployment.IsTerminalProvisioningState() {
		result = ctrl.Result{
			RequeueAfter: 5 * time.Second,
		}
	}
	return result, err
}

// TODO: There's a bit too much duplicated code between this and create deployment -- should be a good way to combine them?
func (gr *GenericReconciler) watchDeployment(ctx context.Context, data *ReconcileMetadata) (ctrl.Result, error) {
	deployment, err := gr.resourceSpecToDeployment(ctx, data)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := gr.KubeClient.PatchHelper(ctx, data.metaObj, func(ctx context.Context, mutObj genruntime.MetaObject) error {
		deployment, err = gr.ARMClient.GetDeployment(ctx, deployment.Id)
		if err != nil {
			return errors.Wrapf(err, "failed getting deployment %q from ARM", deployment.Id)
		}

		var status genruntime.ArmTransformer
		if deployment.Properties != nil && deployment.Properties.ProvisioningState == armclient.SucceededProvisioningState {
			// TODO: There's some overlap here with what updateMetaObject does
			if len(deployment.Properties.OutputResources) == 0 {
				return errors.Errorf("Template deployment didn't have any output resources")
			}

			status, err = gr.getStatus(ctx, deployment.Properties.OutputResources[0].ID, data)
			if err != nil {
				return errors.Wrap(err, "Failed getting status from ARM")
			}
		}

		if deployment.IsTerminalProvisioningState() && !getShouldPreserveDeployment(data.metaObj) {
			err := gr.ARMClient.DeleteDeployment(ctx, deployment.Id)
			if err != nil {
				return errors.Wrapf(err, "failed to delete deployment %q", deployment.Id)
			}
			deployment.Id = ""
			deployment.Name = ""
		}

		err = gr.updateMetaObject(data.metaObj, deployment, status)
		if err != nil {
			return errors.Wrap(err, "failed updating metaObj")
		}

		return nil
	}); err != nil {
		// This is a superfluous error as per https://github.com/kubernetes-sigs/controller-runtime/issues/377
		// The correct handling is just to ignore it and we will get an event shortly with the updated version to patch
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, errors.Wrap(err, "failed to patch")
	}

	result := ctrl.Result{}
	// TODO: This is going to be common... need a wrapper/helper somehow?
	if !deployment.IsTerminalProvisioningState() {
		result = ctrl.Result{
			RequeueAfter: 5 * time.Second,
		}
	}
	return result, err
}

func (gr *GenericReconciler) resourceSpecToDeployment(ctx context.Context, data *ReconcileMetadata) (*armclient.Deployment, error) {
	resourceGroupName, typedArmSpec, err := ResourceSpecToArmResourceSpec(ctx, gr, data.metaObj)
	if err != nil {
		return nil, err
	}

	// TODO: get other deployment details from status and avoid creating a new deployment
	deploymentId, deploymentIdOk := data.metaObj.GetAnnotations()[DeploymentIdAnnotation]
	deploymentName, deploymentNameOk := data.metaObj.GetAnnotations()[DeploymentNameAnnotation]
	if deploymentIdOk != deploymentNameOk {
		return nil, errors.Errorf(
			"deploymentIdOk: %t, deploymentNameOk: %t expected to match, but didn't",
			deploymentIdOk,
			deploymentNameOk)
	}

	var deployment *armclient.Deployment
	if deploymentIdOk && deploymentNameOk {
		deployment = gr.ARMClient.NewDeployment(resourceGroupName, deploymentName, typedArmSpec)
		deployment.Id = deploymentId
	} else {
		deploymentName, err := CreateDeploymentName()
		if err != nil {
			return nil, err
		}
		deployment = gr.ARMClient.NewDeployment(resourceGroupName, deploymentName, typedArmSpec)
	}
	return deployment, nil
}

func hasResourceHashAnnotationChanged(metaObj genruntime.MetaObject) (bool, error) {
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

func getResourceProvisioningState(metaObj genruntime.MetaObject) *armclient.ProvisioningState {
	state, ok := metaObj.GetAnnotations()[ResourceStateAnnotation]
	if !ok {
		return nil
	}

	result := armclient.ProvisioningState(state)
	return &result
}

func getDeploymentId(metaObj genruntime.MetaObject) string {
	id, ok := metaObj.GetAnnotations()[DeploymentIdAnnotation]
	if !ok {
		return ""
	}

	return id
}

func getShouldPreserveDeployment(metaObj genruntime.MetaObject) bool {
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
func getResourceId(metaObj genruntime.MetaObject) string {
	id, ok := metaObj.GetAnnotations()[ResourceIdAnnotation]
	if !ok {
		return ""
	}

	return id
}

func addAnnotations(metaObj genruntime.MetaObject, newAnnotations map[string]string) {
	annotations := metaObj.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	for k, v := range newAnnotations {
		annotations[k] = v
	}

	metaObj.SetAnnotations(annotations)
}
