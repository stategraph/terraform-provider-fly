package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type AppsDataSourceModel struct {
	OrgSlug types.String    `tfsdk:"org_slug"`
	Apps    []AppItemModel  `tfsdk:"apps"`
}

type AppItemModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	OrgSlug types.String `tfsdk:"org_slug"`
	Network types.String `tfsdk:"network"`
	Status  types.String `tfsdk:"status"`
}
