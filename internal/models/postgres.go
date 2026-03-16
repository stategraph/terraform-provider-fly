package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type PostgresClusterResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Org           types.String `tfsdk:"org"`
	Region        types.String `tfsdk:"region"`
	ClusterSize   types.Int64  `tfsdk:"cluster_size"`
	VolumeSize    types.Int64  `tfsdk:"volume_size"`
	VMSize        types.String `tfsdk:"vm_size"`
	EnableBackups types.Bool   `tfsdk:"enable_backups"`
	Status        types.String `tfsdk:"status"`
}

type PostgresAttachmentResourceModel struct {
	ID            types.String `tfsdk:"id"`
	PostgresApp   types.String `tfsdk:"postgres_app"`
	App           types.String `tfsdk:"app"`
	DatabaseName  types.String `tfsdk:"database_name"`
	VariableName  types.String `tfsdk:"variable_name"`
	ConnectionURI types.String `tfsdk:"connection_uri"`
}
