package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stategraph/terraform-provider-fly/internal/models"
	"github.com/stategraph/terraform-provider-fly/pkg/flyctl"
)

var (
	_ datasource.DataSource              = &ipAddressesDataSource{}
	_ datasource.DataSourceWithConfigure = &ipAddressesDataSource{}
)

type ipAddressesDataSource struct {
	flyctl *flyctl.Executor
}

func NewIPAddressesDataSource() datasource.DataSource {
	return &ipAddressesDataSource{}
}

func (d *ipAddressesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_addresses"
}

func (d *ipAddressesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves IP addresses allocated to a Fly.io application.",
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				Description: "The name of the application to look up IP addresses for.",
				Required:    true,
			},
			"ip_addresses": schema.ListNestedAttribute{
				Description: "The list of IP addresses allocated to the application.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the IP address allocation.",
							Computed:    true,
						},
						"address": schema.StringAttribute{
							Description: "The IP address.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of IP address (v4, v6, shared_v4, private_v6).",
							Computed:    true,
						},
						"region": schema.StringAttribute{
							Description: "The region of the IP address.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "The timestamp when the IP address was allocated.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ipAddressesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData),
		)
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_ip_addresses data source requires flyctl to be installed.")
		return
	}
	d.flyctl = pd.Flyctl
}

func (d *ipAddressesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config models.IPAddressesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ips []flyctlIPAddressDS
	err := d.flyctl.RunJSON(ctx, &ips, "ips", "list", "-a", config.App.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading IP addresses", err.Error())
		return
	}

	config.IPAddresses = make([]models.IPAddressModel, len(ips))
	for i, ip := range ips {
		config.IPAddresses[i] = models.IPAddressModel{
			ID:        types.StringValue(ip.ID),
			Address:   types.StringValue(ip.Address),
			Type:      types.StringValue(ip.Type),
			Region:    types.StringValue(ip.Region),
			CreatedAt: types.StringValue(ip.CreatedAt),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// flyctlIPAddressDS represents the JSON output from flyctl ips list.
type flyctlIPAddressDS struct {
	ID        string `json:"id"`
	Address   string `json:"address"`
	Type      string `json:"type"`
	Region    string `json:"region"`
	CreatedAt string `json:"created_at"`
}
