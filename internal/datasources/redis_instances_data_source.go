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
	_ datasource.DataSource              = &redisInstancesDataSource{}
	_ datasource.DataSourceWithConfigure = &redisInstancesDataSource{}
)

type redisInstancesDataSource struct {
	flyctl *flyctl.Executor
}

func NewRedisInstancesDataSource() datasource.DataSource {
	return &redisInstancesDataSource{}
}

func (d *redisInstancesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redis_instances"
}

func (d *redisInstancesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists Fly.io Redis instances.",
		Attributes: map[string]schema.Attribute{
			"instances": schema.ListNestedAttribute{
				Description: "The list of Redis instances.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":     schema.StringAttribute{Computed: true, Description: "Instance ID."},
						"name":   schema.StringAttribute{Computed: true, Description: "Instance name."},
						"status": schema.StringAttribute{Computed: true, Description: "Instance status."},
						"plan":   schema.StringAttribute{Computed: true, Description: "Instance plan."},
						"region": schema.StringAttribute{Computed: true, Description: "Instance region."},
					},
				},
			},
		},
	}
}

func (d *redisInstancesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_redis_instances data source requires flyctl to be installed.")
		return
	}
	d.flyctl = pd.Flyctl
}

// redisInstancesDataSourceModel is the Terraform state model.
type redisInstancesDataSourceModel struct {
	Instances []redisInstanceModel `tfsdk:"instances"`
}

type redisInstanceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Status types.String `tfsdk:"status"`
	Plan   types.String `tfsdk:"plan"`
	Region types.String `tfsdk:"region"`
}

func (d *redisInstancesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config redisInstancesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var instances []flyctlRedisInstanceDS
	err := d.flyctl.RunJSON(ctx, &instances, "redis", "list")
	if err != nil {
		resp.Diagnostics.AddError("Error listing Redis instances", err.Error())
		return
	}

	config.Instances = make([]redisInstanceModel, len(instances))
	for i, inst := range instances {
		config.Instances[i] = redisInstanceModel{
			ID:     types.StringValue(inst.ID),
			Name:   types.StringValue(inst.Name),
			Status: types.StringValue(inst.Status),
			Plan:   types.StringValue(inst.Plan),
			Region: types.StringValue(inst.Region),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// flyctlRedisInstanceDS represents the JSON output from flyctl redis list.
type flyctlRedisInstanceDS struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Plan   string `json:"plan"`
	Region string `json:"region"`
}
