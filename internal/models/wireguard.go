package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type WireGuardPeerResourceModel struct {
	ID         types.String `tfsdk:"id"`
	OrgSlug    types.String `tfsdk:"org_slug"`
	Region     types.String `tfsdk:"region"`
	Name       types.String `tfsdk:"name"`
	PublicKey  types.String `tfsdk:"public_key"`
	Network    types.String `tfsdk:"network"`
	PeerIP     types.String `tfsdk:"peer_ip"`
	EndpointIP types.String `tfsdk:"endpoint_ip"`
	GatewayIP  types.String `tfsdk:"gateway_ip"`
}

type WireGuardTokenResourceModel struct {
	ID      types.String `tfsdk:"id"`
	OrgSlug types.String `tfsdk:"org_slug"`
	Name    types.String `tfsdk:"name"`
	Token   types.String `tfsdk:"token"`
}
