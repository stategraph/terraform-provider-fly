package apiclient

import (
	"context"
	"fmt"

	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

func (c *Client) CreateApp(ctx context.Context, req apimodels.CreateAppRequest) (*apimodels.App, error) {
	var app apimodels.App
	err := c.doJSONWithRetry(ctx, "POST", c.restURL("/apps"), req, &app)
	if err != nil {
		return nil, fmt.Errorf("creating app: %w", err)
	}
	normalizeApp(&app)
	return &app, nil
}

func (c *Client) GetApp(ctx context.Context, appName string) (*apimodels.App, error) {
	var app apimodels.App
	err := c.doJSONWithRetry(ctx, "GET", c.restURL(fmt.Sprintf("/apps/%s", appName)), nil, &app)
	if err != nil {
		return nil, fmt.Errorf("getting app %s: %w", appName, err)
	}
	normalizeApp(&app)
	return &app, nil
}

func (c *Client) DeleteApp(ctx context.Context, appName string) error {
	err := c.doJSONWithRetry(ctx, "DELETE", c.restURL(fmt.Sprintf("/apps/%s", appName)), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting app %s: %w", appName, err)
	}
	return nil
}

func (c *Client) ListApps(ctx context.Context, orgSlug string) ([]apimodels.App, error) {
	url := c.restURL("/apps")
	if orgSlug != "" {
		url += "?org_slug=" + orgSlug
	}
	var resp struct {
		Apps      []apimodels.App `json:"apps"`
		TotalApps int             `json:"total_apps"`
	}
	err := c.doJSONWithRetry(ctx, "GET", url, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("listing apps: %w", err)
	}
	for i := range resp.Apps {
		normalizeApp(&resp.Apps[i])
	}
	return resp.Apps, nil
}

// normalizeApp populates the convenience OrgSlug field from the nested Organization.
func normalizeApp(app *apimodels.App) {
	if app.OrgSlug == "" && app.Organization.Slug != "" {
		app.OrgSlug = app.Organization.Slug
	}
}
