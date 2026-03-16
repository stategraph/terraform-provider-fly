package apiclient

import (
	"context"
	"fmt"

	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

func (c *Client) SetSecret(ctx context.Context, appName, key, value string) (*apimodels.Secret, error) {
	req := apimodels.SetSecretRequest{Value: value, Type: "opaque"}
	var secret apimodels.Secret
	err := c.doJSONWithRetry(ctx, "POST", c.restURL(fmt.Sprintf("/apps/%s/secrets/%s", appName, key)), req, &secret)
	if err != nil {
		return nil, fmt.Errorf("setting secret %s for app %s: %w", key, appName, err)
	}
	return &secret, nil
}

func (c *Client) ListSecrets(ctx context.Context, appName string) ([]apimodels.Secret, error) {
	var resp struct {
		Secrets []apimodels.Secret `json:"secrets"`
	}
	err := c.doJSONWithRetry(ctx, "GET", c.restURL(fmt.Sprintf("/apps/%s/secrets", appName)), nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("listing secrets for app %s: %w", appName, err)
	}
	return resp.Secrets, nil
}

func (c *Client) DeleteSecret(ctx context.Context, appName, key string) error {
	err := c.doJSONWithRetry(ctx, "DELETE", c.restURL(fmt.Sprintf("/apps/%s/secrets/%s", appName, key)), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting secret %s for app %s: %w", key, appName, err)
	}
	return nil
}
