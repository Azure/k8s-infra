//+build integration

package zips_test

import (
	"context"
	"testing"

	"github.com/onsi/gomega"

	"github.com/Azure/k8s-infra/pkg/zips"
)

func TestAzureTemplateClient_Deploy(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	client, err := zips.NewAzureTemplateClient()
	g.Expect(err).To(gomega.BeNil())

	res, err := client.Deploy(context.TODO(), zips.Resource{
		Name:       "foo",
		Location:   "westus2",
		Type:       "Microsoft.Resources/resourceGroups",
		APIVersion: "2018-05-01",
	})
	defer func() {
		// TODO: have a better plan for cleaning up after tests
		if res.ID != "" {
			_ = client.Delete(context.TODO(), res.ID)
		}
	}()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(res.ID).ToNot(gomega.BeEmpty())
	g.Expect(res.Properties).ToNot(gomega.BeNil())
}
