/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package testcommon

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	batch "github.com/Azure/k8s-infra/hack/generated/apis/microsoft.batch/v20170901"
	resources "github.com/Azure/k8s-infra/hack/generated/apis/microsoft.resources/v20200601"
	storage "github.com/Azure/k8s-infra/hack/generated/apis/microsoft.storage/v20190401"
	"github.com/Azure/k8s-infra/hack/generated/pkg/armclient"
)

type KubeTestContext struct {
	TestContext
	KubeClient client.Client
	Ensure     *Ensure
	Match      *KubeMatcher

	Namespace string
}

func (ktc KubeTestContext) ForTestName(name string) KubePerTestContext {
	return KubePerTestContext{
		KubeTestContext: ktc,
		Namer:           ktc.NameConfig.NewResourceNamer(name),
	}
}

func (ktc KubeTestContext) ForTest(t *testing.T) KubePerTestContext {
	return ktc.ForTestName(t.Name())
}

type KubePerTestContext struct {
	KubeTestContext
	Namer ResourceNamer
}

// TODO: State Annotation parameter should be removed once the interface for Status determined and promoted
// TODO: to genruntime. Same for errorAnnotation
func NewKubeTestContext(
	config *rest.Config,
	region string,
	namespace string,
	armClient armclient.Applier,
	stateAnnotation string,
	errorAnnotation string) (*KubeTestContext, error) {

	clientOpts := client.Options{
		Scheme: CreateScheme(),
	}

	kubeClient, err := client.New(config, clientOpts)
	if err != nil {
		return nil, errors.Wrapf(err, "creating kubeclient")
	}

	ensure := NewEnsure(kubeClient, stateAnnotation, errorAnnotation)

	testContext, err := NewTestContext(region, armClient)
	if err != nil {
		return nil, err
	}

	return &KubeTestContext{
		TestContext: *testContext,
		Ensure:      ensure,
		Match:       NewKubeMatcher(ensure),
		KubeClient:  kubeClient,
		Namespace:   namespace,
	}, nil
}

func (tc *KubeTestContext) CreateTestNamespace() error {
	ctx := context.Background()

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: tc.Namespace,
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, tc.KubeClient, ns, func() error {
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "creating namespace")
	}

	return nil
}

func (tc KubePerTestContext) MakeObjectMeta(prefix string) ctrl.ObjectMeta {
	return ctrl.ObjectMeta{
		Name:      tc.Namer.GenerateName(prefix),
		Namespace: tc.Namespace,
	}
}

func (tc KubePerTestContext) MakeObjectMetaWithName(name string) ctrl.ObjectMeta {
	return ctrl.ObjectMeta{
		Name:      name,
		Namespace: tc.Namespace,
	}
}

func (tc KubePerTestContext) NewTestResourceGroup() *resources.ResourceGroup {
	return &resources.ResourceGroup{
		ObjectMeta: tc.MakeObjectMeta("rg"),
		Spec: resources.ResourceGroupSpec{
			Location: tc.AzureRegion,
			// This tag is used for cleanup optimization
			Tags: CreateTestResourceGroupDefaultTags(),
		},
	}
}

func CreateTestResourceGroupDefaultTags() map[string]string {
	return map[string]string{"CreatedAt": time.Now().UTC().Format(time.RFC3339)}
}

// TODO: Code generate this (I think I already do in another branch)
func CreateScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()

	_ = clientgoscheme.AddToScheme(scheme)
	_ = batch.AddToScheme(scheme)
	_ = storage.AddToScheme(scheme)
	_ = resources.AddToScheme(scheme)

	return scheme
}
