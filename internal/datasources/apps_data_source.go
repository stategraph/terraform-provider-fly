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
	_ datasource.DataSource              = &appsDataSource{}
	_ datasource.DataSourceWithConfigure = &appsDataSource{}
)

type appsDataSource struct {
	client *apiclient.Client
}

func NewAppsDataSource() datasource.DataSource {
	return &appsDataSource{}
}

func (d *appsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apps"
}

func (d *appsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists Fly.io applications, optionally filtered by organization.",
		Attributes: map[string]schema.Attribute{
			"org_slug": schema.StringAttribute{
				Description: "Filter apps by organization slug.",
				Optional:    true,
			},
			"apps": schema.ListNestedAttribute{
				Description: "The list of applications.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":       schema.StringAttribute{Computed: true, Description: "App ID."},
						"name":     schema.StringAttribute{Computed: true, Description: "App name."},
						"org_slug": schema.StringAttribute{Computed: true, Description: "Organization slug."},
						"network":  schema.StringAttribute{Computed: true, Description: "Network name."},
						"status":   schema.StringAttribute{Computed: true, Description: "App status."},
					},
				},
			},
		},
	}
}

func (d *appsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *appsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config models.AppsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apps, err := d.client.ListApps(ctx, config.OrgSlug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing apps", err.Error())
		return
	}

	config.Apps = make([]models.AppItemModel, len(apps))
	for i, app := range apps {
		config.Apps[i] = models.AppItemModel{
			ID:      types.StringValue(app.ID),
			Name:    types.StringValue(app.Name),
			OrgSlug: types.StringValue(app.OrgSlug),
			Network: types.StringValue(app.Network),
			Status:  types.StringValue(app.Status),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
