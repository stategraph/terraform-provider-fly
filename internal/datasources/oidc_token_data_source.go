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
	_ datasource.DataSource              = &oidcTokenDataSource{}
	_ datasource.DataSourceWithConfigure = &oidcTokenDataSource{}
)

type oidcTokenDataSource struct {
	client *apiclient.Client
}

func NewOIDCTokenDataSource() datasource.DataSource {
	return &oidcTokenDataSource{}
}

func (d *oidcTokenDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oidc_token"
}

func (d *oidcTokenDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Requests an OIDC token from Fly.io. The token is ephemeral and regenerated on every read.",
		Attributes: map[string]schema.Attribute{
			"aud": schema.StringAttribute{
				Description: "The audience claim for the OIDC token.",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "The generated OIDC token (JWT).",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func (d *oidcTokenDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	d.client = pd.APIClient
}

func (d *oidcTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config models.OIDCTokenDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tokenResp, err := d.client.RequestOIDCToken(ctx, config.Aud.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error requesting OIDC token", err.Error())
		return
	}

	config.Token = types.StringValue(tokenResp.Token)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
