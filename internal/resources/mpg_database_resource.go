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
	_ resource.Resource                = &mpgDatabaseResource{}
	_ resource.ResourceWithConfigure   = &mpgDatabaseResource{}
	_ resource.ResourceWithImportState = &mpgDatabaseResource{}
)

type mpgDatabaseResource struct {
	flyctl *flyctl.Executor
}

func NewMPGDatabaseResource() resource.Resource {
	return &mpgDatabaseResource{}
}

func (r *mpgDatabaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mpg_database"
}

func (r *mpgDatabaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a database in a Fly.io Managed Postgres (MPG) cluster. Import using cluster_id/database_name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the database.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"cluster_id": schema.StringAttribute{
				Description: "The ID of the MPG cluster. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Description: "The name of the database. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *mpgDatabaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_mpg_database resource requires flyctl to be installed.")
		return
	}
	r.flyctl = pd.Flyctl
}

func (r *mpgDatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	var plan models.MPGDatabaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"mpg", "databases", "create",
		plan.ClusterID.ValueString(),
		"--name", plan.Name.ValueString(),
	}

	_, err := r.flyctl.RunMut(ctx, args...)
	if err != nil {
		resp.Diagnostics.AddError("Error creating MPG database", err.Error())
		return
	}

	if r.flyctl.DryRun {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	var results []flyctlMPGDatabase
	err = r.flyctl.RunJSON(ctx, &results, "mpg", "databases", "list", plan.ClusterID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading MPG databases after creation", err.Error())
		return
	}

	var found *flyctlMPGDatabase
	for _, db := range results {
		if db.Name == plan.Name.ValueString() {
			found = &db
			break
		}
	}
	if found == nil {
		resp.Diagnostics.AddError("Error finding MPG database after creation", "Database was created but not found in the list")
		return
	}

	r.setModelFromAPI(&plan, found)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mpgDatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.MPGDatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var results []flyctlMPGDatabase
	err := r.flyctl.RunJSON(ctx, &results, "mpg", "databases", "list", state.ClusterID.ValueString())
	if err != nil {
		if flyctl.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading MPG databases", err.Error())
		return
	}

	var found *flyctlMPGDatabase
	for _, db := range results {
		if db.Name == state.Name.ValueString() {
			found = &db
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

func (r *mpgDatabaseResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	resp.Diagnostics.AddError("Update not supported", "All attributes of fly_mpg_database require replacement.")
}

func (r *mpgDatabaseResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	// flyctl does not support deleting individual databases; just remove from state.
}

func (r *mpgDatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: cluster_id/database_name")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func (r *mpgDatabaseResource) setModelFromAPI(model *models.MPGDatabaseResourceModel, api *flyctlMPGDatabase) {
	model.ID = types.StringValue(api.Name)
	model.Name = types.StringValue(api.Name)
}

type flyctlMPGDatabase struct {
	Name string `json:"name"`
}
