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
	_ datasource.DataSource              = &tokensDataSource{}
	_ datasource.DataSourceWithConfigure = &tokensDataSource{}
)

type tokensDataSource struct {
	flyctl *flyctl.Executor
}

func NewTokensDataSource() datasource.DataSource {
	return &tokensDataSource{}
}

func (d *tokensDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tokens"
}

func (d *tokensDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists Fly.io API tokens for an app or organization.",
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				Description: "The app to list tokens for. Mutually exclusive with org.",
				Optional:    true,
			},
			"org": schema.StringAttribute{
				Description: "The organization to list tokens for. Mutually exclusive with app.",
				Optional:    true,
			},
			"tokens": schema.ListNestedAttribute{
				Description: "The list of tokens.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true, Description: "Token ID."},
						"name":       schema.StringAttribute{Computed: true, Description: "Token name."},
						"type":       schema.StringAttribute{Computed: true, Description: "Token type."},
						"created_at": schema.StringAttribute{Computed: true, Description: "Token creation timestamp."},
					},
				},
			},
		},
	}
}

func (d *tokensDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	if pd.Flyctl == nil {
		resp.Diagnostics.AddError("flyctl required", "The fly_tokens data source requires flyctl to be installed.")
		return
	}
	d.flyctl = pd.Flyctl
}

// tokensDataSourceModel is the Terraform state model.
type tokensDataSourceModel struct {
	App    types.String `tfsdk:"app"`
	Org    types.String `tfsdk:"org"`
	Tokens []tokenModel `tfsdk:"tokens"`
}

type tokenModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (d *tokensDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config tokensDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"tokens", "list"}
	if !config.App.IsNull() && config.App.ValueString() != "" {
		args = append(args, "--app", config.App.ValueString())
	}
	if !config.Org.IsNull() && config.Org.ValueString() != "" {
		args = append(args, "--org", config.Org.ValueString())
	}

	var tokens []flyctlTokenDS
	err := d.flyctl.RunJSON(ctx, &tokens, args...)
	if err != nil {
		resp.Diagnostics.AddError("Error listing tokens", err.Error())
		return
	}

	config.Tokens = make([]tokenModel, len(tokens))
	for i, t := range tokens {
		config.Tokens[i] = tokenModel{
			ID:        types.StringValue(t.ID),
			Name:      types.StringValue(t.Name),
			Type:      types.StringValue(t.Type),
			CreatedAt: types.StringValue(t.CreatedAt),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// flyctlTokenDS represents the JSON output from flyctl tokens list.
type flyctlTokenDS struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
}
