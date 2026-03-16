package apiclient

import (
	"context"
	"fmt"

	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

func (c *Client) RequestOIDCToken(ctx context.Context, aud string) (*apimodels.OIDCTokenResponse, error) {
	req := apimodels.OIDCTokenRequest{Aud: aud}
	var resp apimodels.OIDCTokenResponse
	err := c.doJSONWithRetry(ctx, "POST", c.restURL("/tokens/oidc"), req, &resp)
	if err != nil {
		return nil, fmt.Errorf("requesting OIDC token: %w", err)
	}
	return &resp, nil
}
