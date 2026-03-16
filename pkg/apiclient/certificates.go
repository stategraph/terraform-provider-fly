package apiclient

import (
	"context"
	"fmt"

	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

func (c *Client) AddCertificate(ctx context.Context, appName string, req apimodels.AddCertificateRequest) (*apimodels.Certificate, error) {
	var cert apimodels.Certificate
	err := c.doJSONWithRetry(ctx, "POST", c.restURL(fmt.Sprintf("/apps/%s/certificates", appName)), req, &cert)
	if err != nil {
		return nil, fmt.Errorf("adding certificate for app %s: %w", appName, err)
	}
	return &cert, nil
}

func (c *Client) GetCertificate(ctx context.Context, appName, hostname string) (*apimodels.Certificate, error) {
	var cert apimodels.Certificate
	err := c.doJSONWithRetry(ctx, "GET", c.restURL(fmt.Sprintf("/apps/%s/certificates/%s", appName, hostname)), nil, &cert)
	if err != nil {
		return nil, fmt.Errorf("getting certificate %s for app %s: %w", hostname, appName, err)
	}
	return &cert, nil
}

func (c *Client) DeleteCertificate(ctx context.Context, appName, hostname string) error {
	err := c.doJSONWithRetry(ctx, "DELETE", c.restURL(fmt.Sprintf("/apps/%s/certificates/%s", appName, hostname)), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting certificate %s for app %s: %w", hostname, appName, err)
	}
	return nil
}

func (c *Client) ListCertificates(ctx context.Context, appName string) ([]apimodels.Certificate, error) {
	var resp struct {
		Certificates []apimodels.Certificate `json:"certificates"`
	}
	err := c.doJSONWithRetry(ctx, "GET", c.restURL(fmt.Sprintf("/apps/%s/certificates", appName)), nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("listing certificates for app %s: %w", appName, err)
	}
	return resp.Certificates, nil
}

func (c *Client) AddACMECertificate(ctx context.Context, appName string, req apimodels.ACMECertificateRequest) (*apimodels.Certificate, error) {
	var cert apimodels.Certificate
	err := c.doJSONWithRetry(ctx, "POST", c.restURL(fmt.Sprintf("/apps/%s/certificates/acme", appName)), req, &cert)
	if err != nil {
		return nil, fmt.Errorf("requesting ACME certificate for app %s: %w", appName, err)
	}
	return &cert, nil
}

func (c *Client) AddCustomCertificate(ctx context.Context, appName string, req apimodels.CustomCertificateRequest) (*apimodels.Certificate, error) {
	var cert apimodels.Certificate
	err := c.doJSONWithRetry(ctx, "POST", c.restURL(fmt.Sprintf("/apps/%s/certificates/custom", appName)), req, &cert)
	if err != nil {
		return nil, fmt.Errorf("importing custom certificate for app %s: %w", appName, err)
	}
	return &cert, nil
}

func (c *Client) DeleteACMECertificate(ctx context.Context, appName, hostname string) error {
	err := c.doJSONWithRetry(ctx, "DELETE", c.restURL(fmt.Sprintf("/apps/%s/certificates/%s/acme", appName, hostname)), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting ACME certificate %s for app %s: %w", hostname, appName, err)
	}
	return nil
}

func (c *Client) DeleteCustomCertificate(ctx context.Context, appName, hostname string) error {
	err := c.doJSONWithRetry(ctx, "DELETE", c.restURL(fmt.Sprintf("/apps/%s/certificates/%s/custom", appName, hostname)), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting custom certificate %s for app %s: %w", hostname, appName, err)
	}
	return nil
}

func (c *Client) CheckCertificate(ctx context.Context, appName, hostname string) (*apimodels.Certificate, error) {
	var cert apimodels.Certificate
	err := c.doJSONWithRetry(ctx, "POST", c.restURL(fmt.Sprintf("/apps/%s/certificates/%s/check", appName, hostname)), nil, &cert)
	if err != nil {
		return nil, fmt.Errorf("checking certificate %s for app %s: %w", hostname, appName, err)
	}
	return &cert, nil
}
