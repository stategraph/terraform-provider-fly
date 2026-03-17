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
	"github.com/stategraph/terraform-provider-fly/pkg/apiclient"
	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

var (
	_ resource.Resource                = &appResource{}
	_ resource.ResourceWithConfigure   = &appResource{}
	_ resource.ResourceWithImportState = &appResource{}
)

type appResource struct {
	client *apiclient.Client
}

func NewAppResource() resource.Resource {
	return &appResource{}
}

func (r *appResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (r *appResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Fly.io application. Import using the app name: `terraform import fly_app.example my-app`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the application. Changing this forces a new resource to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"org_slug": schema.StringAttribute{
				Description: "The slug of the organization the app belongs to. Changing this forces a new resource to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network": schema.StringAttribute{
				Description: "The network to attach the app to. Changing this forces a new resource to be created.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_url": schema.StringAttribute{
				Description: "The URL of the application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The current status of the application.",
				Computed:    true,
			},
		},
	}
}

func (r *appResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData),
		)
		return
	}
	r.client = pd.APIClient
}

func (r *appResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, r.client, nil)
	var plan models.AppResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := apimodels.CreateAppRequest{
		AppName: plan.Name.ValueString(),
		OrgSlug: plan.OrgSlug.ValueString(),
		Network: plan.Network.ValueString(),
	}

	app, err := r.client.CreateApp(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating app", err.Error())
		return
	}

	plan.ID = types.StringValue(app.ID)
	// POST response may not include all fields — preserve plan values,
	// only override from API when populated.
	if app.Name != "" {
		plan.Name = types.StringValue(app.Name)
	}
	if app.OrgSlug != "" {
		plan.OrgSlug = types.StringValue(app.OrgSlug)
	}
	if app.Network != "" {
		plan.Network = types.StringValue(app.Network)
	} else if plan.Network.IsNull() || plan.Network.IsUnknown() {
		plan.Network = types.StringValue("default")
	}
	plan.AppURL = types.StringValue(fmt.Sprintf("https://fly.io/apps/%s", plan.Name.ValueString()))
	if app.Status != "" {
		plan.Status = types.StringValue(app.Status)
	} else {
		plan.Status = types.StringValue("pending")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *appResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.AppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.GetApp(ctx, state.Name.ValueString())
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading app", err.Error())
		return
	}

	state.ID = types.StringValue(app.ID)
	state.Name = types.StringValue(app.Name)
	state.OrgSlug = types.StringValue(app.OrgSlug)
	state.Network = types.StringValue(app.Network)
	state.AppURL = types.StringValue(fmt.Sprintf("https://fly.io/apps/%s", app.Name))
	state.Status = types.StringValue(app.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *appResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, r.client, nil)
	resp.Diagnostics.AddError(
		"Update not supported",
		"All attributes of fly_app require replacement. Update should never be called.",
	)
}

func (r *appResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, r.client, nil)
	var state models.AppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteApp(ctx, state.Name.ValueString())
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting app", err.Error())
	}
}

func (r *appResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
