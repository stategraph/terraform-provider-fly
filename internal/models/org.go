package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type OrgResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Slug types.String `tfsdk:"slug"`
}

type OrgMemberResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Org   types.String `tfsdk:"org"`
	Email types.String `tfsdk:"email"`
}
