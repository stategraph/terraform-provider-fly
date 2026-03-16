package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type TigrisBucketResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Org          types.String `tfsdk:"org"`
	Public       types.Bool   `tfsdk:"public"`
	CustomDomain types.String `tfsdk:"custom_domain"`
	Status       types.String `tfsdk:"status"`
}
