package apiclient

import (
	"context"
	"fmt"

	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

func (c *Client) CreateMachine(ctx context.Context, appName string, req apimodels.CreateMachineRequest) (*apimodels.Machine, error) {
	var machine apimodels.Machine
	err := c.doJSONWithRetry(ctx, "POST", c.restURL(fmt.Sprintf("/apps/%s/machines", appName)), req, &machine)
	if err != nil {
		return nil, fmt.Errorf("creating machine in app %s: %w", appName, err)
	}
	return &machine, nil
}

func (c *Client) GetMachine(ctx context.Context, appName, machineID string) (*apimodels.Machine, error) {
	var machine apimodels.Machine
	err := c.doJSONWithRetry(ctx, "GET", c.restURL(fmt.Sprintf("/apps/%s/machines/%s", appName, machineID)), nil, &machine)
	if err != nil {
		return nil, fmt.Errorf("getting machine %s in app %s: %w", machineID, appName, err)
	}
	return &machine, nil
}

func (c *Client) UpdateMachine(ctx context.Context, appName, machineID string, req apimodels.UpdateMachineRequest) (*apimodels.Machine, error) {
	var machine apimodels.Machine
	err := c.doJSON(ctx, "POST", c.restURL(fmt.Sprintf("/apps/%s/machines/%s", appName, machineID)), req, &machine)
	if err != nil {
		return nil, fmt.Errorf("updating machine %s in app %s: %w", machineID, appName, err)
	}
	return &machine, nil
}

func (c *Client) DeleteMachine(ctx context.Context, appName, machineID string) error {
	err := c.doJSONWithRetry(ctx, "DELETE", c.restURL(fmt.Sprintf("/apps/%s/machines/%s", appName, machineID)), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting machine %s in app %s: %w", machineID, appName, err)
	}
	return nil
}

func (c *Client) StartMachine(ctx context.Context, appName, machineID string) error {
	err := c.doJSONWithRetry(ctx, "POST", c.restURL(fmt.Sprintf("/apps/%s/machines/%s/start", appName, machineID)), nil, nil)
	if err != nil {
		return fmt.Errorf("starting machine %s in app %s: %w", machineID, appName, err)
	}
	return nil
}

func (c *Client) StopMachine(ctx context.Context, appName, machineID string) error {
	err := c.doJSONWithRetry(ctx, "POST", c.restURL(fmt.Sprintf("/apps/%s/machines/%s/stop", appName, machineID)), nil, nil)
	if err != nil {
		return fmt.Errorf("stopping machine %s in app %s: %w", machineID, appName, err)
	}
	return nil
}

func (c *Client) WaitForMachine(ctx context.Context, appName, machineID, state string, timeoutSeconds int) error {
	if c.DryRun {
		return nil
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = 60
	}
	url := c.restURL(fmt.Sprintf("/apps/%s/machines/%s/wait?state=%s&timeout=%d", appName, machineID, state, timeoutSeconds))
	err := c.doJSON(ctx, "GET", url, nil, nil)
	if err != nil {
		return fmt.Errorf("waiting for machine %s state %s: %w", machineID, state, err)
	}
	return nil
}

func (c *Client) AcquireLease(ctx context.Context, appName, machineID string, ttlSeconds int) (*apimodels.Lease, error) {
	req := apimodels.LeaseRequest{TTL: ttlSeconds}
	var lease apimodels.Lease
	err := c.doJSONWithRetry(ctx, "POST", c.restURL(fmt.Sprintf("/apps/%s/machines/%s/lease", appName, machineID)), req, &lease)
	if err != nil {
		return nil, fmt.Errorf("acquiring lease on machine %s: %w", machineID, err)
	}
	return &lease, nil
}

func (c *Client) GetLease(ctx context.Context, appName, machineID string) (*apimodels.Lease, error) {
	var lease apimodels.Lease
	err := c.doJSONWithRetry(ctx, "GET", c.restURL(fmt.Sprintf("/apps/%s/machines/%s/lease", appName, machineID)), nil, &lease)
	if err != nil {
		return nil, fmt.Errorf("getting lease on machine %s: %w", machineID, err)
	}
	return &lease, nil
}

func (c *Client) ReleaseLease(ctx context.Context, appName, machineID, nonce string) error {
	err := c.doJSON(ctx, "DELETE", c.restURL(fmt.Sprintf("/apps/%s/machines/%s/lease", appName, machineID)), nil, nil)
	if err != nil {
		return fmt.Errorf("releasing lease on machine %s: %w", machineID, err)
	}
	return nil
}

func (c *Client) ListMachines(ctx context.Context, appName string) ([]apimodels.Machine, error) {
	var machines []apimodels.Machine
	err := c.doJSONWithRetry(ctx, "GET", c.restURL(fmt.Sprintf("/apps/%s/machines", appName)), nil, &machines)
	if err != nil {
		return nil, fmt.Errorf("listing machines in app %s: %w", appName, err)
	}
	return machines, nil
}

func (c *Client) SuspendMachine(ctx context.Context, appName, machineID string) error {
	err := c.doJSONWithRetry(ctx, "POST", c.restURL(fmt.Sprintf("/apps/%s/machines/%s/suspend", appName, machineID)), nil, nil)
	if err != nil {
		return fmt.Errorf("suspending machine %s in app %s: %w", machineID, appName, err)
	}
	return nil
}

func (c *Client) CordonMachine(ctx context.Context, appName, machineID string) error {
	err := c.doJSONWithRetry(ctx, "POST", c.restURL(fmt.Sprintf("/apps/%s/machines/%s/cordon", appName, machineID)), nil, nil)
	if err != nil {
		return fmt.Errorf("cordoning machine %s in app %s: %w", machineID, appName, err)
	}
	return nil
}

func (c *Client) UncordonMachine(ctx context.Context, appName, machineID string) error {
	err := c.doJSONWithRetry(ctx, "POST", c.restURL(fmt.Sprintf("/apps/%s/machines/%s/uncordon", appName, machineID)), nil, nil)
	if err != nil {
		return fmt.Errorf("uncordoning machine %s in app %s: %w", machineID, appName, err)
	}
	return nil
}

func (c *Client) GetMetadata(ctx context.Context, appName, machineID string) (map[string]string, error) {
	var metadata map[string]string
	err := c.doJSONWithRetry(ctx, "GET", c.restURL(fmt.Sprintf("/apps/%s/machines/%s/metadata", appName, machineID)), nil, &metadata)
	if err != nil {
		return nil, fmt.Errorf("getting metadata for machine %s: %w", machineID, err)
	}
	return metadata, nil
}

func (c *Client) SetMetadataKey(ctx context.Context, appName, machineID, key, value string) error {
	err := c.doJSONWithRetry(ctx, "POST", c.restURL(fmt.Sprintf("/apps/%s/machines/%s/metadata/%s", appName, machineID, key)), map[string]string{"value": value}, nil)
	if err != nil {
		return fmt.Errorf("setting metadata key %s on machine %s: %w", key, machineID, err)
	}
	return nil
}

func (c *Client) DeleteMetadataKey(ctx context.Context, appName, machineID, key string) error {
	err := c.doJSONWithRetry(ctx, "DELETE", c.restURL(fmt.Sprintf("/apps/%s/machines/%s/metadata/%s", appName, machineID, key)), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting metadata key %s on machine %s: %w", key, machineID, err)
	}
	return nil
}
