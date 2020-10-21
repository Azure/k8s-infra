/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package controllers_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	storage "github.com/Azure/k8s-infra/hack/generated/apis/microsoft.storage/v20190401"
	"github.com/Azure/k8s-infra/hack/generated/pkg/armclient"
	"github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
)

func Integration_ResourceGroup_CRUD(testContext ControllerTestContext, t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	// Create a resource group
	rg := testContext.NewTestResourceGroup()
	err := testContext.KubeClient.Create(ctx, rg)
	g.Expect(err).ToNot(HaveOccurred())

	// It should be created in Kubernetes
	g.Eventually(testContext.Ensure.ProvisioningComplete(ctx, rg)).Should(BeTrue())
	g.Expect(testContext.Ensure.ProvisioningSuccessful(ctx, rg)).Should(BeTrue())

	g.Expect(rg.Status.Location).To(Equal(testContext.AzureRegion))
	g.Expect(rg.Status.Properties.ProvisioningState).To(Equal(string(armclient.SucceededProvisioningState)))
	g.Expect(rg.Status.ID).ToNot(BeNil())
	armId := rg.Status.ID

	// Delete the resource group
	err = testContext.KubeClient.Delete(ctx, rg)
	g.Expect(err).ToNot(HaveOccurred())
	g.Eventually(testContext.Ensure.Deleted(ctx, rg)).Should(BeTrue())

	// Ensure that the resource group was really deleted in Azure
	// TODO: Do we want to just use an SDK here? This process is quite icky as is...
	exists, err := testContext.AzureClient.HeadResource(
		ctx,
		armId,
		"2020-06-01")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(exists).To(BeFalse())
}

func Integration_StorageAccount_CRUD(testContext ControllerTestContext, t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	// Custom namer because storage accounts have strict names
	namer := testContext.Namer.WithSeparator("")

	// Create a storage account
	accessTier := storage.StorageAccountPropertiesCreateParametersAccessTierHot
	acct := &storage.StorageAccount{
		ObjectMeta: testContext.MakeObjectMetaWithName(namer.GenerateName("stor")),
		Spec: storage.StorageAccounts_Spec{
			Location:   testContext.AzureRegion,
			ApiVersion: "2019-04-01", // TODO: This should be removed from the storage type eventually
			Owner:      genruntime.KnownResourceReference{Name: testContext.SharedResourceGroup.Name},
			Kind:       storage.StorageAccountsSpecKindBlobStorage,
			Sku: storage.Sku{
				Name: storage.SkuNameStandardLRS,
			},
			// TODO: They mark this property as optional but actually it is required
			Properties: &storage.StorageAccountPropertiesCreateParameters{
				AccessTier: &accessTier,
			},
		},
	}
	err := testContext.KubeClient.Create(ctx, acct)
	g.Expect(err).ToNot(HaveOccurred())

	// It should be created in Kubernetes
	g.Eventually(testContext.Ensure.ProvisioningComplete(ctx, acct)).Should(BeTrue())
	g.Expect(testContext.Ensure.ProvisioningSuccessful(ctx, acct)).Should(BeTrue())

	g.Expect(acct.Status.Location).To(Equal(testContext.AzureRegion))
	expectedKind := storage.StorageAccountStatusKindBlobStorage
	g.Expect(acct.Status.Kind).To(Equal(&expectedKind))
	g.Expect(acct.Status.Id).ToNot(BeNil())
	armId := *acct.Status.Id

	// Delete the storage account
	err = testContext.KubeClient.Delete(ctx, acct)
	g.Expect(err).ToNot(HaveOccurred())
	g.Eventually(testContext.Ensure.Deleted(ctx, acct)).Should(BeTrue())

	// Ensure that the resource group was really deleted in Azure
	// TODO: Do we want to just use an SDK here? This process is quite icky as is...
	exists, err := testContext.AzureClient.HeadResource(
		ctx,
		armId,
		"2019-04-01")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(exists).To(BeFalse())
}
