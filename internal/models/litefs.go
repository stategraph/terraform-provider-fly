package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type LiteFSClusterResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Org    types.String `tfsdk:"org"`
	Region types.String `tfsdk:"region"`
	Status types.String `tfsdk:"status"`
}
