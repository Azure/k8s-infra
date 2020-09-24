/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package armclient

import (
	"encoding/json"
	"fmt"

	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/k8s-infra/pkg/zips/duration"
	"github.com/pkg/errors"
)

type (
	ARMMeta struct {
		Name           string `json:"name,omitempty"`
		Type           string `json:"type,omitempty"`
		Id             string `json:"id,omitempty"` // TODO: This is filled out on response to us after PUT
		Location       string `json:"location,omitempty"`
		ResourceGroup  string `json:"-"` // TODO: I feel like these should be being serialized?
		SubscriptionId string `json:"-"`
	}

	ProvisioningState string

	DeploymentScope string

	DetailLevel string

	DeploymentMode string

	DebugSetting struct {
		DetailLevel DetailLevel `json:"detailLevel,omitempty"`
	}

	OutputResource struct {
		ID string `json:"id,omitempty"`
	}

	DeploymentStatus struct {
		ProvisioningState ProvisioningState `json:"provisioningState,omitempty"`
		Timestamp         *date.Time        `json:"timestamp,omitempty"`
		Duration          *duration.ISO8601 `json:"duration,omitempty"`
		CorrelationID     string            `json:"correlationId,omitempty"`
		Outputs           json.RawMessage   `json:"outputs,omitempty"` // TODO: What is this for?
		OutputResources   []OutputResource  `json:"outputResources,omitempty"`
	}

	DeploymentSpec struct {
		DebugSetting *DebugSetting  `json:"debugSetting,omitempty"`
		Mode         DeploymentMode `json:"mode,omitempty"`
		Template     *Template      `json:"template,omitempty"`
	}

	DeploymentProperties struct {
		DeploymentStatus `json:",inline"`
		DeploymentSpec   `json:",inline"`
	}

	Deployment struct {
		ARMMeta    `json:",inline"`
		Scope      DeploymentScope `json:"-"`
		Properties *DeploymentProperties

		PreserveDeployment      bool `json:"-"`
	}
)

const (
	ResourceGroupScope DeploymentScope = "resourceGroup"
	SubscriptionScope  DeploymentScope = "subscription"

	RequestContentDetailLevel            DetailLevel = "requestContent"
	ResponseContentDetailLevel           DetailLevel = "responseContent"
	RequestAndResponseContentDetailLevel DetailLevel = "requestContent,responseContent"

	IncrementalDeploymentMode DeploymentMode = "Incremental"
	CompleteDeploymentMode    DeploymentMode = "Complete"

	SucceededProvisioningState ProvisioningState = "Succeeded"
	FailedProvisioningState    ProvisioningState = "Failed"
	DeletingProvisioningState  ProvisioningState = "Deleting"
	AcceptedProvisioningState  ProvisioningState = "Accepted"
)

func NewResourceGroupDeployment(
	subscriptionId string,
	groupName string,
	deploymentName string,
	resources ...interface{}) *Deployment{

	return &Deployment{
		Scope: ResourceGroupScope,
		Properties: &DeploymentProperties{
			DeploymentSpec: DeploymentSpec{
				DebugSetting: &DebugSetting{
					DetailLevel: RequestAndResponseContentDetailLevel,
				},
				Mode: IncrementalDeploymentMode,
				Template: &Template{
					Schema:         "https://schema.management.azure.com/schemas/2015-01-01/deploymentTemplate.json#",
					ContentVersion: "1.0.0.0",
					Resources:      resources,
				},
			},
		},
		ARMMeta: ARMMeta{
			Name:           deploymentName,
			ResourceGroup:  groupName,
			SubscriptionId: subscriptionId,
		},
	}
}

func NewSubscriptionDeployment(
	subscriptionID string,
	location string,
	deploymentName string,
	resources ...interface{}) *Deployment {

	return &Deployment{
		Scope: SubscriptionScope,
		Properties: &DeploymentProperties{
			DeploymentSpec: DeploymentSpec{
				DebugSetting: &DebugSetting{
					DetailLevel: RequestAndResponseContentDetailLevel,
				},
				Mode: IncrementalDeploymentMode,
				Template: &Template{
					Schema:         "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#",
					ContentVersion: "1.0.0.0",
					Resources:      resources,
				},
			},
		},
		ARMMeta: ARMMeta{
			Name:           deploymentName,
			Location:       location,
			SubscriptionId: subscriptionID,
		},
	}
}

func (d *Deployment) GetEntityPath() (string, error) {
	if err := d.Validate(); err != nil {
		return "", err
	}

	var entityPath string
	switch d.Scope {
	case SubscriptionScope:
		entityPath = fmt.Sprintf(
			"subscriptions/%s/providers/Microsoft.Resources/deployments/%s?api-version=2019-10-01",
			d.SubscriptionId,
			d.Name)
	case ResourceGroupScope:
		entityPath = fmt.Sprintf(
			"subscriptions/%s/resourcegroups/%s/providers/Microsoft.Resources/deployments/%s?api-version=2019-10-01",
			d.SubscriptionId,
			d.ResourceGroup,
			d.Name)
	default:
		return "", errors.Errorf("unknown scope %s", d.Scope)
	}

	return entityPath, nil
}

func (d *Deployment) Validate() error {
	switch d.Scope {
	case SubscriptionScope:
		if d.SubscriptionId == "" || d.Name == "" {
			return errors.Errorf("validate: require subscription ID and name to not be empty")
		}
	case ResourceGroupScope:
		if d.SubscriptionId == "" || d.Name == "" || d.ResourceGroup == "" {
			return errors.Errorf("validate: require subscription ID, name and resource group to not be empty")
		}
	}

	return nil
}

func (d *Deployment) IsTerminalProvisioningState() bool {
	return d.Properties != nil && IsTerminalProvisioningState(d.Properties.ProvisioningState)
}

func (d *Deployment) ProvisioningStateOrUnknown() string {
	if d.Properties == nil {
		return "unknown"
	}
	return string(d.Properties.ProvisioningState)
}

func idWithAPIVersion(resourceID string) string {
	return resourceID + "?api-version=2019-10-01"
}

func IsTerminalProvisioningState(state ProvisioningState) bool {
	return state == SucceededProvisioningState || state == FailedProvisioningState
}
