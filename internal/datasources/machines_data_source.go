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
	_ datasource.DataSource              = &machinesDataSource{}
	_ datasource.DataSourceWithConfigure = &machinesDataSource{}
)

type machinesDataSource struct {
	client *apiclient.Client
}

func NewMachinesDataSource() datasource.DataSource {
	return &machinesDataSource{}
}

func (d *machinesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machines"
}

func (d *machinesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists machines in a Fly.io application.",
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				Description: "The application name.",
				Required:    true,
			},
			"machines": schema.ListNestedAttribute{
				Description: "The list of machines.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true, Description: "Machine ID."},
						"name":        schema.StringAttribute{Computed: true, Description: "Machine name."},
						"region":      schema.StringAttribute{Computed: true, Description: "Region."},
						"state":       schema.StringAttribute{Computed: true, Description: "Current state."},
						"image":       schema.StringAttribute{Computed: true, Description: "Docker image."},
						"private_ip":  schema.StringAttribute{Computed: true, Description: "Private IP."},
						"instance_id": schema.StringAttribute{Computed: true, Description: "Instance ID."},
						"created_at":  schema.StringAttribute{Computed: true, Description: "Created timestamp."},
						"updated_at":  schema.StringAttribute{Computed: true, Description: "Updated timestamp."},
					},
				},
			},
		},
	}
}

func (d *machinesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *machinesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config models.MachinesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	machines, err := d.client.ListMachines(ctx, config.App.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing machines", err.Error())
		return
	}

	config.Machines = make([]models.MachineItemModel, len(machines))
	for i, m := range machines {
		config.Machines[i] = models.MachineItemModel{
			ID:         types.StringValue(m.ID),
			Name:       types.StringValue(m.Name),
			Region:     types.StringValue(m.Region),
			State:      types.StringValue(m.State),
			Image:      types.StringValue(m.Config.Image),
			PrivateIP:  types.StringValue(m.PrivateIP),
			InstanceID: types.StringValue(m.InstanceID),
			CreatedAt:  types.StringValue(m.CreatedAt),
			UpdatedAt:  types.StringValue(m.UpdatedAt),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
