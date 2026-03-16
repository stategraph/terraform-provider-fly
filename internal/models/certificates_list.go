package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type CertificatesDataSourceModel struct {
	App          types.String             `tfsdk:"app"`
	Certificates []CertificateItemModel  `tfsdk:"certificates"`
}

type CertificateItemModel struct {
	ID                    types.String `tfsdk:"id"`
	Hostname              types.String `tfsdk:"hostname"`
	CheckStatus           types.String `tfsdk:"check_status"`
	Source                types.String `tfsdk:"source"`
	CertificateAuthority  types.String `tfsdk:"certificate_authority"`
	CreatedAt             types.String `tfsdk:"created_at"`
}
