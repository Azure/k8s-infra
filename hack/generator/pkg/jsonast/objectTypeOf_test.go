package jsonast

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ObjectTypeOf(t *testing.T) {
	url, err := url.Parse("https://schema.management.azure.com/schemas/2015-01-01/Microsoft.Resources.json#/resourceDefinitions/deployments")
	assert.Nil(t, err)
	name := objectTypeOf(url)
	assert.Equal(t, "deployments", name)
}
