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

var _ datasource.DataSource = &machineDataSource{}
var _ datasource.DataSourceWithConfigure = &machineDataSource{}

type machineDataSource struct {
	client *apiclient.Client
}

func NewMachineDataSource() datasource.DataSource {
	return &machineDataSource{}
}

func (d *machineDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine"
}

func (d *machineDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a Fly.io machine.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Machine ID.",
				Required:    true,
			},
			"app": schema.StringAttribute{
				Description: "App name the machine belongs to.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Machine name.",
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "Region where the machine is deployed.",
				Computed:    true,
			},
			"image": schema.StringAttribute{
				Description: "Docker image the machine is running.",
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "Current machine state.",
				Computed:    true,
			},
			"private_ip": schema.StringAttribute{
				Description: "Private IPv6 address of the machine.",
				Computed:    true,
			},
			"instance_id": schema.StringAttribute{
				Description: "Instance ID of the machine.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When the machine was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the machine was last updated.",
				Computed:    true,
			},
		},
	}
}

func (d *machineDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *models.ProviderData, got %T", req.ProviderData))
		return
	}
	d.client = pd.APIClient
}

func (d *machineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.MachineDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	machine, err := d.client.GetMachine(ctx, data.App.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading machine", err.Error())
		return
	}

	data.Name = types.StringValue(machine.Name)
	data.Region = types.StringValue(machine.Region)
	data.Image = types.StringValue(machine.Config.Image)
	data.State = types.StringValue(machine.State)
	data.PrivateIP = types.StringValue(machine.PrivateIP)
	data.InstanceID = types.StringValue(machine.InstanceID)
	data.CreatedAt = types.StringValue(machine.CreatedAt)
	data.UpdatedAt = types.StringValue(machine.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
