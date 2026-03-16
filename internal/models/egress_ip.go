package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type EgressIPResourceModel struct {
	ID      types.String `tfsdk:"id"`
	App     types.String `tfsdk:"app"`
	Address types.String `tfsdk:"address"`
	Version types.String `tfsdk:"version"`
	Region  types.String `tfsdk:"region"`
	City    types.String `tfsdk:"city"`
}
