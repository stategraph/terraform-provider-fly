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
	_ resource.Resource                = &wireGuardTokenResource{}
	_ resource.ResourceWithConfigure   = &wireGuardTokenResource{}
	_ resource.ResourceWithImportState = &wireGuardTokenResource{}
)

type wireGuardTokenResource struct {
	flyctl *flyctl.Executor
}

func NewWireGuardTokenResource() resource.Resource {
	return &wireGuardTokenResource{}
}

func (r *wireGuardTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wireguard_token"
}

func (r *wireGuardTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a delegated WireGuard token for a Fly.io organization. Import using org_slug/token_name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Synthetic identifier (org_slug/name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"org_slug": schema.StringAttribute{
				Description: "The organization slug. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Description: "The name of the token. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"token": schema.StringAttribute{
				Description: "The generated token value. Only available at creation time.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func (r *wireGuardTokenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_wireguard_token resource requires flyctl to be installed.")
		return
	}
	r.flyctl = pd.Flyctl
}

func (r *wireGuardTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.WireGuardTokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgSlug := plan.OrgSlug.ValueString()
	name := plan.Name.ValueString()

	var result flyctlWireGuardToken
	err := r.flyctl.RunJSON(ctx, &result, "wireguard", "token", "create", "--org", orgSlug, "--name", name)
	if err != nil {
		resp.Diagnostics.AddError("Error creating WireGuard token", err.Error())
		return
	}

	plan.ID = types.StringValue(orgSlug + "/" + name)
	plan.Token = types.StringValue(result.Token)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *wireGuardTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.WireGuardTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tokens []flyctlWireGuardToken
	err := r.flyctl.RunJSON(ctx, &tokens, "wireguard", "token", "list", "--org", state.OrgSlug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading WireGuard tokens", err.Error())
		return
	}

	var found bool
	for _, t := range tokens {
		if t.Name == state.Name.ValueString() {
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	// Token value is only available at creation, preserve from state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *wireGuardTokenResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "All attributes of fly_wireguard_token require replacement.")
}

func (r *wireGuardTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.WireGuardTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.flyctl.Run(ctx, "wireguard", "token", "delete", "--org", state.OrgSlug.ValueString(), "--name", state.Name.ValueString(), "--yes")
	if err != nil {
		resp.Diagnostics.AddError("Error deleting WireGuard token", err.Error())
	}
}

func (r *wireGuardTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected 'org_slug/token_name', got: %s", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_slug"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// flyctlWireGuardToken represents the JSON output from flyctl wireguard token commands.
type flyctlWireGuardToken struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token"`
}
