package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stategraph/terraform-provider-fly/internal/models"
	"github.com/stategraph/terraform-provider-fly/pkg/apiclient"
)

var (
	_ datasource.DataSource              = &certificateDataSource{}
	_ datasource.DataSourceWithConfigure = &certificateDataSource{}
)

type certificateDataSource struct {
	client *apiclient.Client
}

func NewCertificateDataSource() datasource.DataSource {
	return &certificateDataSource{}
}

func (d *certificateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (d *certificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing TLS certificate for a Fly.io application.",
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				Description: "The name of the application the certificate belongs to.",
				Required:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "The hostname of the certificate.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier of the certificate.",
				Computed:    true,
			},
			"check_status": schema.StringAttribute{
				Description: "The validation check status of the certificate.",
				Computed:    true,
			},
			"dns_validation_hostname": schema.StringAttribute{
				Description: "The hostname to use for DNS validation.",
				Computed:    true,
			},
			"dns_validation_target": schema.StringAttribute{
				Description: "The target to use for DNS validation.",
				Computed:    true,
			},
			"source": schema.StringAttribute{
				Description: "The source of the certificate.",
				Computed:    true,
			},
			"issued_at": schema.StringAttribute{
				Description: "The timestamp when the certificate was issued.",
				Computed:    true,
			},
			"certificate_authority": schema.StringAttribute{
				Description: "The certificate authority that issued the certificate.",
				Computed:    true,
			},
		},
	}
}

func (d *certificateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData),
		)
		return
	}
	d.client = pd.APIClient
}

func (d *certificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config models.CertificateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cert, err := d.client.GetCertificate(ctx, config.App.ValueString(), config.Hostname.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading certificate", err.Error())
		return
	}

	config.ID = types.StringValue(cert.ID)
	config.Hostname = types.StringValue(cert.Hostname)
	config.CheckStatus = types.StringValue(cert.CheckStatus)
	config.DNSValidationHostname = types.StringValue(cert.DNSValidationHostname)
	config.DNSValidationTarget = types.StringValue(cert.DNSValidationTarget)
	config.Source = types.StringValue(cert.Source)
	config.IssuedAt = types.StringValue(cert.IssuedAt)
	config.CertificateAuthority = types.StringValue(cert.CertificateAuthority)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
