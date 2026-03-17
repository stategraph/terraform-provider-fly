package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stategraph/terraform-provider-fly/internal/models"
	"github.com/stategraph/terraform-provider-fly/pkg/apiclient"
	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

var (
	_ resource.Resource                = &networkPolicyResource{}
	_ resource.ResourceWithConfigure   = &networkPolicyResource{}
	_ resource.ResourceWithImportState = &networkPolicyResource{}
)

type networkPolicyResource struct {
	client *apiclient.Client
}

func NewNetworkPolicyResource() resource.Resource {
	return &networkPolicyResource{}
}

func (r *networkPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_policy"
}

func (r *networkPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Fly.io network policy. Import using app_name/policy_id.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the network policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app": schema.StringAttribute{
				Description: "The name of the application. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the network policy.",
				Required:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"selector": schema.SingleNestedBlock{
				Description: "Selector to match which Machines the policy applies to.",
				Attributes: map[string]schema.Attribute{
					"all": schema.BoolAttribute{
						Description: "Apply to all Machines in the app.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"metadata": schema.MapAttribute{
						Description: "Match Machines by metadata key-value pairs.",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"rule": schema.ListNestedBlock{
				Description: "Traffic rules for the policy.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.StringAttribute{
							Description: "The rule action: 'allow'.",
							Required:    true,
						},
						"direction": schema.StringAttribute{
							Description: "Traffic direction: 'ingress' or 'egress'.",
							Required:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"port": schema.ListNestedBlock{
							Description: "Port and protocol specifications.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"protocol": schema.StringAttribute{
										Description: "Protocol: 'tcp' or 'udp'.",
										Required:    true,
									},
									"port": schema.Int32Attribute{
										Description: "Port number.",
										Required:    true,
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

func (r *networkPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*models.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *models.ProviderData, got: %T", req.ProviderData))
		return
	}
	r.client = pd.APIClient
}

func (r *networkPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, r.client, nil)
	var plan models.NetworkPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq, diags := networkPolicyModelToAPI(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.CreateNetworkPolicy(ctx, plan.App.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating network policy", err.Error())
		return
	}

	networkPolicyAPIToModel(ctx, policy, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *networkPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.NetworkPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetNetworkPolicy(ctx, state.App.ValueString(), state.ID.ValueString())
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading network policy", err.Error())
		return
	}

	networkPolicyAPIToModel(ctx, policy, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, r.client, nil)
	var plan models.NetworkPolicyResourceModel
	var state models.NetworkPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq, diags := networkPolicyModelToAPI(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateReq.ID = state.ID.ValueString()

	policy, err := r.client.CreateNetworkPolicy(ctx, plan.App.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating network policy", err.Error())
		return
	}

	networkPolicyAPIToModel(ctx, policy, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *networkPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, r.client, nil)
	var state models.NetworkPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNetworkPolicy(ctx, state.App.ValueString(), state.ID.ValueString())
	if err != nil && !apiclient.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting network policy", err.Error())
	}
}

func (r *networkPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected 'app_name/policy_id', got: %s", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func networkPolicyModelToAPI(ctx context.Context, m *models.NetworkPolicyResourceModel) (apimodels.CreateNetworkPolicyRequest, diag.Diagnostics) {
	var allDiags diag.Diagnostics

	req := apimodels.CreateNetworkPolicyRequest{
		Name: m.Name.ValueString(),
	}

	if m.Selector != nil {
		req.Selector.All = m.Selector.All.ValueBool()
		if !m.Selector.Metadata.IsNull() && !m.Selector.Metadata.IsUnknown() {
			metaMap := make(map[string]string)
			diags := m.Selector.Metadata.ElementsAs(ctx, &metaMap, false)
			allDiags.Append(diags...)
			req.Selector.Metadata = metaMap
		}
	}

	for _, rule := range m.Rules {
		apiRule := apimodels.PolicyRule{
			Action:    rule.Action.ValueString(),
			Direction: rule.Direction.ValueString(),
		}
		for _, p := range rule.Ports {
			apiRule.Ports = append(apiRule.Ports, apimodels.PolicyPort{
				Protocol: p.Protocol.ValueString(),
				Port:     int(p.Port.ValueInt32()),
			})
		}
		req.Rules = append(req.Rules, apiRule)
	}

	return req, allDiags
}

func networkPolicyAPIToModel(ctx context.Context, policy *apimodels.NetworkPolicy, m *models.NetworkPolicyResourceModel, diags *diag.Diagnostics) {
	m.ID = types.StringValue(policy.ID)
	m.Name = types.StringValue(policy.Name)

	m.Selector = &models.PolicySelectorModel{
		All: types.BoolValue(policy.Selector.All),
	}
	if len(policy.Selector.Metadata) > 0 {
		metaMap, d := types.MapValueFrom(ctx, types.StringType, policy.Selector.Metadata)
		diags.Append(d...)
		m.Selector.Metadata = metaMap
	} else {
		m.Selector.Metadata = types.MapNull(types.StringType)
	}

	m.Rules = make([]models.PolicyRuleModel, len(policy.Rules))
	for i, rule := range policy.Rules {
		ruleModel := models.PolicyRuleModel{
			Action:    types.StringValue(rule.Action),
			Direction: types.StringValue(rule.Direction),
		}
		ruleModel.Ports = make([]models.PolicyPortModel, len(rule.Ports))
		for j, p := range rule.Ports {
			ruleModel.Ports[j] = models.PolicyPortModel{
				Protocol: types.StringValue(p.Protocol),
				Port:     types.Int32Value(int32(p.Port)),
			}
		}
		m.Rules[i] = ruleModel
	}
}
