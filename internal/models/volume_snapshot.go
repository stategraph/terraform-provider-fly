package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type VolumeSnapshotResourceModel struct {
	ID        types.String `tfsdk:"id"`
	App       types.String `tfsdk:"app"`
	VolumeID  types.String `tfsdk:"volume_id"`
	Size      types.Int32  `tfsdk:"size"`
	Digest    types.String `tfsdk:"digest"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
}

type VolumeSnapshotsDataSourceModel struct {
	App       types.String               `tfsdk:"app"`
	VolumeID  types.String               `tfsdk:"volume_id"`
	Snapshots []VolumeSnapshotItemModel  `tfsdk:"snapshots"`
}

type VolumeSnapshotItemModel struct {
	ID        types.String `tfsdk:"id"`
	Size      types.Int32  `tfsdk:"size"`
	Digest    types.String `tfsdk:"digest"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
}
