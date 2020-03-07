/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package v1

import (
	"github.com/Azure/k8s-infra/pkg/zips"
)

func (rg *ResourceGroup) ToResource() (zips.Resource, error) {
	res := zips.Resource{
		ID:                rg.Status.ID,
		DeploymentID:      rg.Status.DeploymentID,
		Type:              "Microsoft.Resources/resourceGroups",
		Name:              rg.Name,
		APIVersion:        rg.Spec.APIVersion,
		Location:          rg.Spec.Location,
		Tags:              rg.Spec.Tags,
		ManagedBy:         rg.Spec.ManagedBy,
		ProvisioningState: zips.ProvisioningState(rg.Status.ProvisioningState),
	}

	return *res.SetAnnotations(rg.Annotations), nil
}

func (rg *ResourceGroup) FromResource(res zips.Resource) error {
	rg.Status.ID = res.ID
	rg.Status.ProvisioningState = string(res.ProvisioningState)
	rg.Status.DeploymentID = res.DeploymentID
	rg.Spec.Tags = res.Tags
	rg.Spec.ManagedBy = res.ManagedBy
	return nil
}
