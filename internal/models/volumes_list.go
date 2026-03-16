package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type VolumesDataSourceModel struct {
	App     types.String       `tfsdk:"app"`
	Volumes []VolumeItemModel  `tfsdk:"volumes"`
}

type VolumeItemModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Region            types.String `tfsdk:"region"`
	SizeGB            types.Int32  `tfsdk:"size_gb"`
	Encrypted         types.Bool   `tfsdk:"encrypted"`
	State             types.String `tfsdk:"state"`
	Zone              types.String `tfsdk:"zone"`
	AttachedMachineID types.String `tfsdk:"attached_machine_id"`
	CreatedAt         types.String `tfsdk:"created_at"`
}
