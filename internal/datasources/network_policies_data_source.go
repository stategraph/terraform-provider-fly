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
	_ datasource.DataSource              = &networkPoliciesDataSource{}
	_ datasource.DataSourceWithConfigure = &networkPoliciesDataSource{}
)

type networkPoliciesDataSource struct {
	client *apiclient.Client
}

func NewNetworkPoliciesDataSource() datasource.DataSource {
	return &networkPoliciesDataSource{}
}

func (d *networkPoliciesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_policies"
}

func (d *networkPoliciesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists network policies for a Fly.io application.",
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				Description: "The application name.",
				Required:    true,
			},
			"policies": schema.ListNestedAttribute{
				Description: "The list of network policies.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":   schema.StringAttribute{Computed: true, Description: "Policy ID."},
						"name": schema.StringAttribute{Computed: true, Description: "Policy name."},
						"rule": schema.ListNestedAttribute{
							Description: "Traffic rules.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"action":    schema.StringAttribute{Computed: true, Description: "Rule action."},
									"direction": schema.StringAttribute{Computed: true, Description: "Traffic direction."},
									"port": schema.ListNestedAttribute{
										Description: "Port specifications.",
										Computed:    true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"protocol": schema.StringAttribute{Computed: true, Description: "Protocol."},
												"port":     schema.Int32Attribute{Computed: true, Description: "Port number."},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *networkPoliciesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *networkPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config models.NetworkPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policies, err := d.client.ListNetworkPolicies(ctx, config.App.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing network policies", err.Error())
		return
	}

	config.Policies = make([]models.NetworkPolicyListItemModel, len(policies))
	for i, p := range policies {
		item := models.NetworkPolicyListItemModel{
			ID:   types.StringValue(p.ID),
			Name: types.StringValue(p.Name),
		}
		item.Rules = make([]models.PolicyRuleModel, len(p.Rules))
		for j, rule := range p.Rules {
			ruleModel := models.PolicyRuleModel{
				Action:    types.StringValue(rule.Action),
				Direction: types.StringValue(rule.Direction),
			}
			ruleModel.Ports = make([]models.PolicyPortModel, len(rule.Ports))
			for k, port := range rule.Ports {
				ruleModel.Ports[k] = models.PolicyPortModel{
					Protocol: types.StringValue(port.Protocol),
					Port:     types.Int32Value(int32(port.Port)),
				}
			}
			item.Rules[j] = ruleModel
		}
		config.Policies[i] = item
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
