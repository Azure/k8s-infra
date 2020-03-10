package xform

import (
	"context"
	"testing"

	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	azcorev1 "github.com/Azure/k8s-infra/apis/core/v1"
	microsoftnetworkv1 "github.com/Azure/k8s-infra/apis/microsoft.network/v1"
	microsoftresourcesv1 "github.com/Azure/k8s-infra/apis/microsoft.resources/v1"
	"github.com/Azure/k8s-infra/internal/test"
)

type (
	MockClient struct {
		mock.Mock
	}

	MockStatusWriter struct {
		mock.Mock
	}
)

func TestARMConverter_ToResource(t *testing.T) {
	randomName := test.RandomName("foo", 10)
	nn := &client.ObjectKey{
		Namespace: "default",
		Name:      randomName,
	}

	group := newResourceGroup(nn)
	route := newRoute(nn)
	routeTable := newRouteTable(nn)
	routeTable.Spec.ResourceGroupRef = &azcorev1.KnownTypeReference{
		Name:      group.Name,
		Namespace: group.Namespace,
	}

	mc := new(MockClient)
	mc.On("Get", mock.Anything, client.ObjectKey{
		Namespace: route.Namespace,
		Name:      route.Name,
	}, new(microsoftnetworkv1.Route)).Run(func(args mock.Arguments) {
		dst := args.Get(2).(*microsoftnetworkv1.Route)
		route.DeepCopyInto(dst)
	}).Return(nil)

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = microsoftnetworkv1.AddToScheme(scheme)
	converter := NewARMConverter(mc, scheme)
	res, err := converter.ToResource(context.TODO(), routeTable)
	g := gomega.NewGomegaWithT(t)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(res).ToNot(gomega.BeNil())
}

func newRoute(nn *client.ObjectKey) *microsoftnetworkv1.Route {
	return &microsoftnetworkv1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RouteTable",
			APIVersion: microsoftnetworkv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nn.Name + "_route",
			Namespace: nn.Namespace,
		},
		Spec: microsoftnetworkv1.RouteSpec{
			APIVersion: "2019-11-01",
			Properties: &microsoftnetworkv1.RouteSpecProperties{
				AddressPrefix: "10.0.0.0/24",
				NextHopType:   "VnetLocal",
			},
		},
	}
}

func newRouteTable(nn *client.ObjectKey) *microsoftnetworkv1.RouteTable {
	return &microsoftnetworkv1.RouteTable{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RouteTable",
			APIVersion: microsoftnetworkv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nn.Name,
			Namespace: nn.Namespace,
		},
		Spec: microsoftnetworkv1.RouteTableSpec{
			Location:   "westus2",
			APIVersion: "2019-11-01",
			ResourceGroupRef: &azcorev1.KnownTypeReference{
				Name:      nn.Name,
				Namespace: nn.Namespace,
			},
			Properties: &microsoftnetworkv1.RouteTableSpecProperties{
				DisableBGPRoutePropagation: false,
				RouteRefs:                  []azcorev1.KnownTypeReference{},
			},
		},
	}
}

func newResourceGroup(nn *client.ObjectKey) *microsoftresourcesv1.ResourceGroup {
	return &microsoftresourcesv1.ResourceGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ResourceGroup",
			APIVersion: microsoftresourcesv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nn.Name + "_rg",
			Namespace: nn.Namespace,
		},
		Spec: microsoftresourcesv1.ResourceGroupSpec{
			APIVersion: "2019-10-01",
			Location:   "westus2",
		},
	}
}

func (mc *MockClient) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	args := mc.Called(ctx, key, obj)
	return args.Error(0)
}

func (mc *MockClient) List(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
	args := mc.Called(ctx, list, opts)
	return args.Error(0)
}

func (mc *MockClient) Status() client.StatusWriter {
	args := mc.Called()
	return args.Get(0).(client.StatusWriter)
}

func (mc *MockClient) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	args := mc.Called(ctx, obj, opts)
	return args.Error(0)
}

func (mc *MockClient) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	args := mc.Called(ctx, obj, opts)
	return args.Error(0)
}

func (mc *MockClient) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	args := mc.Called(ctx, obj, opts)
	return args.Error(0)
}

func (mc *MockClient) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	args := mc.Called(ctx, obj, patch, opts)
	return args.Error(0)
}

func (mc *MockClient) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...client.DeleteAllOfOption) error {
	args := mc.Called(ctx, obj, opts)
	return args.Error(0)
}

func (msw *MockStatusWriter) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	args := msw.Called(ctx, obj, opts)
	return args.Error(0)
}

func (msw *MockStatusWriter) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	args := msw.Called(ctx, obj, patch, opts)
	return args.Error(0)
}
