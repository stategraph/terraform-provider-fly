package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type MachinesDataSourceModel struct {
	App      types.String         `tfsdk:"app"`
	Machines []MachineItemModel   `tfsdk:"machines"`
}

type MachineItemModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Region     types.String `tfsdk:"region"`
	State      types.String `tfsdk:"state"`
	Image      types.String `tfsdk:"image"`
	PrivateIP  types.String `tfsdk:"private_ip"`
	InstanceID types.String `tfsdk:"instance_id"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}
