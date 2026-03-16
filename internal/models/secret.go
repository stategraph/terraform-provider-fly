package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type SecretResourceModel struct {
	ID        types.String `tfsdk:"id"`
	App       types.String `tfsdk:"app"`
	Key       types.String `tfsdk:"key"`
	Value     types.String `tfsdk:"value"`
	Digest    types.String `tfsdk:"digest"`
	CreatedAt types.String `tfsdk:"created_at"`
}
