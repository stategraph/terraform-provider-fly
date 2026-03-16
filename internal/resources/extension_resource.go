package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stategraph/terraform-provider-fly/internal/models"
	"github.com/stategraph/terraform-provider-fly/pkg/flyctl"
)

// ExtensionConfig describes how to build a generic extension resource.
type ExtensionConfig struct {
	TypeName    string // e.g., "mysql", "kubernetes", "sentry"
	Description string
	HasOrg      bool
	HasRegion   bool
	HasApp      bool
}

var (
	_ resource.Resource                = &extensionResource{}
	_ resource.ResourceWithConfigure   = &extensionResource{}
	_ resource.ResourceWithImportState = &extensionResource{}
)

type extensionResource struct {
	flyctl *flyctl.Executor
	config ExtensionConfig
}

// NewExtensionResource returns a factory function that creates an extension resource
// for the given configuration.
func NewExtensionResource(cfg ExtensionConfig) func() resource.Resource {
	return func() resource.Resource {
		return &extensionResource{config: cfg}
	}
}

func (r *extensionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ext_" + r.config.TypeName
}

func (r *extensionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description:   "The unique identifier of the extension.",
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": schema.StringAttribute{
			Description:   "The name of the extension. Changing this forces a new resource.",
			Required:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"status": schema.StringAttribute{
			Description: "The extension status.",
			Computed:    true,
		},
	}

	if r.config.HasOrg {
		attrs["org"] = schema.StringAttribute{
			Description:   "The organization slug. Changing this forces a new resource.",
			Required:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
		}
	}
	if r.config.HasRegion {
		attrs["region"] = schema.StringAttribute{
			Description:   "The region. Changing this forces a new resource.",
			Required:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
		}
	}
	if r.config.HasApp {
		attrs["app"] = schema.StringAttribute{
			Description:   "The app name. Changing this forces a new resource.",
			Required:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
		}
	}

	resp.Schema = schema.Schema{
		Description: r.config.Description,
		Attributes:  attrs,
	}
}

func (r *extensionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", fmt.Sprintf("The fly_ext_%s resource requires flyctl to be installed.", r.config.TypeName))
		return
	}
	r.flyctl = pd.Flyctl
}

// flyctlExtension is the JSON response struct for extension commands.
type flyctlExtension struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func getStringAttr(ctx context.Context, config tfsdk.Config, attrName string) (string, diag.Diagnostics) {
	var val types.String
	diags := config.GetAttribute(ctx, path.Root(attrName), &val)
	return val.ValueString(), diags
}

func getStringAttrFromPlan(ctx context.Context, plan tfsdk.Plan, attrName string) (string, diag.Diagnostics) {
	var val types.String
	diags := plan.GetAttribute(ctx, path.Root(attrName), &val)
	return val.ValueString(), diags
}

func getStringAttrFromState(ctx context.Context, state tfsdk.State, attrName string) (string, diag.Diagnostics) {
	var val types.String
	diags := state.GetAttribute(ctx, path.Root(attrName), &val)
	return val.ValueString(), diags
}

func (r *extensionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	name, diags := getStringAttrFromPlan(ctx, req.Plan, "name")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"ext", r.config.TypeName, "create", "--name", name}
	if r.config.HasOrg {
		org, d := getStringAttrFromPlan(ctx, req.Plan, "org")
		resp.Diagnostics.Append(d...)
		args = append(args, "--org", org)
	}
	if r.config.HasRegion {
		region, d := getStringAttrFromPlan(ctx, req.Plan, "region")
		resp.Diagnostics.Append(d...)
		args = append(args, "--region", region)
	}
	if r.config.HasApp {
		app, d := getStringAttrFromPlan(ctx, req.Plan, "app")
		resp.Diagnostics.Append(d...)
		args = append(args, "--app", app)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var result flyctlExtension
	err := r.flyctl.RunJSON(ctx, &result, args...)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error creating %s extension", r.config.TypeName), err.Error())
		return
	}

	r.writeStateFromAPI(ctx, &resp.State, &req.Plan, nil, &result, &resp.Diagnostics)
}

func (r *extensionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	name, diags := getStringAttrFromState(ctx, req.State, "name")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result flyctlExtension
	err := r.flyctl.RunJSON(ctx, &result, "ext", r.config.TypeName, "status", name)
	if err != nil {
		if flyctl.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading %s extension", r.config.TypeName), err.Error())
		return
	}

	r.writeStateFromAPI(ctx, &resp.State, nil, &req.State, &result, &resp.Diagnostics)
}

func (r *extensionResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", fmt.Sprintf("All attributes of fly_ext_%s require replacement.", r.config.TypeName))
}

func (r *extensionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	name, diags := getStringAttrFromState(ctx, req.State, "name")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.flyctl.Run(ctx, "ext", r.config.TypeName, "destroy", name, "--yes")
	if err != nil {
		if flyctl.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("Error destroying %s extension", r.config.TypeName), err.Error())
	}
}

func (r *extensionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// writeStateFromAPI writes API response fields into the Terraform state, preserving
// config-only attributes (org, region, app) from the plan or prior state.
func (r *extensionResource) writeStateFromAPI(ctx context.Context, state *tfsdk.State, plan *tfsdk.Plan, priorState *tfsdk.State, api *flyctlExtension, diags *diag.Diagnostics) {
	id := api.ID
	if id == "" {
		id = api.Name
	}
	diags.Append(state.SetAttribute(ctx, path.Root("id"), types.StringValue(id))...)
	if api.Name != "" {
		diags.Append(state.SetAttribute(ctx, path.Root("name"), types.StringValue(api.Name))...)
	}
	diags.Append(state.SetAttribute(ctx, path.Root("status"), types.StringValue(api.Status))...)

	// Preserve config-only attributes from plan or prior state.
	if r.config.HasOrg {
		var org string
		if plan != nil {
			org, _ = getStringAttrFromPlan(ctx, *plan, "org")
		} else if priorState != nil {
			org, _ = getStringAttrFromState(ctx, *priorState, "org")
		}
		diags.Append(state.SetAttribute(ctx, path.Root("org"), types.StringValue(org))...)
	}
	if r.config.HasRegion {
		var region string
		if plan != nil {
			region, _ = getStringAttrFromPlan(ctx, *plan, "region")
		} else if priorState != nil {
			region, _ = getStringAttrFromState(ctx, *priorState, "region")
		}
		diags.Append(state.SetAttribute(ctx, path.Root("region"), types.StringValue(region))...)
	}
	if r.config.HasApp {
		var app string
		if plan != nil {
			app, _ = getStringAttrFromPlan(ctx, *plan, "app")
		} else if priorState != nil {
			app, _ = getStringAttrFromState(ctx, *priorState, "app")
		}
		diags.Append(state.SetAttribute(ctx, path.Root("app"), types.StringValue(app))...)
	}
}
