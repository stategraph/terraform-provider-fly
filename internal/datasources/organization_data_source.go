package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stategraph/terraform-provider-fly/internal/models"
	"github.com/stategraph/terraform-provider-fly/pkg/flyctl"
)

var (
	_ datasource.DataSource              = &organizationDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationDataSource{}
)

type organizationDataSource struct {
	flyctl *flyctl.Executor
}

func NewOrganizationDataSource() datasource.DataSource {
	return &organizationDataSource{}
}

func (d *organizationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (d *organizationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a Fly.io organization by slug.",
		Attributes: map[string]schema.Attribute{
			"slug": schema.StringAttribute{
				Description: "The organization slug to look up.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The organization ID.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The organization name.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The organization type.",
				Computed:    true,
			},
			"paid_plan": schema.BoolAttribute{
				Description: "Whether the organization is on a paid plan.",
				Computed:    true,
			},
		},
	}
}

func (d *organizationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_organization data source requires flyctl to be installed.")
		return
	}
	d.flyctl = pd.Flyctl
}

func (d *organizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config models.OrganizationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var org flyctlOrganization
	err := d.flyctl.RunJSON(ctx, &org, "orgs", "show", config.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading organization", err.Error())
		return
	}

	config.ID = types.StringValue(org.ID)
	config.Name = types.StringValue(org.Name)
	config.Slug = types.StringValue(org.Slug)
	config.Type = types.StringValue(org.Type)
	config.PaidPlan = types.BoolValue(org.PaidPlan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// flyctlOrganization represents the JSON output from flyctl orgs show.
type flyctlOrganization struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Type     string `json:"type"`
	PaidPlan bool   `json:"paid_plan"`
}
