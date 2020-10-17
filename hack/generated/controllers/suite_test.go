/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package controllers_test

import (
	"context"
	"log"
	"os"
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

// TODO: I don't love these package global variables but it seems to be the only way to actually use TestMain to set things up?
var testContext *testcommon.TestContext
var sharedResourceGroup *resources.ResourceGroup

func setup() error {
	ctx := context.Background()
	log.Println("Running test setup")

	SetDefaultEventuallyTimeout(2 * time.Minute)
	SetDefaultEventuallyPollingInterval(5 * time.Second)

	var err error
	testContext, err = testcommon.NewTestContext(
		TestRegion,
		TestNamespace,
		controllers.ResourceStateAnnotation,
		controllers.ResourceErrorAnnotation)
	if err != nil {
		return err
	}

	err = testContext.CreateTestNamespace()
	if err != nil {
		return nil
	}

	// Create a shared resource group, for tests to use
	sharedResourceGroup = testContext.NewTestResourceGroup()
	err = testContext.KubeClient.Create(ctx, sharedResourceGroup)
	if err != nil {
		return errors.Wrapf(err, "creating shared resource group")
	}

	// TODO: Should use AzureName here once it's always set
	log.Printf("Created shared resource group %s\n", sharedResourceGroup.Name)

	// It should be created in Kubernetes
	// TODO: Generic polling helper needed here
	// g.Eventually(testContext.Ensure.Created(ctx, rg)).Should(BeTrue())

	log.Println("Done with test setup")

	return nil
}

func teardown() error {
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

func TestMain(m *testing.M) {

	err := setup()
	if err != nil {
		panic(err)
	}

	// call flag.Parse() here if TestMain uses flags
	code := m.Run()

	err = teardown()
	if err != nil {
		panic(err)
	}

	os.Exit(code)
}

// TODO: Do we need this?
//func PanicRecover(t *testing.T) {
//	if err := recover(); err != nil {
//		t.Logf("caught panic in test: %v", err)
//		t.Logf("stacktrace from panic: \n%s", string(debug.Stack()))
//		t.Fail()
//	}
//}
