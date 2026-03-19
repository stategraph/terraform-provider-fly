package resources

import (
	"context"
	"fmt"
	"strings"

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
	_ resource.Resource                = &mpgUserResource{}
	_ resource.ResourceWithConfigure   = &mpgUserResource{}
	_ resource.ResourceWithImportState = &mpgUserResource{}
)

type mpgUserResource struct {
	flyctl *flyctl.Executor
}

func NewMPGUserResource() resource.Resource {
	return &mpgUserResource{}
}

func (r *mpgUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mpg_user"
}

func (r *mpgUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a user in a Fly.io Managed Postgres (MPG) cluster. Import using cluster_id/username.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the user.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"cluster_id": schema.StringAttribute{
				Description: "The ID of the MPG cluster. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"username": schema.StringAttribute{
				Description: "The username. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"role": schema.StringAttribute{
				Description: "The role of the user.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *mpgUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_mpg_user resource requires flyctl to be installed.")
		return
	}
	r.flyctl = pd.Flyctl
}

func (r *mpgUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	var plan models.MPGUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"mpg", "users", "create",
		"--cluster-id", plan.ClusterID.ValueString(),
		"--username", plan.Username.ValueString(),
	}

	if v := plan.Role.ValueString(); v != "" {
		args = append(args, "--role", v)
	}

	_, err := r.flyctl.RunMut(ctx, args...)
	if err != nil {
		resp.Diagnostics.AddError("Error creating MPG user", err.Error())
		return
	}

	if r.flyctl.DryRun {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	var results []flyctlMPGUser
	err = r.flyctl.RunJSON(ctx, &results, "mpg", "users", "list", "--cluster-id", plan.ClusterID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading MPG users after creation", err.Error())
		return
	}

	var found *flyctlMPGUser
	for _, u := range results {
		if u.Username == plan.Username.ValueString() {
			found = &u
			break
		}
	}
	if found == nil {
		resp.Diagnostics.AddError("Error finding MPG user after creation", "User was created but not found in the list")
		return
	}

	r.setModelFromAPI(&plan, found)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mpgUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.MPGUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var results []flyctlMPGUser
	err := r.flyctl.RunJSON(ctx, &results, "mpg", "users", "list", "--cluster-id", state.ClusterID.ValueString())
	if err != nil {
		if flyctl.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading MPG users", err.Error())
		return
	}

	var found *flyctlMPGUser
	for _, u := range results {
		if u.Username == state.Username.ValueString() {
			found = &u
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

func (r *mpgUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	var plan models.MPGUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"mpg", "users", "set-role",
		"--cluster-id", plan.ClusterID.ValueString(),
		"--username", plan.Username.ValueString(),
		"--role", plan.Role.ValueString(),
	}

	_, err := r.flyctl.RunMut(ctx, args...)
	if err != nil {
		resp.Diagnostics.AddError("Error updating MPG user role", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mpgUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	var state models.MPGUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.flyctl.RunMut(ctx, "mpg", "users", "delete",
		"--cluster-id", state.ClusterID.ValueString(),
		"--username", state.Username.ValueString(),
		"--yes",
	)
	if err != nil {
		if flyctl.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting MPG user", err.Error())
	}
}

func (r *mpgUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: cluster_id/username")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("username"), parts[1])...)
}

func (r *mpgUserResource) setModelFromAPI(model *models.MPGUserResourceModel, api *flyctlMPGUser) {
	model.ID = types.StringValue(api.Username)
	model.Username = types.StringValue(api.Username)
	model.Role = types.StringValue(api.Role)
}

type flyctlMPGUser struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}
