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
	_ datasource.DataSource              = &tigrisBucketsDataSource{}
	_ datasource.DataSourceWithConfigure = &tigrisBucketsDataSource{}
)

type tigrisBucketsDataSource struct {
	flyctl *flyctl.Executor
}

func NewTigrisBucketsDataSource() datasource.DataSource {
	return &tigrisBucketsDataSource{}
}

func (d *tigrisBucketsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tigris_buckets"
}

func (d *tigrisBucketsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists Fly.io Tigris storage buckets.",
		Attributes: map[string]schema.Attribute{
			"buckets": schema.ListNestedAttribute{
				Description: "The list of Tigris buckets.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":     schema.StringAttribute{Computed: true, Description: "Bucket ID."},
						"name":   schema.StringAttribute{Computed: true, Description: "Bucket name."},
						"public": schema.BoolAttribute{Computed: true, Description: "Whether the bucket is public."},
					},
				},
			},
		},
	}
}

func (d *tigrisBucketsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_tigris_buckets data source requires flyctl to be installed.")
		return
	}
	d.flyctl = pd.Flyctl
}

// tigrisBucketsDataSourceModel is the Terraform state model.
type tigrisBucketsDataSourceModel struct {
	Buckets []tigrisBucketModel `tfsdk:"buckets"`
}

type tigrisBucketModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Public types.Bool   `tfsdk:"public"`
}

func (d *tigrisBucketsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config tigrisBucketsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var buckets []flyctlTigrisBucketDS
	err := d.flyctl.RunJSON(ctx, &buckets, "storage", "list")
	if err != nil {
		resp.Diagnostics.AddError("Error listing Tigris buckets", err.Error())
		return
	}

	config.Buckets = make([]tigrisBucketModel, len(buckets))
	for i, b := range buckets {
		config.Buckets[i] = tigrisBucketModel{
			ID:     types.StringValue(b.ID),
			Name:   types.StringValue(b.Name),
			Public: types.BoolValue(b.Public),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// flyctlTigrisBucketDS represents the JSON output from flyctl storage list.
type flyctlTigrisBucketDS struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Public bool   `json:"public"`
}
