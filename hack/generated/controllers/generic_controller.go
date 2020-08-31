/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/Azure/k8s-infra/apis"
	"github.com/Azure/k8s-infra/hack/generated/pkg/armclient"
	"github.com/Azure/k8s-infra/pkg/util/patch"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
)

const (
	// ResourceSigAnnotationKey is an annotation key which holds the value of the hash of the spec
	ResourceSigAnnotationKey = "resource-sig.infra.azure.com"

	// TODO: Delete these later in favor of something in status?
	DeploymentIdAnnotation = "deployment-id.infra.azure.com"
	DeploymentNameAnnotation = "deployment-name.infra.azure.com"
	ResourceStateAnnotation = "resource-state.infra.azure.com"
	ResourceIdAnnotation = "resource-id.infra.azure.com"
	ResourceErrorAnnotation = "resource-error.infra.azure.com"
	// PreserveDeploymentAnnotation is the key which tells the applier to keep or delete the deployment
	PreserveDeploymentAnnotation = "x-preserve-deployment"
)

// GenericReconciler reconciles STUFF
type GenericReconciler struct {
	Client     client.Client
	Log        logr.Logger
	ARMClient  armclient.Applier
	Scheme     *runtime.Scheme
	Recorder   record.EventRecorder
	Name       string
	GVK        schema.GroupVersionKind
	Controller controller.Controller
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

// TODO: is this comment still correct...?
// register takes a manager and a struct describing how to instantiate
// controllers for various types using a generic reconciler function. Only one
// type (using For) may be directly watched by each controller, but zero, one or
// many Owned types are acceptable. This setup allows reconcileFn to have access
// to the concrete type defined as part of a closure, while allowing for
// independent controllers per GVK (== better parallelism, vs 1 controllers
// managing many, many List/Watches)
func register(mgr ctrl.Manager, applier armclient.Applier, obj runtime.Object, log logr.Logger, options controller.Options) error {
	v, err := conversion.EnforcePtr(obj)
	if err != nil {
		return err
	}

	t := v.Type()
	controllerName := fmt.Sprintf("%sController", t.Name())

	// Use the provided GVK to construct a new runtime object of the desired concrete type.
	gvk, err := apiutil.GVKForObject(obj, mgr.GetScheme())
	if err != nil {
		return err
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

	reconciler := &GenericReconciler{
		Client:    mgr.GetClient(),
		ARMClient: applier,
		Scheme:    mgr.GetScheme(),
		Name:      t.Name(),
		Log:       log.WithName(controllerName),
		Recorder:  mgr.GetEventRecorderFor(controllerName),
		GVK:       gvk,
	}

	ctrlBuilder := ctrl.NewControllerManagedBy(mgr).
		For(obj).
		WithOptions(options)

	// TODO: I think we need to do the below, but skipping it for now...
	//if metaObj, ok := obj.(azcorev1.MetaObject); ok {
	//	trls, err := xform.GetTypeReferenceData(metaObj)
	//	if err != nil {
	//		return fmt.Errorf("unable get type reference data for obj %v with: %w", metaObj, err)
	//	}
	//
	//	for _, trl := range trls {
	//		if trl.IsOwned {
	//			gvk := schema.GroupVersionKind{
	//				Group:   trl.Group,
	//				Version: "v1",
	//				Kind:    trl.Kind,
	//			}
	//			ownedObj, err := mgr.GetScheme().New(gvk)
	//			if err != nil {
	//				return fmt.Errorf("unable to create GVK %v with: %w", gvk, err)
	//			}
	//
	//			reconciler.Owns = append(reconciler.Owns, ownedObj)
	//			ctrlBuilder.Owns(ownedObj)
	//		}
	//	}
	//}

	c, err := ctrlBuilder.Build(reconciler)
	if err != nil {
		return fmt.Errorf("unable to build controllers / reconciler with: %w", err)
	}

	reconciler.Controller = c

	return ctrl.NewWebhookManagedBy(mgr).
		For(obj).
		Complete()
}

func (gr *GenericReconciler) GetObject(namespacedName types.NamespacedName, gvk schema.GroupVersionKind) (runtime.Object, error) {
	ctx := context.Background()

	obj, err := gr.Scheme.New(gvk)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create object from gvk %+v with", gr.GVK)
	}

	if err := gr.Client.Get(ctx, namespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return obj, nil
}

// Reconcile will take state in K8s and apply it to Azure
func (gr *GenericReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := gr.Log.WithValues("name", req.Name, "namespace", req.Namespace)

	obj, err := gr.GetObject(req.NamespacedName, gr.GVK)
	if err != nil {
		return ctrl.Result{}, err
	}

	if obj == nil {
		// This means that the resource doesn't exist
		return ctrl.Result{}, nil
	}

	// ToUnstructured and get spec? or else interface that returns spec as interface{}

	// The Go type for the Kubernetes object must understand how to
	// convert itself to/from the corresponding Azure types.
	metaObj, ok := obj.(genruntime.MetaObject)
	if !ok {
		return ctrl.Result{}, fmt.Errorf("object is not a genruntime.MetaObject: %+v - type: %T", obj, obj)
	}

	// TODO: Consider making this a big switch (with clean cases)

	// Like this!:
	// 1. Extract useful fields... put into struct?
	// 2. Call "DetermineObjectAction" on struct
	// 3. Switch on object action

	// reconcile delete
	if !metaObj.GetDeletionTimestamp().IsZero() {
		log.Info("reconcile delete start")
		result, err := gr.reconcileDelete(ctx, log, metaObj)
		if err != nil {
			gr.Recorder.Event(metaObj, v1.EventTypeWarning, "ReconcileDeleteError", err.Error())
			log.Error(err, "reconcile delete error")
			return result, err
		}

		log.Info("reconcile delete complete")
		return result, err
	}

	//if grouped, ok := obj.(azcorev1.Grouped); ok {
	//	ready, err := gr.isResourceGroupReady(ctx, grouped)
	//	if err != nil {
	//		log.Error(err, "failed checking if resource group was ready")
	//		gr.Recorder.Event(metaObj, v1.EventTypeWarning, "GroupReadyError", fmt.Sprintf("isResourceGroupReady failed with: %s", err))
	//		return ctrl.Result{}, err
	//	}
	//
	//	if !ready {
	//		requeueTime := 30 * time.Second
	//		msg := fmt.Sprintf("resource group %q is not ready or not created yet; will try again in about %s", grouped.GetResourceGroupObjectRef().Name, requeueTime)
	//		gr.Recorder.Event(metaObj, v1.EventTypeNormal, "ResourceGroupNotReady", msg)
	//		return ctrl.Result{
	//			RequeueAfter: requeueTime,
	//		}, nil
	//	}
	//}

	// TODO: Check if status of owners is good

	//ownersReady, err := gr.Converter.AreOwnersReady(ctx, metaObj)
	//if err != nil {
	//	log.Error(err, "failed checking owner references are ready")
	//	gr.Recorder.Event(metaObj, v1.EventTypeWarning, "OwnerReferenceCheckError", err.Error())
	//	return ctrl.Result{}, err
	//}
	//
	//if !ownersReady {
	//	gr.Recorder.Event(metaObj, v1.EventTypeNormal, "OwnerReferencesNotReady", "owner reference are not ready; retrying in 30s")
	//	return ctrl.Result{
	//		RequeueAfter: 30 * time.Second,
	//	}, nil
	//}

	log.Info("reconcile apply start")
	result, err := gr.reconcileApply(ctx, metaObj, log)
	if err != nil {
		log.Error(err, "reconcile apply error")
		gr.Recorder.Event(metaObj, v1.EventTypeWarning, "ReconcileError", err.Error())
		return result, err
	}

	// TODO: ApplyDeployment ownership
	//allApplied, err := gr.Converter.ApplyOwnership(ctx, metaObj)
	//if err != nil {
	//	log.Error(err, "failed applying ownership to owned references")
	//	gr.Recorder.Event(metaObj, v1.EventTypeWarning, "OwnerReferencesFailedApply", "owner reference are not ready; retrying in 30s")
	//	return ctrl.Result{}, fmt.Errorf("failed applying ownership to owned references with: %w", err)
	//}
	//
	//if !allApplied {
	//	log.Info("not all owned objects were applied; will requeue for 30 seconds")
	//	result.RequeueAfter = 30 * time.Second
	//}

	log.Info("reconcile apply complete")
	return result, err
}

// reconcileApply will determine what, if anything, has changed on the resource, and apply that state to Azure.
// The Az infra finalizer will be applied.
//
// There are 3 possible state transitions.
// *  New resource (hasChanged = true) --> Start deploying the resource and move provisioning state from "" to "{Accepted || Succeeded || Failed}"
// *  Existing Resource (hasChanged = true) --> Start deploying the resource and move state from "{terminal state} to "{Accepted || Succeeded || Failed}"
// *  Existing Resource (hasChanged = false) --> Probably a status change. Don't do anything as of now.
func (gr *GenericReconciler) reconcileApply(ctx context.Context, metaObj genruntime.MetaObject, log logr.Logger) (ctrl.Result, error) {
	// check if the hash on the resource has changed
	// TODO: Shouldn't this be taking into account the shape of the object in Azure as well?
	hasChanged, err := hasResourceHashAnnotationChanged(metaObj)
	if err != nil {
		err = fmt.Errorf("failed comparing resource hash with: %w", err)
		gr.Recorder.Event(metaObj, v1.EventTypeWarning, "AnnotationError", err.Error())
		return ctrl.Result{}, err
	}

	provisioningState := getResourceProvisioningState(metaObj)

	// if the resource hash (spec) has not changed, don't apply again
	if !hasChanged && provisioningState != nil && armclient.IsTerminalProvisioningState(*provisioningState) {
		msg := fmt.Sprintf("resource spec has not changed and resource is in terminal state: %q", *provisioningState)
		log.V(0).Info(msg)
		gr.Recorder.Event(metaObj, v1.EventTypeNormal, "ResourceHasNotChanged", msg)
		return ctrl.Result{}, nil
	}

	// TODO: is there a better way to do this?
	provisioningStateStr := "<nil>"
	if provisioningState != nil {
		provisioningStateStr = string(*provisioningState)
	}

	deploymentId := getDeploymentId(metaObj)
	msg := fmt.Sprintf(
		"resource spec for %q will be applied to Azure. DeploymentId: %q, Current provisioning state: %q",
		metaObj.GetName(),
		deploymentId,
		provisioningStateStr)
	gr.Recorder.Event(metaObj, v1.EventTypeNormal, "ResourceHasChanged", msg)
	return gr.applySpecChange(ctx, metaObj) // TODO: pass log?
}

// TODO: ??
//func (gr *GenericReconciler) isResourceGroupReady(ctx context.Context, grouped azcorev1.Grouped) (bool, error) {
//	// has a resource group, so check if the resource group is already provisioned
//	groupRef := grouped.GetResourceGroupObjectRef()
//	if groupRef == nil {
//		return false, fmt.Errorf("grouped resources must have a resource group")
//	}
//
//	key := client.ObjectKey{
//		Name:      groupRef.Name,
//		Namespace: groupRef.Namespace,
//	}
//
//	// get the storage version of the resource group regardless of the referenced version
//	var rg microsoftresourcesv1.ResourceGroup
//	if err := gr.Client.Get(ctx, key, &rg); err != nil {
//		if apierrors.IsNotFound(err) {
//			// not able to find the resource, but that's ok. It might not exist yet
//			return false, nil
//		}
//		return false, fmt.Errorf("error GETing rg resource with: %w", err)
//	}
//
//	status, err := statusutil.GetProvisioningState(&rg)
//	if err != nil {
//		return false, fmt.Errorf("unable to get provisioning state for resource group with: %w", err)
//	}
//
//	return zips.ProvisioningState(status) == zips.SucceededProvisioningState, nil
//}

// reconcileDelete will begin and follow a delete operation of a resource. The finalizer will only be removed upon the
// resource actually being deleted in Azure.
//
// There are 2 possible state transitions.
// *  obj.ProvisioningState == \*\ --> Start deleting in Azure and mark state as "Deleting"
// *  obj.ProvisioningState == "Deleting" --> http HEAD to see if resource still exists in Azure. If so, requeue, else, remove finalizer.
func (gr *GenericReconciler) reconcileDelete(ctx context.Context, log logr.Logger, metaObj genruntime.MetaObject) (ctrl.Result, error) {
	provisioningState := getResourceProvisioningState(metaObj)

	_, armResourceSpec, err := ResourceSpecToArmResourceSpec(gr, metaObj)
	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "couldn't convert to armResourceSpec")
	}
	// TODO: Not setting status here
	resource := genruntime.NewArmResource(armResourceSpec, nil, getResourceId(metaObj))

	if provisioningState != nil && *provisioningState == armclient.DeletingProvisioningState {
		msg := fmt.Sprintf("deleting... checking for updated state")
		log.V(0).Info(msg)
		gr.Recorder.Event(metaObj, v1.EventTypeNormal, "ResourceDeleteInProgress", msg)
		return gr.updateFromNonTerminalDeleteState(ctx, log, resource, metaObj)
	} else {
		msg := fmt.Sprintf("start deleting resource in state %v", provisioningState)
		gr.Recorder.Event(metaObj, v1.EventTypeNormal, "ResourceDeleteStart", msg)
		return gr.startDeleteOfResource(ctx, log, resource, metaObj)
	}
}

// startDeleteOfResource will begin the delete of a resource by telling Azure to start deleting it. The resource will be
// marked with the provisioning state of "Deleting".
func (gr *GenericReconciler) startDeleteOfResource(
	ctx context.Context,
	log logr.Logger,
	resource genruntime.ArmResource,
	metaObj genruntime.MetaObject) (ctrl.Result, error) {

	if err := patcher(ctx, gr.Client, metaObj, func(ctx context.Context, mutMetaObject genruntime.MetaObject) error {
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
			controllerutil.RemoveFinalizer(mutMetaObject, apis.AzureInfraFinalizer)
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

	err = patcher(ctx, gr.Client, metaObj, func(ctx context.Context, mutMetaObject genruntime.MetaObject) error {
		controllerutil.RemoveFinalizer(mutMetaObject, apis.AzureInfraFinalizer)
		return nil
	})

	// patcher will try to fetch the object after patching, so ignore not found errors
	if apierrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, err
}

// TODO: it feels like theres a state machine and some formalism hiding here

// TODO: Refactor this function -- it's ugly
// applySpecChange will apply the new spec state to an Azure resource. The resource should then enter
// into a non terminal state and will then be requeued for polling.
func (gr *GenericReconciler) applySpecChange(ctx context.Context, metaObj genruntime.MetaObject) (ctrl.Result, error) {

	deployment, err := gr.resourceSpecToDeployment(metaObj)
	if err != nil {
		err := errors.Wrapf(err, "unable to resource spec to deployment")
		gr.Recorder.Event(metaObj, v1.EventTypeWarning, "ResourceSpecToDeployemntError", err.Error())
		return ctrl.Result{}, err
	}

	if err := patcher(ctx, gr.Client, metaObj, gr.applyDeploymentAndPatchMetaObj); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "failed to patch")
	}

	result := ctrl.Result{}
	if !deployment.IsTerminalProvisioningState() {
		result = ctrl.Result{
			RequeueAfter: 5 * time.Second,
		}
	}
	return result, err
}

func (gr *GenericReconciler) applyDeploymentAndPatchMetaObj(ctx context.Context, metaObj genruntime.MetaObject) error {
	deployment, err := gr.resourceSpecToDeployment(metaObj)
	if err != nil {
		err := errors.Wrapf(err, "unable to resource spec to deployment")
		gr.Recorder.Event(metaObj, v1.EventTypeWarning, "ResourceSpecToDeployemntError", err.Error())
		return err
	}

	controllerutil.AddFinalizer(metaObj, apis.AzureInfraFinalizer)
	deployment, err = gr.applyDeployment(ctx, metaObj, deployment)
	if err != nil {
		return errors.Wrap(err, "failed to apply state to Azure")
	}

	// TODO: we need to update status with provisioning state and whatnot
	// TODO: For now, sticking a few things into annotations
	//if err := gr.Converter.FromResource(resource, mutObj); err != nil {
	//	return err
	//}

	sig, err := SpecSignature(metaObj)
	if err != nil {
		return errors.Wrap(err, "failed to compute resource spec hash")
	}

	annotations := make(map[string]string)
	annotations[ResourceSigAnnotationKey] = sig
	annotations[DeploymentIdAnnotation] = deployment.Id
	annotations[DeploymentNameAnnotation] = deployment.Name

	// TODO: Do we want to just use Azure's annotations here? I bet we don't? We probably want to map
	// TODO: them onto something more robust? For now just use Azure's though.
	annotations[ResourceStateAnnotation] = string(deployment.Properties.ProvisioningState)

	if deployment.IsTerminalProvisioningState() {
		if deployment.Properties.ProvisioningState == armclient.FailedProvisioningState {
			annotations[ResourceErrorAnnotation] = deployment.Properties.Error.String()
		} else if len(deployment.Properties.OutputResources) == 0 {
			return errors.Errorf("Template deployment didn't have any output resources")
		} else {
			resourceId := deployment.Properties.OutputResources[0].ID
			annotations[ResourceIdAnnotation] = resourceId

			// TODO: We technically called this function above already (it's called by resourceSpecToDeployment).
			// TODO: Feels like there should be a cleaner way to get this and avoid the 2x call.
			_, typedArmSpec, err := ResourceSpecToArmResourceSpec(gr, metaObj)
			if err != nil {
				return err
			}

			// TODO: do we tolerate not exists here?
			armStatus, err := NewEmptyArmResourceStatus(metaObj)
			if err != nil {
				return errors.Wrapf(err, "failed to construct empty ARM status object for resource with ID: %q", resourceId)
			}

			// Get the resource
			err = gr.ARMClient.GetResource(ctx, resourceId, typedArmSpec.GetApiVersion(), armStatus)
			if err != nil {
				return errors.Wrapf(err, "failed to get resource with ID: %q", resourceId)
			}

			// Convert the ARM shape to the Kube shape
			status, err := NewEmptyStatus(metaObj)
			if err != nil {
				return errors.Wrapf(err, "failed to construct empty status object for resource with ID: %q", resourceId)
			}

			// TODO: need to use KnownOwner?
			owner := metaObj.Owner()
			correctOwner := genruntime.KnownResourceReference{
				Name: owner.Name,
			}
			// Fill the kube status with the results from the arm status
			err = status.PopulateFromArm(correctOwner, armStatus)
			if err != nil {
				return errors.Wrapf(err, "couldn't convert ARM status to Kubernetes status")
			}

			err = SetStatus(metaObj, status)
			if err != nil {
				return err
			}
		}
	}

	addAnnotations(metaObj, annotations)
	return nil
}

func (gr *GenericReconciler) resourceSpecToDeployment(metaObject genruntime.MetaObject) (*armclient.Deployment, error) {
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

// ApplyDeployment deploys a resource to Azure via a deployment template
func (gr *GenericReconciler) applyDeployment(
	ctx context.Context,
	metaObject genruntime.MetaObject,
	deployment *armclient.Deployment) (*armclient.Deployment, error) {

	provisioningState := getResourceProvisioningState(metaObject)
	shouldPreserveDeployment := getShouldPreserveDeployment(metaObject)

	switch {
	case provisioningState != nil && *provisioningState == armclient.DeletingProvisioningState:
		return deployment, errors.Errorf("resource is currently deleting; it can not be applied")
	case deployment.IsTerminalProvisioningState() && deployment.Id != "" && !shouldPreserveDeployment:
		err := gr.ARMClient.DeleteDeployment(ctx, deployment.Id)
		if err != nil {
			return deployment, err
		}
		deployment.Id = "" // We've deleted the deployment successfully so clear the ID
		return deployment, nil
	case deployment.Id != "":
		// existing deployment is already going, so let's get an updated status
		return gr.ARMClient.GetDeployment(ctx, deployment.Id)
	default:
		// no provisioning state and no deployment ID, so we need to start a new deployment
		return gr.ARMClient.CreateDeployment(ctx, deployment)
	}
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

func patcher(ctx context.Context, c client.Client, metaObj genruntime.MetaObject, mutator func(context.Context, genruntime.MetaObject) error) error {
	patchHelper, err := patch.NewHelper(metaObj, c)
	if err != nil {
		return err
	}

	if err := mutator(ctx, metaObj); err != nil {
		return err
	}

	if err := patchHelper.Patch(ctx, metaObj); err != nil {
		return err
	}

	// fill resourcer with patched updates since patch will copy resourcer
	return c.Get(ctx, client.ObjectKey{
		Namespace: metaObj.GetNamespace(),
		Name:      metaObj.GetName(),
	}, metaObj)
}
