package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type MPGClusterResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Org            types.String `tfsdk:"org"`
	Region         types.String `tfsdk:"region"`
	Plan           types.String `tfsdk:"plan"`
	VolumeSize     types.Int64  `tfsdk:"volume_size"`
	PGMajorVersion types.Int64  `tfsdk:"pg_major_version"`
	EnablePostGIS  types.Bool   `tfsdk:"enable_postgis"`
	Status         types.String `tfsdk:"status"`
	PrimaryRegion  types.String `tfsdk:"primary_region"`
}

type MPGDatabaseResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ClusterID types.String `tfsdk:"cluster_id"`
	Name      types.String `tfsdk:"name"`
}

type MPGUserResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ClusterID types.String `tfsdk:"cluster_id"`
	Username  types.String `tfsdk:"username"`
	Role      types.String `tfsdk:"role"`
}

type MPGAttachmentResourceModel struct {
	ID           types.String `tfsdk:"id"`
	ClusterID    types.String `tfsdk:"cluster_id"`
	App          types.String `tfsdk:"app"`
	Database     types.String `tfsdk:"database"`
	VariableName types.String `tfsdk:"variable_name"`
	ConnectionURI types.String `tfsdk:"connection_uri"`
}
