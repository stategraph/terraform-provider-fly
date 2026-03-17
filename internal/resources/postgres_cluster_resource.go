package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stategraph/terraform-provider-fly/internal/models"
	"github.com/stategraph/terraform-provider-fly/pkg/flyctl"
)

var (
	_ resource.Resource                = &postgresClusterResource{}
	_ resource.ResourceWithConfigure   = &postgresClusterResource{}
	_ resource.ResourceWithImportState = &postgresClusterResource{}
)

type postgresClusterResource struct {
	flyctl *flyctl.Executor
}

func NewPostgresClusterResource() resource.Resource {
	return &postgresClusterResource{}
}

func (r *postgresClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgres_cluster"
}

func (r *postgresClusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Fly.io Postgres cluster. Import using the cluster name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "The unique identifier of the Postgres cluster.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description:   "The name of the Postgres cluster. Changing this forces a new resource.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"org": schema.StringAttribute{
				Description:   "The organization slug. Changing this forces a new resource.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"region": schema.StringAttribute{
				Description:   "The primary region. Changing this forces a new resource.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"cluster_size": schema.Int64Attribute{
				Description: "The number of nodes in the cluster. Changing this forces a new resource.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"volume_size": schema.Int64Attribute{
				Description: "Volume size in GB. Changing this forces a new resource.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"vm_size": schema.StringAttribute{
				Description: "The VM size (e.g., shared-cpu-1x). Changing this forces a new resource.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enable_backups": schema.BoolAttribute{
				Description: "Enable automatic backups. Changing this forces a new resource.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The cluster status.",
				Computed:    true,
			},
		},
	}
}

func (r *postgresClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_postgres_cluster resource requires flyctl to be installed.")
		return
	}
	r.flyctl = pd.Flyctl
}

func (r *postgresClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	var plan models.PostgresClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"postgres", "create",
		"--name", plan.Name.ValueString(),
		"--org", plan.Org.ValueString(),
		"--region", plan.Region.ValueString(),
	}

	if v := plan.ClusterSize.ValueInt64(); v > 0 {
		args = append(args, "--cluster-size", fmt.Sprintf("%d", v))
	}
	if v := plan.VolumeSize.ValueInt64(); v > 0 {
		args = append(args, "--volume-size", fmt.Sprintf("%d", v))
	}
	if v := plan.VMSize.ValueString(); v != "" {
		args = append(args, "--vm-size", v)
	}
	if plan.EnableBackups.ValueBool() {
		args = append(args, "--enable-backups")
	}

	var result flyctlPostgresCluster
	err := r.flyctl.RunJSONMut(ctx, &result, args...)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Postgres cluster", err.Error())
		return
	}

	r.setModelFromAPI(&plan, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *postgresClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.PostgresClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var results []flyctlPostgresCluster
	err := r.flyctl.RunJSON(ctx, &results, "postgres", "list")
	if err != nil {
		resp.Diagnostics.AddError("Error reading Postgres clusters", err.Error())
		return
	}

	var found *flyctlPostgresCluster
	for i := range results {
		if results[i].Name == state.Name.ValueString() {
			found = &results[i]
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.setModelFromAPI(&state, found)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *postgresClusterResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	resp.Diagnostics.AddError("Update not supported", "All attributes of fly_postgres_cluster require replacement.")
}

func (r *postgresClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	var state models.PostgresClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.flyctl.RunMut(ctx, "postgres", "destroy", state.Name.ValueString(), "--yes")
	if err != nil {
		if flyctl.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error destroying Postgres cluster", err.Error())
	}
}

func (r *postgresClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *postgresClusterResource) setModelFromAPI(model *models.PostgresClusterResourceModel, api *flyctlPostgresCluster) {
	model.ID = types.StringValue(api.ID)
	if api.ID == "" {
		model.ID = types.StringValue(api.Name)
	}
	model.Name = types.StringValue(api.Name)
	model.Org = types.StringValue(api.Org)
	model.Region = types.StringValue(api.Region)
	model.ClusterSize = types.Int64Value(int64(api.ClusterSize))
	model.VolumeSize = types.Int64Value(int64(api.VolumeSize))
	model.VMSize = types.StringValue(api.VMSize)
	model.EnableBackups = types.BoolValue(api.EnableBackups)
	model.Status = types.StringValue(api.Status)
}

type flyctlPostgresCluster struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Org           string `json:"org"`
	Region        string `json:"region"`
	ClusterSize   int    `json:"cluster_size"`
	VolumeSize    int    `json:"volume_size"`
	VMSize        string `json:"vm_size"`
	EnableBackups bool   `json:"enable_backups"`
	Status        string `json:"status"`
}
