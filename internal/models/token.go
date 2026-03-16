package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type TokenResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Type   types.String `tfsdk:"type"`
	App    types.String `tfsdk:"app"`
	Org    types.String `tfsdk:"org"`
	Name   types.String `tfsdk:"name"`
	Expiry types.String `tfsdk:"expiry"`
	Token  types.String `tfsdk:"token"`
}
