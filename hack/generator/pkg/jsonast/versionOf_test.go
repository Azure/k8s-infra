package jsonast

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ExtractObjectAndVersion(t *testing.T) {
	url, err := url.Parse("https://schema.management.azure.com/schemas/2015-01-01/Microsoft.Resources.json#/resourceDefinitions/deployments")
	assert.Nil(t, err)
	version, err := versionOf(url)
	assert.Nil(t, err)
	assert.Equal(t, "2015-01-01", version)
}
