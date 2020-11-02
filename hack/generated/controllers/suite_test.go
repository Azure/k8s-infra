/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package controllers_test

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	resources "github.com/Azure/k8s-infra/hack/generated/apis/microsoft.resources/v20200601"
	"github.com/Azure/k8s-infra/hack/generated/controllers"
	"github.com/Azure/k8s-infra/hack/generated/pkg/armclient"
	"github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
	"github.com/Azure/k8s-infra/hack/generated/pkg/testcommon"
)

const (
	TestNamespace          = "k8s-infra-test-ns"
	DefaultResourceTimeout = 2 * time.Minute
)

var testContext ControllerTestContext
var envtestContext *EnvtestContext

type EnvtestContext struct {
	testenv     envtest.Environment
	manager     ctrl.Manager
	stopManager chan struct{}
	recorder    *recorder.Recorder
}

type ControllerTestContext struct {
	*testcommon.KubeTestContext
	SharedResourceGroup *resources.ResourceGroup
}

func (ctc ControllerTestContext) ForTest(t *testing.T) ControllerPerTestContext {
	return ControllerPerTestContext{
		KubePerTestContext:  ctc.KubeTestContext.ForTest(t),
		SharedResourceGroup: ctc.SharedResourceGroup,
	}
}

type ControllerPerTestContext struct {
	testcommon.KubePerTestContext
	SharedResourceGroup *resources.ResourceGroup
}

func (tc ControllerTestContext) SharedResourceGroupOwner() genruntime.KnownResourceReference {
	return genruntime.KnownResourceReference{Name: tc.SharedResourceGroup.Name}
}

func (tc ControllerPerTestContext) SharedResourceGroupOwner() genruntime.KnownResourceReference {
	return genruntime.KnownResourceReference{Name: tc.SharedResourceGroup.Name}
}

// Wraps an inner HTTP roundtripper to add a
// counter for duplicated request URIs.
type RoundTripper struct {
	inner  http.RoundTripper
	counts map[string]uint32
}

func MakeRoundTripper(inner http.RoundTripper) *RoundTripper {
	return &RoundTripper{
		inner:  inner,
		counts: make(map[string]uint32),
	}
}

var COUNT_HEADER string = "TEST-REQUEST-ATTEMPT"

func (rt *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	key := req.Method + ":" + req.URL.String()
	count := rt.counts[key]
	req.Header.Add(COUNT_HEADER, fmt.Sprintf("%v", count))
	rt.counts[key] = count + 1
	return rt.inner.RoundTrip(req)
}

var _ http.RoundTripper = &RoundTripper{}

func setupEnvTest() (*rest.Config, *armclient.AzureTemplateClient, error) {
	envtestContext = &EnvtestContext{
		testenv: envtest.Environment{
			CRDDirectoryPaths: []string{
				"../config/crd/bases/valid",
			},
		},
		stopManager: make(chan struct{}),
	}

	log.Print("Starting envtest")
	config, err := envtestContext.testenv.Start()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "starting envtest environment")
	}

	log.Print("Creating & starting controller-runtime manager")
	mgr, err := ctrl.NewManager(config, ctrl.Options{Scheme: testcommon.CreateScheme()})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "creating controller-runtime manager")
	}

	envtestContext.manager = mgr

	go func() {
		// this blocks until the input chan is closed
		err := envtestContext.manager.Start(envtestContext.stopManager)
		if err != nil {
			log.Fatal(errors.Wrapf(err, "running controller-runtime manager"))
		}
	}()

	r, err := recorder.New("fixtures/arm")
	if err != nil {
		return nil, nil, errors.Wrapf(err, "creating recorder")
	}

	var subscriptionID string
	var authorizer autorest.Authorizer
	if r.Mode() == recorder.ModeRecording {
		// if we are recording, we need auth
		authorizer, err = armclient.AuthorizerFromEnvironment()
		if err != nil {
			return nil, nil, errors.Wrapf(err, "creating authorizer")
		}

		subscriptionID = os.Getenv(auth.SubscriptionID)
		if subscriptionID == "" {
			return nil, nil, errors.Wrapf(err, "required environment variable %q was not supplied", auth.SubscriptionID)
		}
	} else {
		// if we are replaying, we won't need auth
		// and we use a dummy subscription ID
		authorizer = nil
		subscriptionID = uuid.Nil.String()
	}

	r.AddSaveFilter(func(i *cassette.Interaction) error {
		// rewrite all request/response fields to hide the real subscription ID
		// this is *not* a security measure but intended to make the tests updateable from
		// any subscription, so a contributer can update the tests against their own sub
		hideSubID := func(s string) string {
			return strings.ReplaceAll(s, subscriptionID, uuid.Nil.String())
		}

		i.Request.Body = hideSubID(i.Request.Body)
		i.Response.Body = hideSubID(i.Response.Body)
		i.Request.URL = hideSubID(i.Request.URL)

		for _, values := range i.Request.Headers {
			for i := range values {
				values[i] = hideSubID(values[i])
			}
		}

		for _, values := range i.Response.Headers {
			for i := range values {
				values[i] = hideSubID(values[i])
			}
		}

		// remove all Authorization headers from stored requests
		delete(i.Request.Headers, "Authorization")

		// remove all request IDs
		delete(i.Response.Headers, "X-Ms-Correlation-Request-Id")
		delete(i.Response.Headers, "X-Ms-Ratelimit-Remaining-Subscription-Reads")
		delete(i.Response.Headers, "X-Ms-Ratelimit-Remaining-Subscription-Writes")
		delete(i.Response.Headers, "X-Ms-Request-Id")
		delete(i.Response.Headers, "X-Ms-Routing-Request-Id")

		return nil
	})

	// request must match URI & METHOD & our custom header
	r.SetMatcher(func(request *http.Request, i cassette.Request) bool {
		return cassette.DefaultMatcher(request, i) &&
			request.Header.Get(COUNT_HEADER) == i.Headers.Get(COUNT_HEADER)
	})

	envtestContext.recorder = r

	log.Print("Creating ARM client")
	armClient, err := armclient.NewAzureTemplateClient(authorizer, subscriptionID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "creating ARM client")
	}

	// overwrite default HTTP transport on ARM client
	client, ok := armClient.RawClient.Sender.(*http.Client)
	if !ok {
		return nil, nil, errors.Errorf("unexpected type for ARM client Sender: %T", armClient.RawClient.Sender)
	}
	client.Transport = MakeRoundTripper(r)

	var deployments uint32 = 0

	var requeueDelay time.Duration // defaults to 5s when zero is passed
	if r.Mode() == recorder.ModeReplaying {
		// skip requeue delays when replaying
		requeueDelay = 100 * time.Millisecond
	}

	log.Print("Registering custom controllers")
	errs := controllers.RegisterAll(
		envtestContext.manager,
		armClient,
		controllers.KnownTypes,
		klogr.New(),
		controllers.Options{
			DeploymentNameGenerator: func() (string, error) {
				ix := atomic.AddUint32(&deployments, 1)
				bs := make([]byte, 4)
				binary.LittleEndian.PutUint32(bs, ix)
				result := uuid.NewSHA1(uuid.Nil, bs)
				return fmt.Sprintf("k8s_%s", result.String()), nil
			},
			RequeueDelay: requeueDelay,
		})

	if errs != nil {
		return nil, nil, errors.Wrapf(kerrors.NewAggregate(errs), "registering reconcilers")
	}

	return config, armClient, nil
}

func teardownEnvTest() error {
	if envtestContext != nil {
		log.Print("Stopping controller-runtime manager")
		close(envtestContext.stopManager)

		log.Print("Stopping envtest")
		err := envtestContext.testenv.Stop()
		if err != nil {
			return errors.Wrapf(err, "stopping envtest environment")
		}

		log.Print("Stopping recorder")
		err = envtestContext.recorder.Stop()
		if err != nil {
			return errors.Wrapf(err, "stopping recorder")
		}
	}

	return nil
}

func setup(options Options) error {
	ctx := context.Background()
	log.Println("Running test setup")

	gomega.SetDefaultEventuallyTimeout(DefaultResourceTimeout)
	gomega.SetDefaultEventuallyPollingInterval(5 * time.Second)

	var err error
	var config *rest.Config
	var armClient *armclient.AzureTemplateClient
	if options.useEnvTest {
		config, armClient, err = setupEnvTest()
		if err != nil {
			return errors.Wrapf(err, "setting up envtest")
		}
	} else {
		authorizer, err := armclient.AuthorizerFromEnvironment()
		if err != nil {
			return errors.Wrapf(err, "unable to get authorization settings")
		}

		subID := os.Getenv(auth.SubscriptionID)
		if subID == "" {
			return errors.Errorf("env var %q was not set", auth.SubscriptionID)
		}

		armClient, err = armclient.NewAzureTemplateClient(authorizer, subID)
		if err != nil {
			return errors.Wrapf(err, "unable to create ARM client")
		}

		config, err = ctrl.GetConfig()
		if err != nil {
			return errors.Wrapf(err, "unable to retrieve kubeconfig")
		}
	}

	newCtx, err := testcommon.NewKubeTestContext(
		config,
		testcommon.DefaultTestRegion,
		TestNamespace,
		armClient,
		controllers.ResourceStateAnnotation,
		controllers.ResourceErrorAnnotation)

	if err != nil {
		return err
	}

	err = newCtx.CreateTestNamespace()
	if err != nil {
		return err
	}

	// Create a shared resource group, for tests to use
	// we fake the test name here:
	sharedResourceGroup := newCtx.ForTestName("setup").NewTestResourceGroup()
	err = newCtx.KubeClient.Create(ctx, sharedResourceGroup)
	if err != nil {
		return errors.Wrapf(err, "creating shared resource group")
	}

	// TODO: Should use AzureName rather than Name once it's always set
	log.Printf("Created shared resource group '%s'\n", sharedResourceGroup.Name)

	// It should be created in Kubernetes
	err = testcommon.WaitFor(ctx, DefaultResourceTimeout, func(waitCtx context.Context) (bool, error) {
		return newCtx.Ensure.Provisioned(waitCtx, sharedResourceGroup)
	})

	if err != nil {
		return errors.Wrapf(err, "waiting for shared resource group")
	}

	log.Print("Done with test setup")

	testContext = ControllerTestContext{newCtx, sharedResourceGroup}

	return nil
}

func teardown() error {
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
	err = testcommon.WaitFor(ctx, DefaultResourceTimeout, func(waitCtx context.Context) (bool, error) {
		return testContext.Ensure.AllDeleted(waitCtx, resourceGroups)
	})

	if err != nil {
		return errors.Wrapf(err, "waiting for all resource groups to delete")
	}

	err = teardownEnvTest()
	if err != nil {
		return errors.Wrapf(err, "tearing down envtest")
	}

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
	useEnvTest bool
}

func getOptions() Options {
	options := Options{}
	flag.BoolVar(&options.useEnvTest, "envtest", false, "Use the envtest package to run tests? If not, a cluster must be configured already in .kubeconfig.")
	flag.Parse()
	return options
}
