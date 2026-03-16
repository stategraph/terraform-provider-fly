package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type CertificateResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	App                   types.String `tfsdk:"app"`
	Hostname              types.String `tfsdk:"hostname"`
	CheckStatus           types.String `tfsdk:"check_status"`
	DNSValidationHostname types.String `tfsdk:"dns_validation_hostname"`
	DNSValidationTarget   types.String `tfsdk:"dns_validation_target"`
	Source                types.String `tfsdk:"source"`
	IssuedAt              types.String `tfsdk:"issued_at"`
	CertificateAuthority  types.String `tfsdk:"certificate_authority"`
}

type CertificateDataSourceModel struct {
	App                   types.String `tfsdk:"app"`
	Hostname              types.String `tfsdk:"hostname"`
	ID                    types.String `tfsdk:"id"`
	CheckStatus           types.String `tfsdk:"check_status"`
	DNSValidationHostname types.String `tfsdk:"dns_validation_hostname"`
	DNSValidationTarget   types.String `tfsdk:"dns_validation_target"`
	Source                types.String `tfsdk:"source"`
	IssuedAt              types.String `tfsdk:"issued_at"`
	CertificateAuthority  types.String `tfsdk:"certificate_authority"`
}
