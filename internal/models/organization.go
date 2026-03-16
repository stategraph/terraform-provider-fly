package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type OrganizationDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Slug     types.String `tfsdk:"slug"`
	Type     types.String `tfsdk:"type"`
	PaidPlan types.Bool   `tfsdk:"paid_plan"`
}
