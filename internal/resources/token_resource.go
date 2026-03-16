package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stategraph/terraform-provider-fly/internal/models"
	"github.com/stategraph/terraform-provider-fly/pkg/flyctl"
)

var (
	_ resource.Resource              = &tokenResource{}
	_ resource.ResourceWithConfigure = &tokenResource{}
)

type tokenResource struct {
	flyctl *flyctl.Executor
}

func NewTokenResource() resource.Resource {
	return &tokenResource{}
}

func (r *tokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_token"
}

func (r *tokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Fly.io deploy or org token. Import is not supported; all changes force replacement.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "The unique identifier of the token.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"type": schema.StringAttribute{
				Description:   "The token type (e.g., deploy, org). Changing this forces a new resource.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"app": schema.StringAttribute{
				Description:   "The app name (for deploy tokens). Changing this forces a new resource.",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"org": schema.StringAttribute{
				Description:   "The organization slug (for org tokens). Changing this forces a new resource.",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Description:   "The token name. Changing this forces a new resource.",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"expiry": schema.StringAttribute{
				Description:   "The token expiry duration (e.g., 720h). Changing this forces a new resource.",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"token": schema.StringAttribute{
				Description: "The generated token value.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func (r *tokenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_token resource requires flyctl to be installed.")
		return
	}
	r.flyctl = pd.Flyctl
}

func (r *tokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.TokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"tokens", "create", plan.Type.ValueString(), "--json"}

	if v := plan.App.ValueString(); v != "" {
		args = append(args, "--app", v)
	}
	if v := plan.Org.ValueString(); v != "" {
		args = append(args, "--org", v)
	}
	if v := plan.Name.ValueString(); v != "" {
		args = append(args, "--name", v)
	}
	if v := plan.Expiry.ValueString(); v != "" {
		args = append(args, "--expiry", v)
	}

	var result flyctlToken
	err := r.flyctl.RunJSON(ctx, &result, args...)
	if err != nil {
		resp.Diagnostics.AddError("Error creating token", err.Error())
		return
	}

	r.setModelFromAPI(&plan, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *tokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.TokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"tokens", "list", "--json"}
	if v := state.App.ValueString(); v != "" {
		args = append(args, "--app", v)
	}
	if v := state.Org.ValueString(); v != "" {
		args = append(args, "--org", v)
	}

	out, err := r.flyctl.Run(ctx, args...)
	if err != nil {
		if flyctl.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading token", err.Error())
		return
	}

	var tokens []flyctlToken
	if err := json.Unmarshal(out.Stdout, &tokens); err != nil {
		resp.Diagnostics.AddError("Error parsing token list", err.Error())
		return
	}

	var found *flyctlToken
	for _, t := range tokens {
		if t.ID == state.ID.ValueString() {
			t := t
			found = &t
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Preserve the token value from state since list doesn't return it.
	savedToken := state.Token
	r.setModelFromAPI(&state, found)
	state.Token = savedToken
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *tokenResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "All attributes of fly_token require replacement.")
}

func (r *tokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.TokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.flyctl.Run(ctx, "tokens", "revoke", state.ID.ValueString(), "--yes")
	if err != nil {
		if flyctl.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error revoking token", err.Error())
	}
}

func (r *tokenResource) setModelFromAPI(model *models.TokenResourceModel, api *flyctlToken) {
	model.ID = types.StringValue(api.ID)
	if api.ID == "" {
		model.ID = types.StringValue(api.Name)
	}
	model.Name = types.StringValue(api.Name)
	model.Token = types.StringValue(api.Token)
}

type flyctlToken struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	App    string `json:"app"`
	Org    string `json:"org"`
	Expiry string `json:"expiry"`
	Token  string `json:"token"`
}
