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
	_ datasource.DataSource              = &regionsDataSource{}
	_ datasource.DataSourceWithConfigure = &regionsDataSource{}
)

type regionsDataSource struct {
	flyctl *flyctl.Executor
}

func NewRegionsDataSource() datasource.DataSource {
	return &regionsDataSource{}
}

func (d *regionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_regions"
}

func (d *regionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists available Fly.io regions.",
		Attributes: map[string]schema.Attribute{
			"regions": schema.ListNestedAttribute{
				Description: "The list of regions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"code":    schema.StringAttribute{Computed: true, Description: "Region code."},
						"name":    schema.StringAttribute{Computed: true, Description: "Region name."},
						"gateway": schema.BoolAttribute{Computed: true, Description: "Whether region is a gateway."},
					},
				},
			},
		},
	}
}

func (d *regionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_regions data source requires flyctl to be installed.")
		return
	}
	d.flyctl = pd.Flyctl
}

func (d *regionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config models.RegionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var regions []flyctlRegion
	err := d.flyctl.RunJSON(ctx, &regions, "platform", "regions")
	if err != nil {
		resp.Diagnostics.AddError("Error listing regions", err.Error())
		return
	}

	config.Regions = make([]models.RegionModel, len(regions))
	for i, r := range regions {
		config.Regions[i] = models.RegionModel{
			Code:    types.StringValue(r.Code),
			Name:    types.StringValue(r.Name),
			Gateway: types.BoolValue(r.Gateway),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// flyctlRegion represents the JSON output from flyctl platform regions.
type flyctlRegion struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Gateway bool   `json:"gateway"`
}
