package apiclient

import (
	"context"
	"fmt"

	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

func (c *Client) CreateNetworkPolicy(ctx context.Context, appName string, req apimodels.CreateNetworkPolicyRequest) (*apimodels.NetworkPolicy, error) {
	var policy apimodels.NetworkPolicy
	err := c.doJSONWithRetry(ctx, "POST", c.restURL(fmt.Sprintf("/apps/%s/network_policies", appName)), req, &policy)
	if err != nil {
		return nil, fmt.Errorf("creating network policy for app %s: %w", appName, err)
	}
	return &policy, nil
}

func (c *Client) ListNetworkPolicies(ctx context.Context, appName string) ([]apimodels.NetworkPolicy, error) {
	var policies []apimodels.NetworkPolicy
	err := c.doJSONWithRetry(ctx, "GET", c.restURL(fmt.Sprintf("/apps/%s/network_policies", appName)), nil, &policies)
	if err != nil {
		return nil, fmt.Errorf("listing network policies for app %s: %w", appName, err)
	}
	return policies, nil
}

func (c *Client) GetNetworkPolicy(ctx context.Context, appName, policyID string) (*apimodels.NetworkPolicy, error) {
	var policy apimodels.NetworkPolicy
	err := c.doJSONWithRetry(ctx, "GET", c.restURL(fmt.Sprintf("/apps/%s/network_policies/%s", appName, policyID)), nil, &policy)
	if err != nil {
		return nil, fmt.Errorf("getting network policy %s for app %s: %w", policyID, appName, err)
	}
	return &policy, nil
}

func (c *Client) DeleteNetworkPolicy(ctx context.Context, appName, policyID string) error {
	err := c.doJSONWithRetry(ctx, "DELETE", c.restURL(fmt.Sprintf("/apps/%s/network_policies/%s", appName, policyID)), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting network policy %s for app %s: %w", policyID, appName, err)
	}
	return nil
}
