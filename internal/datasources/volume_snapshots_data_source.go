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
	_ datasource.DataSource              = &volumeSnapshotsDataSource{}
	_ datasource.DataSourceWithConfigure = &volumeSnapshotsDataSource{}
)

type volumeSnapshotsDataSource struct {
	client *apiclient.Client
}

func NewVolumeSnapshotsDataSource() datasource.DataSource {
	return &volumeSnapshotsDataSource{}
}

func (d *volumeSnapshotsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_snapshots"
}

func (d *volumeSnapshotsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists snapshots for a Fly.io volume.",
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				Description: "The application name.",
				Required:    true,
			},
			"volume_id": schema.StringAttribute{
				Description: "The volume ID.",
				Required:    true,
			},
			"snapshots": schema.ListNestedAttribute{
				Description: "The list of snapshots.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true, Description: "Snapshot ID."},
						"size":       schema.Int32Attribute{Computed: true, Description: "Snapshot size."},
						"digest":     schema.StringAttribute{Computed: true, Description: "Snapshot digest."},
						"status":     schema.StringAttribute{Computed: true, Description: "Snapshot status."},
						"created_at": schema.StringAttribute{Computed: true, Description: "Created timestamp."},
					},
				},
			},
		},
	}
}

func (d *volumeSnapshotsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *volumeSnapshotsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config models.VolumeSnapshotsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	snapshots, err := d.client.ListVolumeSnapshots(ctx, config.App.ValueString(), config.VolumeID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing volume snapshots", err.Error())
		return
	}

	config.Snapshots = make([]models.VolumeSnapshotItemModel, len(snapshots))
	for i, s := range snapshots {
		config.Snapshots[i] = models.VolumeSnapshotItemModel{
			ID:        types.StringValue(s.ID),
			Size:      types.Int32Value(int32(s.Size)),
			Digest:    types.StringValue(s.Digest),
			Status:    types.StringValue(s.Status),
			CreatedAt: types.StringValue(s.CreatedAt),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
