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

	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	resources "github.com/Azure/k8s-infra/hack/generated/apis/microsoft.resources/v20200601"
	"github.com/Azure/k8s-infra/hack/generated/controllers"
	"github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
	"github.com/Azure/k8s-infra/hack/generated/pkg/testcommon"
)

const (
	TestNamespace          = "k8s-infra-test-ns"
	DefaultTestRegion      = "westus" // Could make this an env variable if we wanted
	DefaultResourceTimeout = 2 * time.Minute
)

var testContext ControllerGlobalContext

// Augment main Kube context with resource group
type ControllerGlobalContext struct {
	testcommon.KubeGlobalContext

	// a pseudo-test-context used by setup/teardown
	setupTestContext    testcommon.KubePerTestContext
	sharedResourceGroup *resources.ResourceGroup
}

func (ctx ControllerGlobalContext) ForTestName(testName string) (ControllerPerTestContext, error) {
	baseCtx, err := ctx.KubeGlobalContext.ForTestName(testName)
	if err != nil {
		return ControllerPerTestContext{}, err
	}

	return ControllerPerTestContext{
		KubePerTestContext:  baseCtx,
		SharedResourceGroup: ctx.sharedResourceGroup,
	}, nil
}

func (ctx ControllerGlobalContext) ForTest(t *testing.T) (ControllerPerTestContext, error) {
	return ctx.ForTestName(t.Name())
}

type ControllerPerTestContext struct {
	testcommon.KubePerTestContext
	SharedResourceGroup *resources.ResourceGroup
}

func (tc ControllerPerTestContext) SharedResourceGroupOwner() genruntime.KnownResourceReference {
	return genruntime.KnownResourceReference{Name: tc.SharedResourceGroup.Name}
}

func setup(options Options) error {
	ctx := context.Background()
	log.Println("Running test setup")

	gomega.SetDefaultEventuallyTimeout(DefaultResourceTimeout)
	gomega.SetDefaultEventuallyPollingInterval(5 * time.Second)

	kubeContext := testcommon.NewKubeContext(
		options.useEnvTest,
		options.recordReplay,
		TestNamespace,
		DefaultTestRegion,
		controllers.ResourceStateAnnotation,
		controllers.ResourceErrorAnnotation)

	setupCtx, err := kubeContext.ForTestName("setup_and_teardown")
	if err != nil {
		return err
	}

	// Create a shared resource group, for tests to use
	// we fake the test name here:
	sharedResourceGroup := setupCtx.NewTestResourceGroup()
	err = setupCtx.KubeClient.Create(ctx, sharedResourceGroup)
	if err != nil {
		return errors.Wrapf(err, "creating shared resource group")
	}

	// TODO: Should use AzureName rather than Name once it's always set
	log.Printf("Created shared resource group '%s'\n", sharedResourceGroup.Name)

	// It should be created in Kubernetes
	err = testcommon.WaitFor(ctx, DefaultResourceTimeout, func(waitCtx context.Context) (bool, error) {
		return setupCtx.Ensure.Provisioned(waitCtx, sharedResourceGroup)
	})

	if err != nil {
		return errors.Wrapf(err, "waiting for shared resource group")
	}

	// set global context var
	testContext = ControllerGlobalContext{
		KubeGlobalContext:   kubeContext,
		setupTestContext:    setupCtx,
		sharedResourceGroup: sharedResourceGroup,
	}

	log.Print("Done with test setup")

	return nil
}

func teardown() error {
	log.Println("Started common controller test teardown")

	ctx := context.Background()

	setupCtx := testContext.setupTestContext
	kubeClient := setupCtx.KubeClient

	// List all of the resource groups
	rgList := &resources.ResourceGroupList{}
	err := setupCtx.KubeClient.List(ctx, rgList, &client.ListOptions{Namespace: setupCtx.Namespace()})
	if err != nil {
		return errors.Wrap(err, "listing resource groups")
	}

	// Delete any leaked resource groups
	var errs []error

	var resourceGroups []runtime.Object

	for _, rg := range rgList.Items {
		rg := rg // Copy so that we can safely take addr
		resourceGroups = append(resourceGroups, &rg)
		err := kubeClient.Delete(ctx, &rg)
		if err != nil {
			errs = append(errs, err)
		}
	}
	err = kerrors.NewAggregate(errs)
	if err != nil {
		return err
	}

	// Don't block forever waiting for delete to complete
	err = testcommon.WaitFor(ctx, DefaultResourceTimeout, func(waitCtx context.Context) (bool, error) {
		return setupCtx.Ensure.AllDeleted(waitCtx, resourceGroups)
	})

	if err != nil {
		return errors.Wrapf(err, "waiting for all resource groups to delete")
	}

	setupCtx.Cleanup()

	log.Println("Finished common controller test teardown")
	return nil
}

func TestMain(m *testing.M) {
	options := getOptions()
	os.Exit(testcommon.SetupTeardownTestMain(
		m,
		true,
		func() error {
			return setup(options)
		},
		teardown))
}

type Options struct {
	useEnvTest   bool
	recordReplay bool
}

func getOptions() Options {
	options := Options{}
	flag.BoolVar(&options.useEnvTest, "envtest", false, "Use the envtest package to run tests? If not, a cluster must be configured already in .kubeconfig.")
	flag.BoolVar(&options.recordReplay, "recordReplay", false, "Record/replay the ARM requests/responses.")
	flag.Parse()
	return options
}
