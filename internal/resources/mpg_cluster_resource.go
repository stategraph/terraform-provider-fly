package resources

import (
	"context"
	"fmt"
	"strings"

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
	_ resource.Resource                = &mpgClusterResource{}
	_ resource.ResourceWithConfigure   = &mpgClusterResource{}
	_ resource.ResourceWithImportState = &mpgClusterResource{}
)

type mpgClusterResource struct {
	flyctl *flyctl.Executor
}

func NewMPGClusterResource() resource.Resource {
	return &mpgClusterResource{}
}

func (r *mpgClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mpg_cluster"
}

func (r *mpgClusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Fly.io Managed Postgres (MPG) cluster. Import using org_slug/cluster_name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "The name of the cluster. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"org": schema.StringAttribute{
				Description: "The organization slug. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"region": schema.StringAttribute{
				Description: "The primary region. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"plan": schema.StringAttribute{
				Description: "The plan (e.g., free, starter, standard). Changing this forces a new resource.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
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
			"pg_major_version": schema.Int64Attribute{
				Description: "PostgreSQL major version. Changing this forces a new resource.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"enable_postgis": schema.BoolAttribute{
				Description: "Enable PostGIS extension. Changing this forces a new resource.",
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
			"primary_region": schema.StringAttribute{
				Description: "The primary region (computed from status).",
				Computed:    true,
			},
		},
	}
}

func (r *mpgClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_mpg_cluster resource requires flyctl to be installed.")
		return
	}
	r.flyctl = pd.Flyctl
}

func (r *mpgClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	var plan models.MPGClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"mpg", "create",
		"--name", plan.Name.ValueString(),
		"--org", plan.Org.ValueString(),
		"--region", plan.Region.ValueString(),
	}

	if v := plan.Plan.ValueString(); v != "" {
		args = append(args, "--plan", v)
	}
	if v := plan.VolumeSize.ValueInt64(); v > 0 {
		args = append(args, "--volume-size", fmt.Sprintf("%d", v))
	}
	if v := plan.PGMajorVersion.ValueInt64(); v > 0 {
		args = append(args, "--pg-major-version", fmt.Sprintf("%d", v))
	}
	if plan.EnablePostGIS.ValueBool() {
		args = append(args, "--enable-postgis")
	}

	_, err := r.flyctl.RunMut(ctx, args...)
	if err != nil {
		resp.Diagnostics.AddError("Error creating MPG cluster", err.Error())
		return
	}

	if r.flyctl.DryRun {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	cluster, err := r.findClusterByName(ctx, plan.Name.ValueString(), plan.Org.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading MPG cluster after creation", err.Error())
		return
	}
	if cluster == nil {
		resp.Diagnostics.AddError("Error finding MPG cluster after creation", "Cluster was created but not found in the list")
		return
	}

	r.setModelFromAPI(&plan, cluster)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mpgClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.MPGClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.findClusterByName(ctx, state.Name.ValueString(), state.Org.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading MPG cluster", err.Error())
		return
	}
	if cluster == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.setModelFromAPI(&state, cluster)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *mpgClusterResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	resp.Diagnostics.AddError("Update not supported", "All attributes of fly_mpg_cluster require replacement.")
}

func (r *mpgClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	var state models.MPGClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.flyctl.RunMut(ctx, "mpg", "destroy", state.ID.ValueString(), "--yes")
	if err != nil {
		if flyctl.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error destroying MPG cluster", err.Error())
	}
}

func (r *mpgClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
	} else {
		// Allow import by name only (org will be empty, works if FLY_ORG is set)
		resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
	}
}

func (r *mpgClusterResource) findClusterByName(ctx context.Context, name string, org string) (*flyctlMPGCluster, error) {
	args := []string{"mpg", "list"}
	if org != "" {
		args = append(args, "--org", org)
	}
	var results []flyctlMPGCluster
	err := r.flyctl.RunJSON(ctx, &results, args...)
	if err != nil {
		return nil, err
	}
	for i := range results {
		if results[i].Name == name {
			return &results[i], nil
		}
	}
	return nil, nil
}

func (r *mpgClusterResource) setModelFromAPI(model *models.MPGClusterResourceModel, api *flyctlMPGCluster) {
	id := api.ID
	if id == "" {
		id = api.Name
	}
	model.ID = types.StringValue(id)
	model.Name = types.StringValue(api.Name)
	model.Status = types.StringValue(api.Status)
	model.PrimaryRegion = types.StringValue(api.PrimaryRegion)
	model.Region = types.StringValue(api.Region)
	model.Plan = types.StringValue(api.Plan)
	model.VolumeSize = types.Int64Value(int64(api.VolumeSize))
	model.PGMajorVersion = types.Int64Value(int64(api.PGMajorVersion))
	model.EnablePostGIS = types.BoolValue(api.EnablePostGIS)
}

type flyctlMPGCluster struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Status         string `json:"status"`
	PrimaryRegion  string `json:"primary_region"`
	Region         string `json:"region"`
	Plan           string `json:"plan"`
	Org            string `json:"org"`
	VolumeSize     int    `json:"volume_size"`
	PGMajorVersion int    `json:"pg_major_version"`
	EnablePostGIS  bool   `json:"enable_postgis"`
}
