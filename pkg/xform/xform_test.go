/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package xform

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Azure/go-autorest/autorest/to"

	azcorev1 "github.com/Azure/k8s-infra/apis/core/v1"
	microsoftcomputev1 "github.com/Azure/k8s-infra/apis/microsoft.compute/v1"
	microsoftnetworkv1 "github.com/Azure/k8s-infra/apis/microsoft.network/v1"
	microsoftresourcesv1 "github.com/Azure/k8s-infra/apis/microsoft.resources/v1"
	"github.com/Azure/k8s-infra/internal/test"
	"github.com/Azure/k8s-infra/pkg/util/ownerutil"
	"github.com/Azure/k8s-infra/pkg/zips"
)

type (
	MockClient struct {
		mock.Mock
	}

	FakeRouteTableProperties struct {
		DisableBGPRoutePropagation bool                  `json:"disableBGPRoutePropagation,omitempty"`
		Routes                     []FakeRouteProperties `json:"routes,omitempty"`
	}

	FakeRouteProperties struct {
		ID string `json:"id"`
	}
)

func TestARMConverter_ToResource(t *testing.T) {
	cases := []struct {
		Name   string
		Setup  func(g *gomega.WithT, kClient *MockClient) azcorev1.MetaObject
		Expect func(g *gomega.WithT, metaObject azcorev1.MetaObject, res *zips.Resource)
	}{
		{
			Name: "RouteTableAndRoutes",
			Setup: func(g *gomega.WithT, mc *MockClient) azcorev1.MetaObject {
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

				routeTable.Spec.Properties.RouteRefs = []azcorev1.KnownTypeReference{
					{
						Name:      route.Name,
						Namespace: route.Namespace,
					},
				}

				mc.On("Get", mock.Anything, client.ObjectKey{
					Namespace: route.Namespace,
					Name:      route.Name,
				}, new(microsoftnetworkv1.Route)).Run(func(args mock.Arguments) {
					dst := args.Get(2).(*microsoftnetworkv1.Route)
					route.DeepCopyInto(dst)
					dst.Status.ID = "sub/1234/blah"
				}).Return(nil)

				return routeTable
			},
			Expect: func(g *gomega.WithT, metaObject azcorev1.MetaObject, res *zips.Resource) {
				routeTableWithRouteIDs := struct {
					Name       string
					Location   string
					APIVersion string
					Type       string
					Properties struct {
						Routes []struct {
							ID string `json:"id"`
						}
					}
				}{}

				bits, _ := json.Marshal(res)
				g.Expect(json.Unmarshal(bits, &routeTableWithRouteIDs)).ToNot(gomega.HaveOccurred())
				g.Expect(routeTableWithRouteIDs.Properties.Routes).To(gomega.HaveLen(1))
				g.Expect(routeTableWithRouteIDs.Properties.Routes[0].ID).To(gomega.Equal("sub/1234/blah"))
				routeTable := metaObject.(*microsoftnetworkv1.RouteTable)
				g.Expect(routeTableWithRouteIDs.Name).To(gomega.Equal(routeTable.Name))
				g.Expect(routeTableWithRouteIDs.Location).To(gomega.Equal(routeTable.Spec.Location))
				g.Expect(routeTableWithRouteIDs.Type).To(gomega.Equal(routeTable.ResourceType()))
				g.Expect(routeTableWithRouteIDs.APIVersion).To(gomega.Equal(routeTable.Spec.APIVersion))
			},
		},
		{
			Name: "VirtualMachineWithEmbeddedNetworkInterface",
			Setup: func(g *gomega.WithT, mc *MockClient) azcorev1.MetaObject {
				randomName := test.RandomName("foo", 10)
				nn := &client.ObjectKey{
					Namespace: "default",
					Name:      randomName,
				}

				group := newResourceGroup(nn)
				vm := newVM(nn)
				nic := newNetworkInterface(nn)
				vm.Spec.Properties.NetworkProfile.NetworkInterfaceRefs = &[]microsoftcomputev1.NetworkInterfaceReference{
					{
						NetworkInterfaceReferenceProperties: &microsoftcomputev1.NetworkInterfaceReferenceProperties{
							Primary: to.BoolPtr(true),
						},
						KnownTypeReference: azcorev1.KnownTypeReference{
							Name:      nic.Name,
							Namespace: nic.Namespace,
						},
					},
				}

				vm.Spec.ResourceGroupRef = &azcorev1.KnownTypeReference{
					Name:      group.Name,
					Namespace: group.Namespace,
				}

				mc.On("Get", mock.Anything, client.ObjectKey{
					Namespace: nic.Namespace,
					Name:      nic.Name,
				}, new(microsoftnetworkv1.NetworkInterface)).Run(func(args mock.Arguments) {
					dst := args.Get(2).(*microsoftnetworkv1.NetworkInterface)
					nic.DeepCopyInto(dst)
					dst.Status.ID = "sub/1234/foo"
				}).Return(nil)

				return vm
			},
			Expect: func(g *gomega.WithT, metaObject azcorev1.MetaObject, res *zips.Resource) {
				vmWithNetworkInterfaceEmbedded := struct {
					Name       string
					Location   string
					APIVersion string
					Type       string
					Properties struct {
						NetworkProfile struct {
							NetworkInterfaces []struct {
								Properties struct {
									Primary *bool `json:"primary,omitempty"`
								}
								ID string `json:"id,omitempty"`
							} `json:"networkInterfaces"`
						}
					}
				}{}

				bits, _ := json.Marshal(res)
				g.Expect(json.Unmarshal(bits, &vmWithNetworkInterfaceEmbedded)).ToNot(gomega.HaveOccurred())
				g.Expect(vmWithNetworkInterfaceEmbedded.Properties.NetworkProfile).ToNot(gomega.BeNil())
				g.Expect(vmWithNetworkInterfaceEmbedded.Properties.NetworkProfile.NetworkInterfaces).To(gomega.HaveLen(1))
				g.Expect(vmWithNetworkInterfaceEmbedded.Properties.NetworkProfile.NetworkInterfaces[0].ID).To(gomega.Equal("sub/1234/foo"))
				vm, ok := metaObject.(*microsoftcomputev1.VirtualMachine)
				g.Expect(ok).To(gomega.BeTrue())
				g.Expect(vmWithNetworkInterfaceEmbedded.Name).To(gomega.Equal(vm.Name))
				g.Expect(vmWithNetworkInterfaceEmbedded.Location).To(gomega.Equal(vm.Spec.Location))
				g.Expect(vmWithNetworkInterfaceEmbedded.Type).To(gomega.Equal(vm.ResourceType()))
				g.Expect(vmWithNetworkInterfaceEmbedded.APIVersion).To(gomega.Equal(vm.Spec.APIVersion))
			},
		},
	}

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = microsoftnetworkv1.AddToScheme(scheme)
	_ = microsoftcomputev1.AddToScheme(scheme)
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			mc := new(MockClient)
			metaObject := c.Setup(g, mc)
			converter := NewARMConverter(mc, scheme)
			res, err := converter.ToResource(context.TODO(), metaObject)
			g.Expect(err).ToNot(gomega.HaveOccurred())
			g.Expect(res).ToNot(gomega.BeNil())
			c.Expect(g, metaObject, res)
		})
	}
}

func TestARMConverter_FromResource(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = microsoftnetworkv1.AddToScheme(scheme)
	mc := new(MockClient)
	converter := NewARMConverter(mc, scheme)

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

	routeTable.Spec.Properties.RouteRefs = []azcorev1.KnownTypeReference{
		{
			Name:      route.Name,
			Namespace: route.Namespace,
		},
	}

	props := &FakeRouteTableProperties{
		DisableBGPRoutePropagation: routeTable.Spec.Properties.DisableBGPRoutePropagation,
		Routes: []FakeRouteProperties{
			{
				ID: "fakeRouteID",
			},
		},
	}

	bits, err := json.Marshal(props)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	resource := &zips.Resource{
		ID:                "someID",
		ProvisioningState: "Accepted",
		ResourceGroup:     group.Name,
		DeploymentID:      "someDeploymentID",
		Name:              routeTable.Name,
		Location:          routeTable.Spec.Location,
		Type:              routeTable.ResourceType(),
		APIVersion:        routeTable.Spec.APIVersion,
		Properties:        bits,
	}

	err = converter.FromResource(resource, routeTable)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(routeTable.Spec.Properties.DisableBGPRoutePropagation).To(gomega.Equal(routeTable.Spec.Properties.DisableBGPRoutePropagation))
	g.Expect(routeTable.Spec.Properties.RouteRefs).To(gomega.HaveLen(1))
	g.Expect(routeTable.Spec.Properties.RouteRefs[0]).To(gomega.Equal(azcorev1.KnownTypeReference{
		Name:      route.Name,
		Namespace: route.Namespace,
	}))
}

func Test_resourceName(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	randomName := test.RandomName("foo", 10)
	nn := &client.ObjectKey{
		Namespace: "default",
		Name:      randomName,
	}

	routeTable := newRouteTable(nn)
	route := newRoute(nn)
	route.OwnerReferences = ownerutil.EnsureOwnerRef(route.OwnerReferences, metav1.OwnerReference{
		APIVersion: routeTable.APIVersion,
		Kind:       routeTable.Kind,
		Name:       routeTable.Name,
		UID:        routeTable.UID,
	})

	res := new(zips.Resource)
	err := setOwnerInfluencedFields(res, route, ownerReferenceStates{
		{
			Obj:   routeTable,
			State: string(zips.SucceededProvisioningState),
		},
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(res.Name).To(gomega.Equal(fmt.Sprintf("%s/%s", routeTable.Name, route.Name)))
	g.Expect(res.ResourceGroup).To(gomega.Equal(routeTable.GetResourceGroupObjectRef().Name))
}

func Test_resourceTypeToParentTypesInOrder(t *testing.T) {
	cases := []struct {
		Name         string
		ResourceType string
		Parents      []string
	}{
		{
			Name:         "OnlyOneParent",
			ResourceType: "Microsoft.Networks/loadBalancers/routeTables",
			Parents: []string{
				"Microsoft.Networks/loadBalancers",
			},
		},
		{
			Name:         "NoParents",
			ResourceType: "Microsoft.Networks/loadBalancers",
			Parents:      []string{},
		},
		{
			Name:         "TwoParents",
			ResourceType: "Microsoft.Networks/loadBalancers/routeTables/bazzFoos",
			Parents: []string{
				"Microsoft.Networks/loadBalancers",
				"Microsoft.Networks/loadBalancers/routeTables",
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			p := resourceTypeToParentTypesInOrder(c.ResourceType)
			g := gomega.NewGomegaWithT(t)
			g.Expect(p).To(gomega.Equal(c.Parents))
		})
	}
}

func TestIsOwnerNotFound(t *testing.T) {
	err := fmt.Errorf("oops with: %w", &OwnerNotFoundError{
		Owner: "foo",
	})
	g := gomega.NewGomegaWithT(t)
	g.Expect(IsOwnerNotFound(err)).To(gomega.BeTrue())
}

func newVM(nn *client.ObjectKey) *microsoftcomputev1.VirtualMachine {
	return &microsoftcomputev1.VirtualMachine{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualMachine",
			APIVersion: microsoftcomputev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nn.Name + "_vm",
			Namespace: nn.Namespace,
		},
		Spec: microsoftcomputev1.VirtualMachineSpec{
			APIVersion: "2019-12-01",
			Location:   "westus2",
			Properties: &microsoftcomputev1.VirtualMachineProperties{
				NetworkProfile: new(microsoftcomputev1.NetworkProfile),
			},
		},
	}
}

func newNetworkInterface(nn *client.ObjectKey) *microsoftnetworkv1.NetworkInterface {
	return &microsoftnetworkv1.NetworkInterface{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NetworkInterface",
			APIVersion: microsoftnetworkv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nn.Name + "_iface",
			Namespace: nn.Namespace,
		},
		Spec: microsoftnetworkv1.NetworkInterfaceSpec{
			APIVersion: "2019-11-01",
			Properties: &microsoftnetworkv1.NetworkInterfaceSpecProperties{},
		},
	}
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
