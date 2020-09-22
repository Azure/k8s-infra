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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
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

type ReconcileAction string

const (
	ReconcileActionNoAction        = ReconcileAction("NoAction")
	ReconcileActionBeginDeployment = ReconcileAction("BeginDeployment")
	ReconcileActionWatchDeployment = ReconcileAction("WatchDeployment")
	ReconcileActionBeginDelete     = ReconcileAction("BeginDelete")
	ReconcileActionWatchDelete     = ReconcileAction("WatchDelete")
)

type ReconcileActionFunc = func(ctx context.Context, action ReconcileAction, obj *ReconcileMetadata) (ctrl.Result, error)

func RegisterAll(mgr ctrl.Manager, applier armclient.Applier, objs []runtime.Object, log logr.Logger, options controller.Options) []error {
	var errs []error
	for _, obj := range objs {
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

	// TODO: Do we need to add any index fields here? DavidJ's controller index's status.id - see its usage
	// TODO: of IndexField

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

	obj, err := gr.KubeClient.GetObjectOrDefault(ctx, req.NamespacedName, gr.GVK)
	if err != nil {
		return ctrl.Result{}, err
	}

	if obj == nil {
		// This means that the resource doesn't exist
		return ctrl.Result{}, nil
	}

	// Always operate on a copy rather than the object from the client, as per
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/controllers.md, which says:
	// Never mutate original objects! Caches are shared across controllers, this means that if you mutate your "copy"
	// (actually a reference or shallow copy) of an object, you'll mess up other controllers (not just your own).
	obj = obj.DeepCopyObject()

	// The Go type for the Kubernetes object must understand how to
	// convert itself to/from the corresponding Azure types.
	metaObj, ok := obj.(genruntime.MetaObject)
	if !ok {
		return ctrl.Result{}, errors.Errorf("object is not a genruntime.MetaObject: %+v - type: %T", obj, obj)
	}

	objWrapper := NewReconcileMetadata(metaObj, log)
	action, actionFunc, err := gr.DetermineReconcileAction(objWrapper)

	if err != nil {
		log.Error(err, "error determining reconcile action")
		gr.Recorder.Event(metaObj, v1.EventTypeWarning, "DetermineReconcileActionError", err.Error())
	}

	result, err := actionFunc(ctx, action, objWrapper)
	if err != nil {
		log.Error(err, "Error during reconcile", "action", action)
		gr.Recorder.Event(metaObj, v1.EventTypeWarning, "ReconcileActionError", err.Error())
	}

	return result, err
}

func (gr *GenericReconciler) DetermineReconcileAction(obj *ReconcileMetadata) (ReconcileAction, ReconcileActionFunc, error) {
	if !obj.metaObj.GetDeletionTimestamp().IsZero() {
		if obj.resourceProvisioningState != nil && *obj.resourceProvisioningState == armclient.DeletingProvisioningState {
			return ReconcileActionWatchDelete, gr.UpdateFromNonTerminalDeleteState, nil
		}
		return ReconcileActionBeginDelete, gr.StartDeleteOfResource, nil
	} else {
		hasChanged, err := genruntime.HasResourceSpecHashChanged(obj.metaObj)
		if err != nil {
			return ReconcileActionNoAction, NoAction, errors.Wrap(err, "failed comparing resource hash")
		}

		//if !hasChanged && r.IsTerminalProvisioningState() && hasStatus {
		if !hasChanged && obj.IsTerminalProvisioningState() {
			// TODO: Do we want to log here?
			msg := fmt.Sprintf("resource spec has not changed and resource is in terminal state: %q", *obj.resourceProvisioningState)
			obj.log.V(0).Info(msg)
			return ReconcileActionNoAction, NoAction, nil
		}

		if obj.resourceProvisioningState != nil && *obj.resourceProvisioningState == armclient.DeletingProvisioningState {
			return ReconcileActionNoAction, NoAction, errors.Errorf("resource is currently deleting; it can not be applied")
		}

		if obj.deploymentId != "" {
			// There is an ongoing deployment we need to monitor
			return ReconcileActionWatchDeployment, gr.WatchDeployment, nil
		}

		return ReconcileActionBeginDeployment, gr.CreateDeployment, nil
	}
}

//////////////////////////////////////////
// Actions
//////////////////////////////////////////

func NoAction(ctx context.Context, action ReconcileAction, obj *ReconcileMetadata) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// StartDeleteOfResource will begin the delete of a resource by telling Azure to start deleting it. The resource will be
// marked with the provisioning state of "Deleting".
func (gr *GenericReconciler) StartDeleteOfResource(
	ctx context.Context,
	action ReconcileAction,
	obj *ReconcileMetadata) (ctrl.Result, error) {

	msg := "Starting delete of resource"
	obj.log.Info(msg)
	gr.Recorder.Event(obj.metaObj, v1.EventTypeNormal, string(action), msg)

	resource, err := gr.constructArmResource(ctx, obj)
	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "couldn't convert to armResourceSpec")
	}

	err = gr.KubeClient.PatchHelper(ctx, obj.metaObj, func(ctx context.Context, mutObj genruntime.MetaObject) error {
		if resource.GetId() != "" {
			emptyStatus, err := NewEmptyArmResourceStatus(obj.metaObj)
			if err != nil {
				return errors.Wrapf(err, "failed trying to create empty status for %s", resource.GetId())
			}

			err = gr.ARMClient.BeginDeleteResource(ctx, resource.GetId(), resource.Spec().GetApiVersion(), emptyStatus)
			if err != nil {
				return errors.Wrapf(err, "failed trying to delete resource %s", resource.Spec().GetType())
			}

			genruntime.SetResourceState(obj.metaObj, string(armclient.DeletingProvisioningState))
		} else {
			controllerutil.RemoveFinalizer(mutObj, GenericControllerFinalizer)
		}

		return nil
	})

	if err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, errors.Wrap(err, "failed to patch after starting delete")
	}

	// delete has started, check back to seen when the finalizer can be removed
	return ctrl.Result{
		RequeueAfter: 5 * time.Second,
	}, nil
}

// UpdateFromNonTerminalDeleteState will call Azure to check if the resource still exists. If so, it will requeue, else,
// the finalizer will be removed.
func (gr *GenericReconciler) UpdateFromNonTerminalDeleteState(
	ctx context.Context,
	action ReconcileAction,
	obj *ReconcileMetadata) (ctrl.Result, error) {

	msg := "Continue monitoring ongoing delete"
	obj.log.Info(msg)
	gr.Recorder.Event(obj.metaObj, v1.EventTypeNormal, string(action), msg)

	resource, err := gr.constructArmResource(ctx, obj)
	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "couldn't convert to armResourceSpec")
	}

	// already deleting, just check to see if it still exists and if it's gone, remove finalizer
	found, err := gr.ARMClient.HeadResource(ctx, resource.GetId(), resource.Spec().GetApiVersion())
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "failed to head resource")
	}

	if found {
		obj.log.V(0).Info("Found resource: continuing to wait for deletion...")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	err = gr.KubeClient.PatchHelper(ctx, obj.metaObj, func(ctx context.Context, mutObj genruntime.MetaObject) error {
		controllerutil.RemoveFinalizer(mutObj, GenericControllerFinalizer)
		return nil
	})

	// patcher will try to fetch the object after patching, so ignore not found errors
	if apierrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, err
}

func (gr *GenericReconciler) CreateDeployment(ctx context.Context, action ReconcileAction, data *ReconcileMetadata) (ctrl.Result, error) {
	deployment, err := gr.resourceSpecToDeployment(ctx, data)
	if err != nil {
		return ctrl.Result{}, err
	}

	// TODO: Could somehow have a method that grouped both of these calls
	data.log.Info("Starting new deployment to Azure", "action", string(action), "id", deployment.Id)
	gr.Recorder.Event(data.metaObj, v1.EventTypeNormal, string(action), fmt.Sprintf("Starting new deployment to Azure with ID %q", deployment.Id))

	err = gr.KubeClient.PatchHelper(ctx, data.metaObj, func(ctx context.Context, mutObj genruntime.MetaObject) error {
		deployment, err = gr.ARMClient.CreateDeployment(ctx, deployment)
		if err != nil {
			return err
		}

		err = updateMetaObject(mutObj, deployment, nil) // Status is always nil here
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
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
func (gr *GenericReconciler) WatchDeployment(ctx context.Context, action ReconcileAction, data *ReconcileMetadata) (ctrl.Result, error) {
	deployment, err := gr.resourceSpecToDeployment(ctx, data)
	if err != nil {
		return ctrl.Result{}, err
	}

	var status genruntime.ArmTransformer
	err = gr.KubeClient.PatchHelper(ctx, data.metaObj, func(ctx context.Context, mutObj genruntime.MetaObject) error {

		deployment, err = gr.ARMClient.GetDeployment(ctx, deployment.Id)
		if err != nil {
			return errors.Wrapf(err, "failed getting deployment %q from ARM", deployment.Id)
		}

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

		err = updateMetaObject(mutObj, deployment, status)
		if err != nil {
			return errors.Wrap(err, "failed updating metaObj")
		}

		return nil
	})

	if err != nil {
		// This is a superfluous error as per https://github.com/kubernetes-sigs/controller-runtime/issues/377
		// The correct handling is just to ignore it and we will get an event shortly with the updated version to patch
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, errors.Wrap(err, "failed to patch")
	}

	// TODO: Could somehow have a method that grouped both of these calls
	currentState := genruntime.GetResourceProvisioningStateOrDefault(data.metaObj)
	data.log.Info("Monitoring deployment", "action", string(action), "id", deployment.Id, "state", currentState)
	gr.Recorder.Event(data.metaObj, v1.EventTypeNormal, string(action), fmt.Sprintf("Monitoring Azure deployment ID=%q, state=%q", deployment.Id, currentState))

	// We do two patches here because if we remove the deployment before we've actually confirmed we persisted
	// the resource ID, then we will be unable to get the resource ID the next time around. Only once we have
	// persisted the resource ID can we safely delete the deployment
	if deployment.IsTerminalProvisioningState() && !genruntime.GetShouldPreserveDeployment(data.metaObj) {
		data.log.Info("Deleting deployment", "ID", deployment.Id)
		err = gr.KubeClient.PatchHelper(ctx, data.metaObj, func(ctx context.Context, mutObj genruntime.MetaObject) error {
			err := gr.ARMClient.DeleteDeployment(ctx, deployment.Id)
			if err != nil {
				return errors.Wrapf(err, "failed to delete deployment %q", deployment.Id)
			}
			deployment.Id = ""
			deployment.Name = ""

			err = updateMetaObject(mutObj, deployment, status)
			if err != nil {
				return errors.Wrap(err, "failed updating metaObj")
			}

			return nil
		})

		if err != nil {
			// This is a superfluous error as per https://github.com/kubernetes-sigs/controller-runtime/issues/377
			// The correct handling is just to ignore it and we will get an event shortly with the updated version to patch
			if apierrors.IsNotFound(err) {
				return ctrl.Result{}, nil
			}

			return ctrl.Result{}, errors.Wrap(err, "failed to patch")
		}
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

//////////////////////////////////////////
// Other helpers
//////////////////////////////////////////

func (gr *GenericReconciler) constructArmResource(ctx context.Context, obj *ReconcileMetadata) (genruntime.ArmResource, error) {
	_, armResourceSpec, err := ResourceSpecToArmResourceSpec(ctx, gr.ResourceResolver, obj.metaObj)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't convert to armResourceSpec")
	}
	// TODO: Not setting status here
	resource := genruntime.NewArmResource(armResourceSpec, nil, genruntime.GetResourceId(obj.metaObj))

	return resource, nil
}

func (gr *GenericReconciler) getStatus(ctx context.Context, id string, data *ReconcileMetadata) (genruntime.ArmTransformer, error) {
	_, typedArmSpec, err := ResourceSpecToArmResourceSpec(ctx, gr.ResourceResolver, data.metaObj)
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

func (gr *GenericReconciler) resourceSpecToDeployment(ctx context.Context, data *ReconcileMetadata) (*armclient.Deployment, error) {
	resourceGroupName, typedArmSpec, err := ResourceSpecToArmResourceSpec(ctx, gr.ResourceResolver, data.metaObj)
	if err != nil {
		return nil, err
	}

	// TODO: get other deployment details from status and avoid creating a new deployment
	deploymentId, deploymentIdOk := genruntime.GetDeploymentId(data.metaObj)
	deploymentName, deploymentNameOk := genruntime.GetDeploymentName(data.metaObj)
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
