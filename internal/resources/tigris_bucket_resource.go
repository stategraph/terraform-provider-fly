package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stategraph/terraform-provider-fly/internal/models"
	"github.com/stategraph/terraform-provider-fly/pkg/flyctl"
)

var (
	_ resource.Resource                = &tigrisBucketResource{}
	_ resource.ResourceWithConfigure   = &tigrisBucketResource{}
	_ resource.ResourceWithImportState = &tigrisBucketResource{}
)

type tigrisBucketResource struct {
	flyctl *flyctl.Executor
}

func NewTigrisBucketResource() resource.Resource {
	return &tigrisBucketResource{}
}

func (r *tigrisBucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tigris_bucket"
}

func (r *tigrisBucketResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Fly.io Tigris storage bucket. Import using the bucket name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "The unique identifier of the bucket.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description:   "The name of the bucket. Changing this forces a new resource.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"org": schema.StringAttribute{
				Description:   "The organization slug. Changing this forces a new resource.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"public": schema.BoolAttribute{
				Description: "Whether the bucket is publicly accessible.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"custom_domain": schema.StringAttribute{
				Description: "Custom domain for the bucket.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The bucket status.",
				Computed:    true,
			},
		},
	}
}

func (r *tigrisBucketResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_tigris_bucket resource requires flyctl to be installed.")
		return
	}
	r.flyctl = pd.Flyctl
}

func (r *tigrisBucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	var plan models.TigrisBucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"storage", "create",
		"--name", plan.Name.ValueString(),
		"--org", plan.Org.ValueString(),
		"--json",
	}

	if !plan.Public.IsNull() && plan.Public.ValueBool() {
		args = append(args, "--public")
	}

	var result flyctlTigrisBucket
	err := r.flyctl.RunJSONMut(ctx, &result, args...)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Tigris bucket", err.Error())
		return
	}

	r.setModelFromAPI(&plan, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *tigrisBucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.TigrisBucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result flyctlTigrisBucket
	err := r.flyctl.RunJSON(ctx, &result, "storage", "status", state.Name.ValueString(), "--json")
	if err != nil {
		if flyctl.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Tigris bucket", err.Error())
		return
	}

	r.setModelFromAPI(&state, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *tigrisBucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	var plan models.TigrisBucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"storage", "update", plan.Name.ValueString(), "--json"}

	if !plan.Public.IsNull() {
		if plan.Public.ValueBool() {
			args = append(args, "--public")
		}
	}
	if v := plan.CustomDomain.ValueString(); v != "" {
		args = append(args, "--custom-domain", v)
	}

	var result flyctlTigrisBucket
	err := r.flyctl.RunJSONMut(ctx, &result, args...)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Tigris bucket", err.Error())
		return
	}

	r.setModelFromAPI(&plan, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *tigrisBucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	var state models.TigrisBucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.flyctl.RunMut(ctx, "storage", "destroy", state.Name.ValueString(), "--yes")
	if err != nil {
		if flyctl.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error destroying Tigris bucket", err.Error())
	}
}

func (r *tigrisBucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *tigrisBucketResource) setModelFromAPI(model *models.TigrisBucketResourceModel, api *flyctlTigrisBucket) {
	model.ID = types.StringValue(api.ID)
	if api.ID == "" {
		model.ID = types.StringValue(api.Name)
	}
	model.Name = types.StringValue(api.Name)
	model.Status = types.StringValue(api.Status)
	model.Public = types.BoolValue(api.Public)
	model.CustomDomain = types.StringValue(api.CustomDomain)
}

type flyctlTigrisBucket struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Org          string `json:"org"`
	Public       bool   `json:"public"`
	CustomDomain string `json:"custom_domain"`
}
