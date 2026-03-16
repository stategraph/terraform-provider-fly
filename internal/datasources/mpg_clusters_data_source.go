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
	_ datasource.DataSource              = &mpgClustersDataSource{}
	_ datasource.DataSourceWithConfigure = &mpgClustersDataSource{}
)

type mpgClustersDataSource struct {
	flyctl *flyctl.Executor
}

func NewMPGClustersDataSource() datasource.DataSource {
	return &mpgClustersDataSource{}
}

func (d *mpgClustersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mpg_clusters"
}

func (d *mpgClustersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists Fly.io Managed Postgres (MPG) clusters.",
		Attributes: map[string]schema.Attribute{
			"clusters": schema.ListNestedAttribute{
				Description: "The list of MPG clusters.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":     schema.StringAttribute{Computed: true, Description: "Cluster ID."},
						"name":   schema.StringAttribute{Computed: true, Description: "Cluster name."},
						"status": schema.StringAttribute{Computed: true, Description: "Cluster status."},
						"region": schema.StringAttribute{Computed: true, Description: "Cluster region."},
					},
				},
			},
		},
	}
}

func (d *mpgClustersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_mpg_clusters data source requires flyctl to be installed.")
		return
	}
	d.flyctl = pd.Flyctl
}

// mpgClustersDataSourceModel is the Terraform state model.
type mpgClustersDataSourceModel struct {
	Clusters []mpgClusterModel `tfsdk:"clusters"`
}

type mpgClusterModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Status types.String `tfsdk:"status"`
	Region types.String `tfsdk:"region"`
}

func (d *mpgClustersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config mpgClustersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var clusters []flyctlMPGClusterDS
	err := d.flyctl.RunJSON(ctx, &clusters, "mpg", "list")
	if err != nil {
		resp.Diagnostics.AddError("Error listing MPG clusters", err.Error())
		return
	}

	config.Clusters = make([]mpgClusterModel, len(clusters))
	for i, c := range clusters {
		config.Clusters[i] = mpgClusterModel{
			ID:     types.StringValue(c.ID),
			Name:   types.StringValue(c.Name),
			Status: types.StringValue(c.Status),
			Region: types.StringValue(c.Region),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// flyctlMPGClusterDS represents the JSON output from flyctl mpg list.
type flyctlMPGClusterDS struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Region string `json:"region"`
}
