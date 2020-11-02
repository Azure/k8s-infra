/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package testcommon

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	resources "github.com/Azure/k8s-infra/hack/generated/apis/microsoft.resources/v20200601"
	"github.com/Azure/k8s-infra/hack/generated/pkg/armclient"
)

// If you modify this make sure to modify the cleanup-test-azure-resources target in the Makefile too
const ResourcePrefix = "k8sinfratest"
const DefaultTestRegion = "westus" // Could make this an env variable if we wanted

type TestContext struct {
	NameConfig  *ResourceNameConfig
	AzureClient armclient.Applier

	AzureRegion string
}

func (tc TestContext) ForTest(t *testing.T) ArmPerTestContext {
	return ArmPerTestContext{
		TestContext: tc,
		Namer:       tc.NameConfig.NewResourceNamer(t.Name()),
	}
}

type ArmPerTestContext struct {
	TestContext
	Namer ResourceNamer
}

func NewTestContext(region string, armClient armclient.Applier) (*TestContext, error) {
	return &TestContext{
		AzureClient: armClient,
		AzureRegion: region,
		NameConfig:  NewResourceNameConfig(ResourcePrefix, "-", 6),
	}, nil
}

func (tc ArmPerTestContext) NewTestResourceGroup() *resources.ResourceGroup {
	return &resources.ResourceGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name: tc.Namer.GenerateName("rg"),
		},
		Spec: resources.ResourceGroupSpec{
			Location: tc.AzureRegion,
			Tags:     CreateTestResourceGroupDefaultTags(),
		},
	}
}
