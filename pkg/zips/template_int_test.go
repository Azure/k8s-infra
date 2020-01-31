//+build integration

/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package zips_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/onsi/gomega"

	"github.com/Azure/k8s-infra/pkg/zips"
)

var (
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyz123456789")
)

func init() {
	rand.Seed(time.Now().Unix())
}

func TestAzureTemplateClient_Deploy(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	client, err := zips.NewAzureTemplateClient()
	g.Expect(err).To(gomega.BeNil())

	random := RandomName("foo", 10)
	res, err := client.Apply(context.TODO(), zips.Resource{
		Name:       random,
		Location:   "westus2",
		Type:       "Microsoft.Resources/resourceGroups",
		APIVersion: "2018-05-01",
	})
	defer func() {
		// TODO: have a better plan for cleaning up after tests
		if res.ID != "" {
			_ = client.Delete(context.TODO(), res)
		}

		if res.DeploymentID != "" {
			dep := zips.Resource{
				ID:   res.DeploymentID,
				Type: "Microsoft.Resources/deployments",
			}
			_ = client.Delete(context.TODO(), dep)
		}
	}()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(res.ID).ToNot(gomega.BeEmpty())
	g.Expect(res.Properties).ToNot(gomega.BeNil())
}

// RandomName generates a random Event Hub name tagged with the suite id
func RandomName(prefix string, length int) string {
	return RandomString(prefix, length)
}

// RandomString generates a random string with prefix
func RandomString(prefix string, length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return prefix + string(b)
}
