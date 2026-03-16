package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type VolumeResourceModel struct {
	ID                types.String `tfsdk:"id"`
	App               types.String `tfsdk:"app"`
	Name              types.String `tfsdk:"name"`
	Region            types.String `tfsdk:"region"`
	SizeGB            types.Int32  `tfsdk:"size_gb"`
	Encrypted         types.Bool   `tfsdk:"encrypted"`
	SnapshotID        types.String `tfsdk:"snapshot_id"`
	SourceVolumeID    types.String `tfsdk:"source_volume_id"`
	SnapshotRetention types.Int32  `tfsdk:"snapshot_retention"`
	AutoBackupEnabled types.Bool   `tfsdk:"auto_backup_enabled"`
	RequireUniqueZone types.Bool   `tfsdk:"require_unique_zone"`

	// Computed
	State             types.String `tfsdk:"state"`
	Zone              types.String `tfsdk:"zone"`
	AttachedMachineID types.String `tfsdk:"attached_machine_id"`
	CreatedAt         types.String `tfsdk:"created_at"`
}

type VolumeDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	App               types.String `tfsdk:"app"`
	Name              types.String `tfsdk:"name"`
	Region            types.String `tfsdk:"region"`
	SizeGB            types.Int32  `tfsdk:"size_gb"`
	Encrypted         types.Bool   `tfsdk:"encrypted"`
	State             types.String `tfsdk:"state"`
	Zone              types.String `tfsdk:"zone"`
	AttachedMachineID types.String `tfsdk:"attached_machine_id"`
	CreatedAt         types.String `tfsdk:"created_at"`
}
