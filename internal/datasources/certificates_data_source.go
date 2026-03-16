package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stategraph/terraform-provider-fly/internal/models"
	"github.com/stategraph/terraform-provider-fly/pkg/apiclient"
)

var (
	_ datasource.DataSource              = &certificatesDataSource{}
	_ datasource.DataSourceWithConfigure = &certificatesDataSource{}
)

type certificatesDataSource struct {
	client *apiclient.Client
}

func NewCertificatesDataSource() datasource.DataSource {
	return &certificatesDataSource{}
}

func (d *certificatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificates"
}

func (d *certificatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists TLS certificates for a Fly.io application.",
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				Description: "The application name.",
				Required:    true,
			},
			"certificates": schema.ListNestedAttribute{
				Description: "The list of certificates.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                    schema.StringAttribute{Computed: true, Description: "Certificate ID."},
						"hostname":              schema.StringAttribute{Computed: true, Description: "Hostname."},
						"check_status":          schema.StringAttribute{Computed: true, Description: "Validation status."},
						"source":                schema.StringAttribute{Computed: true, Description: "Certificate source."},
						"certificate_authority": schema.StringAttribute{Computed: true, Description: "Certificate authority."},
						"created_at":            schema.StringAttribute{Computed: true, Description: "Created timestamp."},
					},
				},
			},
		},
	}
}

func (d *certificatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	d.client = pd.APIClient
}

func (d *certificatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config models.CertificatesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	certs, err := d.client.ListCertificates(ctx, config.App.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing certificates", err.Error())
		return
	}

	config.Certificates = make([]models.CertificateItemModel, len(certs))
	for i, c := range certs {
		config.Certificates[i] = models.CertificateItemModel{
			ID:                   types.StringValue(c.ID),
			Hostname:             types.StringValue(c.Hostname),
			CheckStatus:          types.StringValue(c.CheckStatus),
			Source:               types.StringValue(c.Source),
			CertificateAuthority: types.StringValue(c.CertificateAuthority),
			CreatedAt:            types.StringValue(c.CreatedAt),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
