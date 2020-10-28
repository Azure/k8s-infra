/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package armclient

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/devigned/tab"
)

type Client struct {
	autorest.Client
	Host string
}

const UserAgent = "k8sinfra-generated"

func NewClient(authorizer autorest.Authorizer) *Client {

	autorestClient := autorest.NewClientWithUserAgent(UserAgent)
	// Disable retries by default
	autorestClient.RetryAttempts = 0
	autorestClient.Authorizer = authorizer

	c := &Client{
		Client: autorestClient,
		Host:   azure.PublicCloud.ResourceManagerEndpoint, // TODO: We need to support other endpoints
	}

	return c
}

func (c *Client) WithExponentialRetries(attempts int, backoff time.Duration, maxBackoff time.Duration) *Client {
	// Copy the client
	result := *c
	result.SendDecorators = nil
	// Deep copy the send decorators
	for _, item := range c.SendDecorators {
		result.SendDecorators = append(result.SendDecorators, item)
	}

	// There's no place to set a backoff cap on the actual client?
	result.RetryAttempts = attempts
	result.RetryDuration = backoff

	result.SendDecorators = append(
		result.SendDecorators,
		autorest.DoRetryForStatusCodesWithCap(
			result.RetryAttempts,
			result.RetryDuration,
			maxBackoff,
			autorest.StatusCodesForRetry...))

	return &result
}

func (c *Client) PutDeployment(ctx context.Context, deployment *Deployment) (*Deployment, error) {
	entityPath, err := deployment.GetEntityPath()
	if err != nil {
		return nil, err
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json"),
		autorest.WithJSON(deployment))

	req, err := c.newRequest(ctx, http.MethodPut, entityPath)
	if err != nil {
		tab.For(ctx).Error(err)
		return nil, err
	}

	req, err = preparer.Prepare(req)
	if err != nil {
		tab.For(ctx).Error(err)
		return nil, err
	}

	resp, err := c.Send(req)
	if err != nil {
		tab.For(ctx).Error(err)
		return nil, err
	}

	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(deployment),
		autorest.ByClosing())
	if err != nil {
		tab.For(ctx).Error(err)
		return nil, err
	}

	return deployment, nil
}

func (c *Client) GetResource(ctx context.Context, resourceID string, resource interface{}) error {

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json"))

	req, err := c.newRequest(ctx, http.MethodGet, resourceID)
	if err != nil {
		return err
	}

	req, err = preparer.Prepare(req)
	if err != nil {
		tab.For(ctx).Error(err)
		return err
	}

	resp, err := c.Send(req)
	if err != nil {
		tab.For(ctx).Error(err)
		return err
	}

	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(resource),
		autorest.ByClosing())
	if err != nil {
		tab.For(ctx).Error(err)
		return err
	}

	// TODO: Dropped NotFound stuff here

	return nil
}

// DeleteResource will make an HTTP DELETE call to the resourceId and attempt to fill the resource with the response.
// If the body of the response is empty, the resource will be nil.
func (c *Client) DeleteResource(ctx context.Context, resourceID string, resource interface{}) error {

	// TODO: This content is basically the same as GET above
	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json"))

	req, err := c.newRequest(ctx, http.MethodDelete, resourceID)
	if err != nil {
		return err
	}

	req, err = preparer.Prepare(req)
	if err != nil {
		tab.For(ctx).Error(err)
		return err
	}

	resp, err := c.Send(req)
	if err != nil {
		tab.For(ctx).Error(err)
		return err
	}

	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusAccepted),
		autorest.ByUnmarshallingJSON(resource),
		autorest.ByClosing())

	if err != nil {
		if IsNotFound(err) {
			// you asked it to be gone, well, it is.
			return nil
		}

		tab.For(ctx).Error(err)
		return err
	}

	return nil
}

func (c *Client) newRequest(ctx context.Context, method string, entityPath string) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, c.Host+strings.TrimPrefix(entityPath, "/"), nil)
}

func IsNotFound(err error) bool {
	var typedError *azure.RequestError
	if errors.As(err, &typedError) {
		if typedError.Response != nil && typedError.Response.StatusCode == 404 {
			return true
		}
	}

	return false
}
