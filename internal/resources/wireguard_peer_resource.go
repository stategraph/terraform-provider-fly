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
	_ resource.Resource                = &wireGuardPeerResource{}
	_ resource.ResourceWithConfigure   = &wireGuardPeerResource{}
	_ resource.ResourceWithImportState = &wireGuardPeerResource{}
)

type wireGuardPeerResource struct {
	flyctl *flyctl.Executor
}

func NewWireGuardPeerResource() resource.Resource {
	return &wireGuardPeerResource{}
}

func (r *wireGuardPeerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wireguard_peer"
}

func (r *wireGuardPeerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a WireGuard peer for a Fly.io organization. Import using org_slug/peer_name.",
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
			"region": schema.StringAttribute{
				Description: "The region for the WireGuard peer. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Description: "The name of the WireGuard peer. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"public_key": schema.StringAttribute{
				Description: "The WireGuard public key. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"network": schema.StringAttribute{
				Description: "The network for the peer. Changing this forces a new resource.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"peer_ip": schema.StringAttribute{
				Description: "The assigned peer IP address.",
				Computed:    true,
			},
			"endpoint_ip": schema.StringAttribute{
				Description: "The WireGuard endpoint IP address.",
				Computed:    true,
			},
			"gateway_ip": schema.StringAttribute{
				Description: "The gateway IP address.",
				Computed:    true,
			},
		},
	}
}

func (r *wireGuardPeerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_wireguard_peer resource requires flyctl to be installed.")
		return
	}
	r.flyctl = pd.Flyctl
}

func (r *wireGuardPeerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	var plan models.WireGuardPeerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgSlug := plan.OrgSlug.ValueString()
	region := plan.Region.ValueString()
	name := plan.Name.ValueString()
	pubkey := plan.PublicKey.ValueString()

	args := []string{"wireguard", "create", orgSlug, region, name, "--pubkey", pubkey}
	if network := plan.Network.ValueString(); network != "" {
		args = append(args, "--network", network)
	}

	var peer flyctlWireGuardPeer
	err := r.flyctl.RunJSONMut(ctx, &peer, args...)
	if err != nil {
		resp.Diagnostics.AddError("Error creating WireGuard peer", err.Error())
		return
	}

	plan.ID = types.StringValue(orgSlug + "/" + name)
	plan.PeerIP = types.StringValue(peer.PeerIP)
	plan.EndpointIP = types.StringValue(peer.EndpointIP)
	plan.GatewayIP = types.StringValue(peer.GatewayIP)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *wireGuardPeerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.WireGuardPeerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var peers []flyctlWireGuardPeer
	err := r.flyctl.RunJSON(ctx, &peers, "wireguard", "list", "--org", state.OrgSlug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading WireGuard peers", err.Error())
		return
	}

	var found bool
	for _, peer := range peers {
		if peer.Name == state.Name.ValueString() {
			state.PeerIP = types.StringValue(peer.PeerIP)
			state.EndpointIP = types.StringValue(peer.EndpointIP)
			state.GatewayIP = types.StringValue(peer.GatewayIP)
			state.PublicKey = types.StringValue(peer.PublicKey)
			state.Region = types.StringValue(peer.Region)
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

func (r *wireGuardPeerResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	resp.Diagnostics.AddError("Update not supported", "All attributes of fly_wireguard_peer require replacement.")
}

func (r *wireGuardPeerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, nil, r.flyctl)
	var state models.WireGuardPeerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.flyctl.RunMut(ctx, "wireguard", "remove", state.OrgSlug.ValueString(), state.Name.ValueString(), "--yes")
	if err != nil {
		resp.Diagnostics.AddError("Error deleting WireGuard peer", err.Error())
	}
}

func (r *wireGuardPeerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected 'org_slug/peer_name', got: %s", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_slug"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// flyctlWireGuardPeer represents the JSON output from flyctl wireguard commands.
type flyctlWireGuardPeer struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Region     string `json:"region"`
	PublicKey  string `json:"pubkey"`
	Network    string `json:"network"`
	PeerIP     string `json:"peerip"`
	EndpointIP string `json:"endpointip"`
	GatewayIP  string `json:"gatewayip"`
}
