package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stategraph/terraform-provider-fly/internal/models"
	"github.com/stategraph/terraform-provider-fly/pkg/flyctl"
)

var (
	_ resource.Resource                = &redisResource{}
	_ resource.ResourceWithConfigure   = &redisResource{}
	_ resource.ResourceWithImportState = &redisResource{}
)

type redisResource struct {
	flyctl *flyctl.Executor
}

func NewRedisResource() resource.Resource {
	return &redisResource{}
}

func (r *redisResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redis"
}

func (r *redisResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Fly.io Redis database. Import using the database name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "The unique identifier of the Redis database.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description:   "The name of the Redis database. Changing this forces a new resource.",
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
			"plan": schema.StringAttribute{
				Description: "The plan (e.g., free, starter, standard).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"replica_regions": schema.ListAttribute{
				Description: "List of replica regions.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"enable_eviction": schema.BoolAttribute{
				Description: "Enable eviction policy.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The Redis database status.",
				Computed:    true,
			},
			"primary_url": schema.StringAttribute{
				Description: "The primary connection URL.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func (r *redisResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_redis resource requires flyctl to be installed.")
		return
	}
	r.flyctl = pd.Flyctl
}

func (r *redisResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.RedisResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"redis", "create",
		"--name", plan.Name.ValueString(),
		"--org", plan.Org.ValueString(),
		"--region", plan.Region.ValueString(),
		"--json",
	}

	if v := plan.Plan.ValueString(); v != "" {
		args = append(args, "--plan", v)
	}
	if plan.EnableEviction.ValueBool() {
		args = append(args, "--enable-eviction")
	}

	var result flyctlRedis
	err := r.flyctl.RunJSON(ctx, &result, args...)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Redis database", err.Error())
		return
	}

	r.setModelFromAPI(ctx, &plan, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *redisResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.RedisResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result flyctlRedis
	err := r.flyctl.RunJSON(ctx, &result, "redis", "status", state.Name.ValueString(), "--json")
	if err != nil {
		if flyctl.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Redis database", err.Error())
		return
	}

	r.setModelFromAPI(ctx, &state, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *redisResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.RedisResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"redis", "update", plan.Name.ValueString(), "--json"}

	if v := plan.Plan.ValueString(); v != "" {
		args = append(args, "--plan", v)
	}
	if !plan.EnableEviction.IsNull() {
		if plan.EnableEviction.ValueBool() {
			args = append(args, "--enable-eviction")
		}
	}

	var result flyctlRedis
	err := r.flyctl.RunJSON(ctx, &result, args...)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Redis database", err.Error())
		return
	}

	r.setModelFromAPI(ctx, &plan, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *redisResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.RedisResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.flyctl.Run(ctx, "redis", "destroy", state.Name.ValueString(), "--yes")
	if err != nil {
		if flyctl.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error destroying Redis database", err.Error())
	}
}

func (r *redisResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *redisResource) setModelFromAPI(ctx context.Context, model *models.RedisResourceModel, api *flyctlRedis) {
	model.ID = types.StringValue(api.ID)
	if api.ID == "" {
		model.ID = types.StringValue(api.Name)
	}
	model.Name = types.StringValue(api.Name)
	model.Status = types.StringValue(api.Status)
	model.Region = types.StringValue(api.Region)
	model.Plan = types.StringValue(api.Plan)
	model.EnableEviction = types.BoolValue(api.EnableEviction)
	model.PrimaryURL = types.StringValue(api.PrimaryURL)
	if api.ReplicaRegions == nil {
		api.ReplicaRegions = []string{}
	}
	list, d := types.ListValueFrom(ctx, types.StringType, api.ReplicaRegions)
	if !d.HasError() {
		model.ReplicaRegions = list
	}
}

type flyctlRedis struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Status         string   `json:"status"`
	Region         string   `json:"region"`
	Plan           string   `json:"plan"`
	Org            string   `json:"org"`
	ReplicaRegions []string `json:"replica_regions"`
	EnableEviction bool     `json:"enable_eviction"`
	PrimaryURL     string   `json:"primary_url"`
}
