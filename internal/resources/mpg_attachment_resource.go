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
	_ resource.Resource                = &mpgAttachmentResource{}
	_ resource.ResourceWithConfigure   = &mpgAttachmentResource{}
	_ resource.ResourceWithImportState = &mpgAttachmentResource{}
)

type mpgAttachmentResource struct {
	flyctl *flyctl.Executor
}

func NewMPGAttachmentResource() resource.Resource {
	return &mpgAttachmentResource{}
}

func (r *mpgAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mpg_attachment"
}

func (r *mpgAttachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Attaches a Fly.io Managed Postgres (MPG) cluster to an app. Import using cluster_id/app.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the attachment.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"cluster_id": schema.StringAttribute{
				Description: "The ID of the MPG cluster. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"app": schema.StringAttribute{
				Description: "The app to attach. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"database": schema.StringAttribute{
				Description: "The database to use for the attachment. Changing this forces a new resource.",
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
				Description: "The connection URI set on the app.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *mpgAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_mpg_attachment resource requires flyctl to be installed.")
		return
	}
	r.flyctl = pd.Flyctl
}

func (r *mpgAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	var plan models.MPGAttachmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"mpg", "attach",
		"--cluster-id", plan.ClusterID.ValueString(),
		"--app", plan.App.ValueString(),
	}

	if v := plan.Database.ValueString(); v != "" {
		args = append(args, "--database", v)
	}
	if v := plan.VariableName.ValueString(); v != "" {
		args = append(args, "--variable-name", v)
	}

	var result flyctlMPGAttachment
	err := r.flyctl.RunJSONMut(ctx, &result, args...)
	if err != nil {
		resp.Diagnostics.AddError("Error attaching MPG cluster", err.Error())
		return
	}

	r.setModelFromAPI(&plan, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mpgAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.MPGAttachmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// There is no direct read/status command for attachments.
	// Preserve the existing state as-is.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *mpgAttachmentResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	resp.Diagnostics.AddError("Update not supported", "All attributes of fly_mpg_attachment require replacement.")
}

func (r *mpgAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	var state models.MPGAttachmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.flyctl.RunMut(ctx, "mpg", "detach",
		"--cluster-id", state.ClusterID.ValueString(),
		"--app", state.App.ValueString(),
		"--yes",
	)
	if err != nil {
		if flyctl.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error detaching MPG cluster", err.Error())
	}
}

func (r *mpgAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: cluster_id/app")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), parts[1])...)
}

func (r *mpgAttachmentResource) setModelFromAPI(model *models.MPGAttachmentResourceModel, api *flyctlMPGAttachment) {
	model.ID = types.StringValue(model.ClusterID.ValueString() + "/" + model.App.ValueString())
	model.ConnectionURI = types.StringValue(api.ConnectionURI)
	model.VariableName = types.StringValue(api.VariableName)
	model.Database = types.StringValue(api.Database)
}

type flyctlMPGAttachment struct {
	ConnectionURI string `json:"connection_uri"`
	VariableName  string `json:"variable_name"`
	Database      string `json:"database"`
}
