package armclient

import (
	"context"
)
type Client interface {
	DoRequest(ctx context.Context, method, url string) (string, error)
	DoRequestWithBody(ctx context.Context, method, url, body string) (string, error)
}

// ResourceResponse Resources list rest type
type ResourceResponse struct {
	Resources []Resource `json:"value"`
}

// Resource is a resource in azure
type Resource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Sku  struct {
		Name string `json:"name"`
		Tier string `json:"tier"`
	} `json:"sku"`
	Kind       string `json:"kind"`
	Location   string `json:"location"`
	Properties struct {
		ProvisioningState string `json:"provisioningState"`
	} `json:"properties"`
}