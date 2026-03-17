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
	_ resource.Resource                = &egressIPResource{}
	_ resource.ResourceWithConfigure   = &egressIPResource{}
	_ resource.ResourceWithImportState = &egressIPResource{}
)

type egressIPResource struct {
	flyctl *flyctl.Executor
}

func NewEgressIPResource() resource.Resource {
	return &egressIPResource{}
}

func (r *egressIPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_egress_ip"
}

func (r *egressIPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Allocates a static egress IP address for a Fly.io application. Import using app_name/ip_id.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the egress IP allocation.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"app": schema.StringAttribute{
				Description: "The application name. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"address": schema.StringAttribute{
				Description: "The allocated egress IP address.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"version": schema.StringAttribute{
				Description: "The IP version (v4, v6).",
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "The region of the egress IP.",
				Computed:    true,
			},
			"city": schema.StringAttribute{
				Description: "The city of the egress IP.",
				Computed:    true,
			},
		},
	}
}

func (r *egressIPResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_egress_ip resource requires flyctl to be installed.")
		return
	}
	r.flyctl = pd.Flyctl
}

func (r *egressIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	var plan models.EgressIPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appName := plan.App.ValueString()

	// Allocate command doesn't support --json, use Run then list.
	_, err := r.flyctl.RunMut(ctx, "ips", "allocate-egress", "-a", appName, "--yes")
	if err != nil {
		resp.Diagnostics.AddError("Error allocating egress IP", err.Error())
		return
	}

	// List IPs to find the newly allocated egress IP.
	var ips []flyctlEgressIP
	err = r.flyctl.RunJSON(ctx, &ips, "ips", "list", "-a", appName)
	if err != nil {
		resp.Diagnostics.AddError("Error reading IPs after egress allocation", err.Error())
		return
	}

	var found *flyctlEgressIP
	for i := range ips {
		if ips[i].ID != "" {
			found = &ips[i]
			// Take the last one (newest)
		}
	}
	if found == nil {
		resp.Diagnostics.AddError("Error finding allocated egress IP", "Egress IP was allocated but not found in the list")
		return
	}

	plan.ID = types.StringValue(found.ID)
	plan.Address = types.StringValue(found.Address)
	plan.Version = types.StringValue(found.Version)
	plan.Region = types.StringValue(found.Region)
	plan.City = types.StringValue(found.City)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *egressIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.EgressIPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ips []flyctlEgressIP
	err := r.flyctl.RunJSON(ctx, &ips, "ips", "list", "-a", state.App.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading egress IPs", err.Error())
		return
	}

	var found bool
	for _, ip := range ips {
		if ip.ID == state.ID.ValueString() {
			state.Address = types.StringValue(ip.Address)
			state.Version = types.StringValue(ip.Version)
			state.Region = types.StringValue(ip.Region)
			state.City = types.StringValue(ip.City)
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *egressIPResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	resp.Diagnostics.AddError("Update not supported", "All attributes of fly_egress_ip require replacement.")
}

func (r *egressIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	var state models.EgressIPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.flyctl.RunMut(ctx, "ips", "release-egress", state.Address.ValueString(), "-a", state.App.ValueString())
	if err != nil {
		if flyctl.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error releasing egress IP", err.Error())
	}
}

func (r *egressIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected 'app_name/ip_id', got: %s", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// flyctlEgressIP represents the JSON output from flyctl egress IP commands.
type flyctlEgressIP struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Version string `json:"version"`
	Region  string `json:"region"`
	City    string `json:"city"`
}
