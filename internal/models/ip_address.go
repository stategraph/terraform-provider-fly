package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type IPAddressResourceModel struct {
	ID        types.String `tfsdk:"id"`
	App       types.String `tfsdk:"app"`
	Type      types.String `tfsdk:"type"`
	Region    types.String `tfsdk:"region"`
	Address   types.String `tfsdk:"address"`
	CreatedAt types.String `tfsdk:"created_at"`
}

type IPAddressModel struct {
	ID        types.String `tfsdk:"id"`
	Address   types.String `tfsdk:"address"`
	Type      types.String `tfsdk:"type"`
	Region    types.String `tfsdk:"region"`
	CreatedAt types.String `tfsdk:"created_at"`
}

type IPAddressesDataSourceModel struct {
	App         types.String     `tfsdk:"app"`
	IPAddresses []IPAddressModel `tfsdk:"ip_addresses"`
}
