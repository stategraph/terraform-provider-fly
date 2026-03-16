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
	_ datasource.DataSource              = &appDataSource{}
	_ datasource.DataSourceWithConfigure = &appDataSource{}
)

type appDataSource struct {
	client *apiclient.Client
}

func NewAppDataSource() datasource.DataSource {
	return &appDataSource{}
}

func (d *appDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (d *appDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing Fly.io application.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the application to look up.",
				Required:    true,
			},
			"org_slug": schema.StringAttribute{
				Description: "The slug of the organization the app belongs to.",
				Computed:    true,
			},
			"network": schema.StringAttribute{
				Description: "The network the app is attached to.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The current status of the application.",
				Computed:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "The hostname of the application.",
				Computed:    true,
			},
		},
	}
}

func (d *appDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *appDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config models.AppDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := d.client.GetApp(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading app", err.Error())
		return
	}

	config.Name = types.StringValue(app.Name)
	config.OrgSlug = types.StringValue(app.OrgSlug)
	config.Network = types.StringValue(app.Network)
	config.Status = types.StringValue(app.Status)
	config.Hostname = types.StringValue(app.Hostname)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
