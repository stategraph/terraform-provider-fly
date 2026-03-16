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
	_ resource.Resource                = &postgresAttachmentResource{}
	_ resource.ResourceWithConfigure   = &postgresAttachmentResource{}
	_ resource.ResourceWithImportState = &postgresAttachmentResource{}
)

type postgresAttachmentResource struct {
	flyctl *flyctl.Executor
}

func NewPostgresAttachmentResource() resource.Resource {
	return &postgresAttachmentResource{}
}

func (r *postgresAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgres_attachment"
}

func (r *postgresAttachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Attaches a Fly.io Postgres cluster to an app. Import using postgres_app/app.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "The unique identifier of the attachment (postgres_app/app).",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"postgres_app": schema.StringAttribute{
				Description:   "The name of the Postgres app. Changing this forces a new resource.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"app": schema.StringAttribute{
				Description:   "The app to attach. Changing this forces a new resource.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"database_name": schema.StringAttribute{
				Description: "The database name to use for the attachment. Changing this forces a new resource.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"variable_name": schema.StringAttribute{
				Description: "The environment variable name for the connection URI. Changing this forces a new resource.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"connection_uri": schema.StringAttribute{
				Description:   "The connection URI set on the app.",
				Computed:      true,
				Sensitive:     true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *postgresAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_postgres_attachment resource requires flyctl to be installed.")
		return
	}
	r.flyctl = pd.Flyctl
}

func (r *postgresAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.PostgresAttachmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"postgres", "attach",
		plan.PostgresApp.ValueString(),
		"--app", plan.App.ValueString(),
	}

	if v := plan.DatabaseName.ValueString(); v != "" {
		args = append(args, "--database-name", v)
	}
	if v := plan.VariableName.ValueString(); v != "" {
		args = append(args, "--variable-name", v)
	}

	var result flyctlPostgresAttachment
	err := r.flyctl.RunJSON(ctx, &result, args...)
	if err != nil {
		resp.Diagnostics.AddError("Error attaching Postgres cluster", err.Error())
		return
	}

	r.setModelFromAPI(&plan, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *postgresAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.PostgresAttachmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// There is no direct read/status command for Postgres attachments.
	// Preserve the existing state as-is.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *postgresAttachmentResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "All attributes of fly_postgres_attachment require replacement.")
}

func (r *postgresAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.PostgresAttachmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.flyctl.Run(ctx, "postgres", "detach",
		state.PostgresApp.ValueString(),
		"--app", state.App.ValueString(),
		"--yes",
	)
	if err != nil {
		if flyctl.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error detaching Postgres cluster", err.Error())
	}
}

func (r *postgresAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: postgres_app/app")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("postgres_app"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), parts[1])...)
}

func (r *postgresAttachmentResource) setModelFromAPI(model *models.PostgresAttachmentResourceModel, api *flyctlPostgresAttachment) {
	model.ID = types.StringValue(model.PostgresApp.ValueString() + "/" + model.App.ValueString())
	model.ConnectionURI = types.StringValue(api.ConnectionURI)
	model.VariableName = types.StringValue(api.VariableName)
	model.DatabaseName = types.StringValue(api.DatabaseName)
}

type flyctlPostgresAttachment struct {
	ConnectionURI string `json:"connection_uri"`
	VariableName  string `json:"variable_name"`
	DatabaseName  string `json:"database_name"`
}
