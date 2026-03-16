package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type ExtMySQLResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Org    types.String `tfsdk:"org"`
	Region types.String `tfsdk:"region"`
	Status types.String `tfsdk:"status"`
}

type ExtKubernetesResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Org    types.String `tfsdk:"org"`
	Region types.String `tfsdk:"region"`
	Status types.String `tfsdk:"status"`
}

type ExtSentryResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	App    types.String `tfsdk:"app"`
	Status types.String `tfsdk:"status"`
}
