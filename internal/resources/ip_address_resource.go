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
	_ resource.Resource                = &ipAddressResource{}
	_ resource.ResourceWithConfigure   = &ipAddressResource{}
	_ resource.ResourceWithImportState = &ipAddressResource{}
)

type ipAddressResource struct {
	flyctl *flyctl.Executor
}

func NewIPAddressResource() resource.Resource {
	return &ipAddressResource{}
}

func (r *ipAddressResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address"
}

func (r *ipAddressResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Fly.io IP address allocation. Import using app_name/ip_address_id: `terraform import fly_ip_address.example my-app/ip-id`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the IP address allocation.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app": schema.StringAttribute{
				Description: "The name of the application to allocate the IP address for. Changing this forces a new resource to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The type of IP address to allocate (v4, v6, shared_v4, private_v6). Changing this forces a new resource to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Description: "The region to allocate the IP address in. Changing this forces a new resource to be created.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"address": schema.StringAttribute{
				Description: "The IP address that was allocated.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the IP address was allocated.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ipAddressResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_ip_address resource requires flyctl to be installed.")
		return
	}
	r.flyctl = pd.Flyctl
}

func (r *ipAddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.IPAddressResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addrType := plan.Type.ValueString()
	appName := plan.App.ValueString()

	// Map type to flyctl allocate subcommand.
	var allocateCmd string
	switch addrType {
	case "v4":
		allocateCmd = "allocate-v4"
	case "v6":
		allocateCmd = "allocate-v6"
	case "shared_v4":
		allocateCmd = "allocate-v4"
	case "private_v6":
		allocateCmd = "allocate-v6"
	default:
		allocateCmd = "allocate-v4"
	}

	args := []string{"ips", allocateCmd, "-a", appName}
	if addrType == "shared_v4" {
		args = append(args, "--shared")
	}
	if addrType == "private_v6" {
		args = append(args, "--private")
	}
	if region := plan.Region.ValueString(); region != "" {
		args = append(args, "--region", region)
	}

	// Allocate commands don't support --json, so use Run then list to get details.
	_, err := r.flyctl.Run(ctx, args...)
	if err != nil {
		resp.Diagnostics.AddError("Error allocating IP address", err.Error())
		return
	}

	// List IPs to find the newly allocated one.
	ips, err := r.listIPs(ctx, appName)
	if err != nil {
		resp.Diagnostics.AddError("Error reading IP addresses after allocation", err.Error())
		return
	}

	// Find the newest IP matching the requested type.
	var found *flyctlIPAddress
	for i := range ips {
		ip := &ips[i]
		if ip.Type == addrType {
			if found == nil || ip.CreatedAt > found.CreatedAt {
				found = ip
			}
		}
	}
	if found == nil {
		resp.Diagnostics.AddError("Error finding allocated IP", "IP address was allocated but not found in the list")
		return
	}

	plan.ID = types.StringValue(found.ID)
	plan.Address = types.StringValue(found.Address)
	plan.Type = types.StringValue(found.Type)
	plan.Region = types.StringValue(found.Region)
	plan.CreatedAt = types.StringValue(found.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipAddressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.IPAddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ips, err := r.listIPs(ctx, state.App.ValueString())
	if err != nil {
		if flyctl.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading IP addresses", err.Error())
		return
	}

	var found bool
	for _, ip := range ips {
		if ip.ID == state.ID.ValueString() {
			state.Address = types.StringValue(ip.Address)
			state.Type = types.StringValue(ip.Type)
			state.Region = types.StringValue(ip.Region)
			state.CreatedAt = types.StringValue(ip.CreatedAt)
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

func (r *ipAddressResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"All attributes of fly_ip_address require replacement. Update should never be called.",
	)
}

func (r *ipAddressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.IPAddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.flyctl.Run(ctx, "ips", "release", state.Address.ValueString(), "-a", state.App.ValueString())
	if err != nil {
		if flyctl.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error releasing IP address", err.Error())
	}
}

func (r *ipAddressResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID in the format 'app_name/ip_address_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func (r *ipAddressResource) listIPs(ctx context.Context, appName string) ([]flyctlIPAddress, error) {
	var ips []flyctlIPAddress
	err := r.flyctl.RunJSON(ctx, &ips, "ips", "list", "-a", appName)
	return ips, err
}

// flyctlIPAddress represents the JSON output from flyctl ips commands.
type flyctlIPAddress struct {
	ID        string `json:"id"`
	Address   string `json:"address"`
	Type      string `json:"type"`
	Region    string `json:"region"`
	CreatedAt string `json:"created_at"`
}
