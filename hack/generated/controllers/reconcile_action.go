/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package controllers

import (
	"fmt"
	"github.com/Azure/k8s-infra/hack/generated/pkg/armclient"
	"github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

// TODO: Not sure that the name of this file is what we want

type ReconcileAction string

const (
	ReconcileActionNoAction        = ReconcileAction("NoAction")
	ReconcileActionBeginDeployment = ReconcileAction("BeginDeployment")
	ReconcileActionWatchDeployment = ReconcileAction("WatchDeployment")
	ReconcileActionBeginDelete     = ReconcileAction("BeginDelete")
	ReconcileActionWatchDelete     = ReconcileAction("WatchDelete")
)

// TODO: Naming
type ReconcileMetadata struct {
	log                       logr.Logger
	metaObj                   genruntime.MetaObject
	resourceProvisioningState *armclient.ProvisioningState
	preserveDeployment        bool
	deploymentId              string
}

func NewReconcileMetadata(metaObj genruntime.MetaObject, log logr.Logger) *ReconcileMetadata {
	// TODO: We could do some sort of lazy-load thing here so that we don't have to preload
	// TODO: stuff we don't need... but for now not bothering as I am not sure if there is a
	// TODO: clean generic way to do that in go and also don't know perf impact of doing it all up
	// TODO: front like we are now
	return &ReconcileMetadata{
		metaObj:                   metaObj,
		log:                       log,
		resourceProvisioningState: getResourceProvisioningState(metaObj),
		preserveDeployment:        getShouldPreserveDeployment(metaObj),
		deploymentId:              getDeploymentId(metaObj),
	}
}

func (r *ReconcileMetadata) IsTerminalProvisioningState() bool {
	return r.resourceProvisioningState != nil && armclient.IsTerminalProvisioningState(*r.resourceProvisioningState)
}

func (r *ReconcileMetadata) DetermineReconcileAction() (ReconcileAction, error) {
	if !r.metaObj.GetDeletionTimestamp().IsZero() {
		if r.resourceProvisioningState != nil && *r.resourceProvisioningState == armclient.DeletingProvisioningState {
			return ReconcileActionWatchDelete, nil
		}
		return ReconcileActionBeginDelete, nil
	} else {
		hasChanged, err := hasResourceHashAnnotationChanged(r.metaObj)
		if err != nil {
			return ReconcileActionNoAction, errors.Wrap(err, "failed comparing resource hash")
		}

		if !hasChanged && r.IsTerminalProvisioningState() {
			// TODO: Do we want to log here?
			msg := fmt.Sprintf("resource spec has not changed and resource is in terminal state: %q", *r.resourceProvisioningState)
			r.log.V(0).Info(msg)
			return ReconcileActionNoAction, nil
		}

		if r.resourceProvisioningState != nil && *r.resourceProvisioningState == armclient.DeletingProvisioningState {
			return ReconcileActionNoAction, errors.Errorf("resource is currently deleting; it can not be applied")
		}

		if r.deploymentId != "" {
			// There is an ongoing deployment we need to monitor
			return ReconcileActionWatchDeployment, nil
		}

		return ReconcileActionBeginDeployment, nil
	}
}
