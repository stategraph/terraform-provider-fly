package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type OIDCTokenDataSourceModel struct {
	Aud   types.String `tfsdk:"aud"`
	Token types.String `tfsdk:"token"`
}
