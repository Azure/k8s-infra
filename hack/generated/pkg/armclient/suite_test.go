/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package armclient_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/k8s-infra/hack/generated/pkg/armclient"
	"github.com/Azure/k8s-infra/hack/generated/pkg/testcommon"
)

const DefaultEventuallyTimeout = 3 * time.Minute

type ArmClientTestContext struct {
	*testcommon.TestContext
	Match                   *testcommon.ArmMatcher
	SharedResourceGroupName string
}

var testContext *ArmClientTestContext

func setup() error {
	log.Println("Running test setup")
	ctx := context.Background()

	gomega.SetDefaultEventuallyTimeout(DefaultEventuallyTimeout)
	gomega.SetDefaultEventuallyPollingInterval(5 * time.Second)

	authorizer, err := armclient.AuthorizerFromEnvironment()
	if err != nil {
		return err
	}

	subID := os.Getenv(auth.SubscriptionID)
	if subID == "" {
		return errors.Errorf("env var %q was not set", auth.SubscriptionID)
	}

	applier, err := armclient.NewAzureTemplateClient(authorizer, subID)
	if err != nil {
		return err
	}

	tc, err := testcommon.NewTestContext(testcommon.DefaultTestRegion, applier)
	if err != nil {
		return err
	}

	setupCtx := tc.ForTestName("setup")

	resourceGroup := setupCtx.NewTestResourceGroup()
	resourceGroupSpec, err := resourceGroup.Spec.ConvertToArm(resourceGroup.Name)
	if err != nil {
		return nil
	}

	deploymentName := setupCtx.Namer.GenerateName("deployment")
	deployment := armclient.NewSubscriptionDeployment(
		tc.AzureClient.SubscriptionID(),
		tc.AzureRegion,
		deploymentName,
		resourceGroupSpec)

	log.Printf(
		"Creating shared resource group %q (via deployment %q) in subscription %q\n",
		resourceGroup.Name,
		deploymentName,
		tc.AzureClient.SubscriptionID())

	deployment, err = tc.AzureClient.CreateDeployment(ctx, deployment)
	if err != nil {
		return errors.Wrap(err, "creating shared deployment")
	}

	err = testcommon.WaitFor(ctx, DefaultEventuallyTimeout, func(ctx context.Context) (bool, error) {
		deployment, err = tc.AzureClient.GetDeployment(ctx, deployment.Id)
		if err != nil {
			return false, err
		}

		return deployment.Properties != nil && deployment.Properties.ProvisioningState == armclient.SucceededProvisioningState, nil
	})
	if err != nil {
		return errors.Wrapf(err, "waiting for deployment %s to succeed", deploymentName)
	}

	log.Println("Done with test setup")

	testContext = &ArmClientTestContext{
		TestContext:             tc,
		Match:                   testcommon.NewArmMatcher(tc.AzureClient),
		SharedResourceGroupName: resourceGroup.Name,
	}
	return nil
}

func teardown() error {
	log.Println("Started common controller test teardown")
	ctx := context.Background()

	sharedResourceGroupId, err := armclient.MakeArmResourceId(
		testContext.AzureClient.SubscriptionID(),
		"resourceGroups",
		testContext.SharedResourceGroupName)

	if err != nil {
		return err
	}

	// TODO: Icky hardcoded API version here
	err = testContext.AzureClient.BeginDeleteResource(ctx, sharedResourceGroupId, "2020-06-01", nil)
	if err != nil {
		return errors.Wrapf(err, "deleting shared resource group %s", sharedResourceGroupId)
	}

	log.Println("Finished common controller test teardown")
	return nil
}

func TestMain(m *testing.M) {
	os.Exit(testcommon.SetupTeardownTestMain(m, true, setup, teardown))
}
