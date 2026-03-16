package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stategraph/terraform-provider-fly/internal/models"
	"github.com/stategraph/terraform-provider-fly/pkg/flyctl"
)

var (
	_ resource.Resource                = &liteFSClusterResource{}
	_ resource.ResourceWithConfigure   = &liteFSClusterResource{}
	_ resource.ResourceWithImportState = &liteFSClusterResource{}
)

type liteFSClusterResource struct {
	flyctl *flyctl.Executor
}

func NewLiteFSClusterResource() resource.Resource {
	return &liteFSClusterResource{}
}

func (r *liteFSClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_litefs_cluster"
}

func (r *liteFSClusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Fly.io LiteFS Cloud cluster. Import using the cluster name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "The unique identifier of the LiteFS cluster.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description:   "The name of the LiteFS cluster. Changing this forces a new resource.",
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
			"status": schema.StringAttribute{
				Description: "The cluster status.",
				Computed:    true,
			},
		},
	}
}

func (r *liteFSClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_litefs_cluster resource requires flyctl to be installed.")
		return
	}
	r.flyctl = pd.Flyctl
}

func (r *liteFSClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.LiteFSClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"litefs-cloud", "clusters", "create",
		"--name", plan.Name.ValueString(),
		"--org", plan.Org.ValueString(),
		"--region", plan.Region.ValueString(),
	}

	var result flyctlLiteFSCluster
	err := r.flyctl.RunJSON(ctx, &result, args...)
	if err != nil {
		resp.Diagnostics.AddError("Error creating LiteFS cluster", err.Error())
		return
	}

	r.setModelFromAPI(&plan, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *liteFSClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.LiteFSClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var results []flyctlLiteFSCluster
	err := r.flyctl.RunJSON(ctx, &results, "litefs-cloud", "clusters", "list")
	if err != nil {
		resp.Diagnostics.AddError("Error reading LiteFS clusters", err.Error())
		return
	}

	var found *flyctlLiteFSCluster
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

func (r *liteFSClusterResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "All attributes of fly_litefs_cluster require replacement.")
}

func (r *liteFSClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.LiteFSClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.flyctl.Run(ctx, "litefs-cloud", "clusters", "destroy", state.Name.ValueString(), "--yes")
	if err != nil {
		if flyctl.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error destroying LiteFS cluster", err.Error())
	}
}

func (r *liteFSClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *liteFSClusterResource) setModelFromAPI(model *models.LiteFSClusterResourceModel, api *flyctlLiteFSCluster) {
	model.ID = types.StringValue(api.ID)
	if api.ID == "" {
		model.ID = types.StringValue(api.Name)
	}
	model.Name = types.StringValue(api.Name)
	if api.Org != "" {
		model.Org = types.StringValue(api.Org)
	}
	if api.Region != "" {
		model.Region = types.StringValue(api.Region)
	}
	model.Status = types.StringValue(api.Status)
}

type flyctlLiteFSCluster struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Org    string `json:"org"`
	Region string `json:"region"`
	Status string `json:"status"`
}
