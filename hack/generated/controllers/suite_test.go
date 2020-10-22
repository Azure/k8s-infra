/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package controllers_test

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	resources "github.com/Azure/k8s-infra/hack/generated/apis/microsoft.resources/v20200601"
	"github.com/Azure/k8s-infra/hack/generated/controllers"
	"github.com/Azure/k8s-infra/hack/generated/pkg/testcommon"
)

const (
	TestRegion    = "westus"
	TestNamespace = "k8s-infra-test-ns"
)

var testContext ControllerTestContext

type ControllerTestContext struct {
	*testcommon.TestContext
	SharedResourceGroup *resources.ResourceGroup
}

func setup() (ControllerTestContext, error) {
	ctx := context.Background()
	log.Println("Running test setup")

	SetDefaultEventuallyTimeout(2 * time.Minute)
	SetDefaultEventuallyPollingInterval(5 * time.Second)

	testContext, err := testcommon.NewTestContext(
		TestRegion,
		TestNamespace,
		controllers.ResourceStateAnnotation,
		controllers.ResourceErrorAnnotation)

	if err != nil {
		return ControllerTestContext{}, err
	}

	err = testContext.CreateTestNamespace()
	if err != nil {
		return ControllerTestContext{}, err
	}

	// Create a shared resource group, for tests to use
	sharedResourceGroup := testContext.NewTestResourceGroup()
	err = testContext.KubeClient.Create(ctx, sharedResourceGroup)
	if err != nil {
		return ControllerTestContext{}, errors.Wrapf(err, "creating shared resource group")
	}

	// TODO: Should use AzureName here once it's always set
	log.Printf("Created shared resource group %s\n", sharedResourceGroup.Name)

	// It should be created in Kubernetes
	// TODO: Generic polling helper needed here
	// g.Eventually(testContext.Ensure.Created(ctx, rg)).Should(BeTrue())

	log.Println("Done with test setup")

	return ControllerTestContext{testContext, sharedResourceGroup}, nil
}

func teardown(testContext ControllerTestContext) error {
	log.Println("Started common controller test teardown")

	ctx := context.Background()

	// List all of the resource groups
	rgList := &resources.ResourceGroupList{}
	err := testContext.KubeClient.List(ctx, rgList, &client.ListOptions{Namespace: testContext.Namespace})
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

// testMainWrapper is a wrapper that can be called by TestMain so that we can use defer
func testMainWrapper(m *testing.M) int {

	flag.Parse()
	if testing.Short() {
		log.Println("Skipping slow tests in short mode")
		return 0
	}

	ctx, setupErr := setup()
	// Save context globally as well for use by tests
	testContext = ctx
	if setupErr != nil {
		panic(setupErr)
	}

	defer func() {
		if teardownErr := teardown(ctx); teardownErr != nil {
			panic(teardownErr)
		}
	}()

	return m.Run()
}

func TestMain(m *testing.M) {
	os.Exit(testMainWrapper(m))

}

// TODO: Do we need this?
//func PanicRecover(t *testing.T) {
//	if err := recover(); err != nil {
//		t.Logf("caught panic in test: %v", err)
//		t.Logf("stacktrace from panic: \n%s", string(debug.Stack()))
//		t.Fail()
//	}
//}
