package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type AppResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	OrgSlug types.String `tfsdk:"org_slug"`
	Network types.String `tfsdk:"network"`
	AppURL  types.String `tfsdk:"app_url"`
	Status  types.String `tfsdk:"status"`
}

type AppDataSourceModel struct {
	Name     types.String `tfsdk:"name"`
	OrgSlug  types.String `tfsdk:"org_slug"`
	Network  types.String `tfsdk:"network"`
	Status   types.String `tfsdk:"status"`
	Hostname types.String `tfsdk:"hostname"`
}
