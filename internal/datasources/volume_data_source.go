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
	_ datasource.DataSource              = &volumeDataSource{}
	_ datasource.DataSourceWithConfigure = &volumeDataSource{}
)

type volumeDataSource struct {
	client *apiclient.Client
}

func NewVolumeDataSource() datasource.DataSource {
	return &volumeDataSource{}
}

func (d *volumeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (d *volumeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing Fly.io volume.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the volume.",
				Required:    true,
			},
			"app": schema.StringAttribute{
				Description: "The name of the app the volume belongs to.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the volume.",
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "The region where the volume is located.",
				Computed:    true,
			},
			"size_gb": schema.Int32Attribute{
				Description: "The size of the volume in GB.",
				Computed:    true,
			},
			"encrypted": schema.BoolAttribute{
				Description: "Whether the volume is encrypted.",
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "The current state of the volume.",
				Computed:    true,
			},
			"zone": schema.StringAttribute{
				Description: "The zone the volume is in.",
				Computed:    true,
			},
			"attached_machine_id": schema.StringAttribute{
				Description: "The ID of the machine the volume is attached to.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the volume was created.",
				Computed:    true,
			},
		},
	}
}

func (d *volumeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = pd.APIClient
}

func (d *volumeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config models.VolumeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	volume, err := d.client.GetVolume(ctx, config.App.ValueString(), config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading volume", err.Error())
		return
	}

	config.ID = types.StringValue(volume.ID)
	config.App = types.StringValue(volume.App)
	config.Name = types.StringValue(volume.Name)
	config.Region = types.StringValue(volume.Region)
	config.SizeGB = types.Int32Value(int32(volume.SizeGB))
	config.Encrypted = types.BoolValue(volume.Encrypted)
	config.State = types.StringValue(volume.State)
	config.Zone = types.StringValue(volume.Zone)
	config.AttachedMachineID = types.StringValue(volume.AttachedMachineID)
	config.CreatedAt = types.StringValue(volume.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
