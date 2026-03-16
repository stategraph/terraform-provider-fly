package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type RegionsDataSourceModel struct {
	Regions []RegionModel `tfsdk:"regions"`
}

type RegionModel struct {
	Code    types.String `tfsdk:"code"`
	Name    types.String `tfsdk:"name"`
	Gateway types.Bool   `tfsdk:"gateway"`
}
