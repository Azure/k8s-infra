/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package controllers

import (
	"github.com/Azure/k8s-infra/hack/generated/pkg/armclient"
	"github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// TODO: Anybody have a better name?
type ReconcileMetadata struct {
	log                       logr.Logger
	metaObj                   genruntime.MetaObject
	resourceProvisioningState *armclient.ProvisioningState
	preserveDeployment        bool
	deploymentId              string
}

func NewReconcileMetadata(metaObj genruntime.MetaObject, log logr.Logger) *ReconcileMetadata {

	var provisioningState *armclient.ProvisioningState
	stateStr := genruntime.GetResourceProvisioningStateOrDefault(metaObj)
	if stateStr != "" {
		s := armclient.ProvisioningState(stateStr)
		provisioningState = &s
	}

	return &ReconcileMetadata{
		metaObj:                   metaObj,
		log:                       log,
		resourceProvisioningState: provisioningState,
		preserveDeployment:        genruntime.GetShouldPreserveDeployment(metaObj),
		deploymentId:              genruntime.GetDeploymentIdOrDefault(metaObj),
	}
}

func (r *ReconcileMetadata) IsTerminalProvisioningState() bool {
	return r.resourceProvisioningState != nil && armclient.IsTerminalProvisioningState(*r.resourceProvisioningState)
}

func updateMetaObject(
	metaObj genruntime.MetaObject,
	deployment *armclient.Deployment,
	status genruntime.ArmTransformer) error {

	controllerutil.AddFinalizer(metaObj, GenericControllerFinalizer)

	sig, err := genruntime.SpecSignature(metaObj)
	if err != nil {
		return errors.Wrap(err, "failed to compute resource spec hash")
	}

	genruntime.SetDeploymentId(metaObj, deployment.Id)
	genruntime.SetDeploymentName(metaObj, deployment.Name)
	// TODO: Do we want to just use Azure's annotations here? I bet we don't? We probably want to map
	// TODO: them onto something more robust? For now just use Azure's though.
	genruntime.SetResourceState(metaObj, string(deployment.Properties.ProvisioningState))
	genruntime.SetResourceSignature(metaObj, sig)
	if deployment.IsTerminalProvisioningState() {
		if deployment.Properties.ProvisioningState == armclient.FailedProvisioningState {
			genruntime.SetResourceError(metaObj, deployment.Properties.Error.String())
		} else if len(deployment.Properties.OutputResources) > 0 {
			resourceId := deployment.Properties.OutputResources[0].ID
			genruntime.SetResourceId(metaObj, resourceId)

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

	return nil
}
