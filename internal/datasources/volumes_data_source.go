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
	_ datasource.DataSource              = &volumesDataSource{}
	_ datasource.DataSourceWithConfigure = &volumesDataSource{}
)

type volumesDataSource struct {
	client *apiclient.Client
}

func NewVolumesDataSource() datasource.DataSource {
	return &volumesDataSource{}
}

func (d *volumesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volumes"
}

func (d *volumesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists volumes in a Fly.io application.",
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				Description: "The application name.",
				Required:    true,
			},
			"volumes": schema.ListNestedAttribute{
				Description: "The list of volumes.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                  schema.StringAttribute{Computed: true, Description: "Volume ID."},
						"name":                schema.StringAttribute{Computed: true, Description: "Volume name."},
						"region":              schema.StringAttribute{Computed: true, Description: "Region."},
						"size_gb":             schema.Int32Attribute{Computed: true, Description: "Size in GB."},
						"encrypted":           schema.BoolAttribute{Computed: true, Description: "Whether encrypted."},
						"state":               schema.StringAttribute{Computed: true, Description: "Volume state."},
						"zone":                schema.StringAttribute{Computed: true, Description: "Zone."},
						"attached_machine_id": schema.StringAttribute{Computed: true, Description: "Attached machine ID."},
						"created_at":          schema.StringAttribute{Computed: true, Description: "Created timestamp."},
					},
				},
			},
		},
	}
}

func (d *volumesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *volumesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config models.VolumesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	volumes, err := d.client.ListVolumes(ctx, config.App.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing volumes", err.Error())
		return
	}

	config.Volumes = make([]models.VolumeItemModel, len(volumes))
	for i, v := range volumes {
		config.Volumes[i] = models.VolumeItemModel{
			ID:                types.StringValue(v.ID),
			Name:              types.StringValue(v.Name),
			Region:            types.StringValue(v.Region),
			SizeGB:            types.Int32Value(int32(v.SizeGB)),
			Encrypted:         types.BoolValue(v.Encrypted),
			State:             types.StringValue(v.State),
			Zone:              types.StringValue(v.Zone),
			AttachedMachineID: types.StringValue(v.AttachedMachineID),
			CreatedAt:         types.StringValue(v.CreatedAt),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
