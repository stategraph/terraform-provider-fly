package apiclient

import (
	"context"
	"fmt"

	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

func (c *Client) CreateVolume(ctx context.Context, appName string, req apimodels.CreateVolumeRequest) (*apimodels.Volume, error) {
	var volume apimodels.Volume
	err := c.doJSONWithRetry(ctx, "POST", c.restURL(fmt.Sprintf("/apps/%s/volumes", appName)), req, &volume)
	if err != nil {
		return nil, fmt.Errorf("creating volume in app %s: %w", appName, err)
	}
	return &volume, nil
}

func (c *Client) GetVolume(ctx context.Context, appName, volumeID string) (*apimodels.Volume, error) {
	var volume apimodels.Volume
	err := c.doJSONWithRetry(ctx, "GET", c.restURL(fmt.Sprintf("/apps/%s/volumes/%s", appName, volumeID)), nil, &volume)
	if err != nil {
		return nil, fmt.Errorf("getting volume %s in app %s: %w", volumeID, appName, err)
	}
	return &volume, nil
}

func (c *Client) DeleteVolume(ctx context.Context, appName, volumeID string) error {
	err := c.doJSONWithRetry(ctx, "DELETE", c.restURL(fmt.Sprintf("/apps/%s/volumes/%s", appName, volumeID)), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting volume %s in app %s: %w", volumeID, appName, err)
	}
	return nil
}

func (c *Client) ExtendVolume(ctx context.Context, appName, volumeID string, sizeGB int) (*apimodels.ExtendVolumeResponse, error) {
	req := apimodels.ExtendVolumeRequest{SizeGB: sizeGB}
	var resp apimodels.ExtendVolumeResponse
	err := c.doJSONWithRetry(ctx, "PUT", c.restURL(fmt.Sprintf("/apps/%s/volumes/%s/extend", appName, volumeID)), req, &resp)
	if err != nil {
		return nil, fmt.Errorf("extending volume %s in app %s: %w", volumeID, appName, err)
	}
	return &resp, nil
}

func (c *Client) ListVolumes(ctx context.Context, appName string) ([]apimodels.Volume, error) {
	var volumes []apimodels.Volume
	err := c.doJSONWithRetry(ctx, "GET", c.restURL(fmt.Sprintf("/apps/%s/volumes", appName)), nil, &volumes)
	if err != nil {
		return nil, fmt.Errorf("listing volumes in app %s: %w", appName, err)
	}
	return volumes, nil
}

func (c *Client) UpdateVolume(ctx context.Context, appName, volumeID string, req apimodels.UpdateVolumeRequest) (*apimodels.Volume, error) {
	var volume apimodels.Volume
	err := c.doJSONWithRetry(ctx, "PUT", c.restURL(fmt.Sprintf("/apps/%s/volumes/%s", appName, volumeID)), req, &volume)
	if err != nil {
		return nil, fmt.Errorf("updating volume %s in app %s: %w", volumeID, appName, err)
	}
	return &volume, nil
}

func (c *Client) ListVolumeSnapshots(ctx context.Context, appName, volumeID string) ([]apimodels.VolumeSnapshot, error) {
	var snapshots []apimodels.VolumeSnapshot
	err := c.doJSONWithRetry(ctx, "GET", c.restURL(fmt.Sprintf("/apps/%s/volumes/%s/snapshots", appName, volumeID)), nil, &snapshots)
	if err != nil {
		return nil, fmt.Errorf("listing snapshots for volume %s in app %s: %w", volumeID, appName, err)
	}
	return snapshots, nil
}

func (c *Client) CreateVolumeSnapshot(ctx context.Context, appName, volumeID string) (*apimodels.VolumeSnapshot, error) {
	var snapshot apimodels.VolumeSnapshot
	err := c.doJSONWithRetry(ctx, "POST", c.restURL(fmt.Sprintf("/apps/%s/volumes/%s/snapshots", appName, volumeID)), nil, &snapshot)
	if err != nil {
		return nil, fmt.Errorf("creating snapshot for volume %s in app %s: %w", volumeID, appName, err)
	}
	return &snapshot, nil
}
