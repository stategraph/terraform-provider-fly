package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type RedisResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Org            types.String `tfsdk:"org"`
	Region         types.String `tfsdk:"region"`
	Plan           types.String `tfsdk:"plan"`
	ReplicaRegions types.List   `tfsdk:"replica_regions"`
	EnableEviction types.Bool   `tfsdk:"enable_eviction"`
	Status         types.String `tfsdk:"status"`
	PrimaryURL     types.String `tfsdk:"primary_url"`
}
