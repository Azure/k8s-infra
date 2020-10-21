/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package controllers_test

import (
	"context"
	"log"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"

	resources "github.com/Azure/k8s-infra/hack/generated/apis/microsoft.resources/v20200601"
	"github.com/Azure/k8s-infra/hack/generated/controllers"
	"github.com/Azure/k8s-infra/hack/generated/pkg/testcommon"
)

const (
	TestRegion    = "westus"
	TestNamespace = "k8s-infra-test-ns"
)

type ControllerTestContext struct {
	*testcommon.TestContext
	SharedResourceGroup *resources.ResourceGroup
}

func setup() (ControllerTestContext, error) {
	ctx := context.Background()
	log.Println("Running test setup")

	SetDefaultEventuallyTimeout(2 * time.Minute)
	SetDefaultEventuallyPollingInterval(5 * time.Second)

	var result ControllerTestContext
	var err error

	testContext, err := testcommon.NewTestContext(
		TestRegion,
		TestNamespace,
		controllers.ResourceStateAnnotation,
		controllers.ResourceErrorAnnotation)

	if err != nil {
		return result, err
	}

	err = testContext.CreateTestNamespace()
	if err != nil {
		return result, err
	}

	// Create a shared resource group, for tests to use
	sharedResourceGroup := testContext.NewTestResourceGroup()
	err = testContext.KubeClient.Create(ctx, sharedResourceGroup)
	if err != nil {
		return result, errors.Wrapf(err, "creating shared resource group")
	}

	// TODO: Should use AzureName here once it's always set
	log.Printf("Created shared resource group %s\n", sharedResourceGroup.Name)

	// It should be created in Kubernetes
	// TODO: Generic polling helper needed here
	// g.Eventually(testContext.Ensure.Created(ctx, rg)).Should(BeTrue())

	log.Println("Done with test setup")

	result.TestContext = testContext
	result.SharedResourceGroup = sharedResourceGroup

	return result, nil
}

func teardown(testContext ControllerTestContext) error {
	log.Println("Started common controller test teardown")

	ctx := context.Background()

	// List all of the resource groups
	rgList := &resources.ResourceGroupList{}
	err := testContext.KubeClient.List(ctx, rgList)
	if err != nil {
		return errors.Wrap(err, "listing resource groups")
	}

	// Delete any leaked resource groups
	var errs []error

	var resourceGroups []runtime.Object

	for _, rg := range rgList.Items {
		rg := rg // Copy so that we can safely take addr
		resourceGroups = append(resourceGroups, &rg)
		err := testContext.KubeClient.Delete(ctx, &rg)
		if err != nil {
			errs = append(errs, err)
		}
	}
	err = kerrors.NewAggregate(errs)
	if err != nil {
		return err
	}

	// Don't block forever waiting for delete to complete
	waitCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	allDeleted := false

	for !allDeleted {
		select {
		case <-waitCtx.Done():
			return waitCtx.Err()
		default:
			allDeleted, err = testContext.Ensure.AllDeleted(waitCtx, resourceGroups)()
			if err != nil {
				return err
			}
		}
	}

	log.Println("Finished common controller test teardown")
	return nil
}

type controllerTest struct {
	name string
	test func(ctx ControllerTestContext, t *testing.T)
}

func runTests(t *testing.T, tests []controllerTest) {
	ctx, setupErr := setup()
	if setupErr != nil {
		panic(setupErr)
	}

	defer func() {
		if teardownErr := teardown(ctx); teardownErr != nil {
			panic(teardownErr)
		}
	}()

	for _, test := range tests {
		if !t.Run(test.name, func(t *testing.T) { test.test(ctx, t) }) {
			return
		}
	}
}

func Test_Controller_Integrations_Slow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping slow tests in short mode")
	}

	tests := []controllerTest{
		{"ResourceGroup CRUD", Integration_ResourceGroup_CRUD},
		{"StorageAccount CRUD", Integration_StorageAccount_CRUD},
	}

	runTests(t, tests)
}

func Test_Controller_Integrations_Fast(t *testing.T) {
	// none, yet
}

// TODO: Do we need this?
//func PanicRecover(t *testing.T) {
//	if err := recover(); err != nil {
//		t.Logf("caught panic in test: %v", err)
//		t.Logf("stacktrace from panic: \n%s", string(debug.Stack()))
//		t.Fail()
//	}
//}
