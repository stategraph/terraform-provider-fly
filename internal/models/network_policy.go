package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type NetworkPolicyResourceModel struct {
	ID       types.String         `tfsdk:"id"`
	App      types.String         `tfsdk:"app"`
	Name     types.String         `tfsdk:"name"`
	Selector *PolicySelectorModel `tfsdk:"selector"`
	Rules    []PolicyRuleModel    `tfsdk:"rule"`
}

type PolicySelectorModel struct {
	All      types.Bool `tfsdk:"all"`
	Metadata types.Map  `tfsdk:"metadata"`
}

type PolicyRuleModel struct {
	Action    types.String      `tfsdk:"action"`
	Direction types.String      `tfsdk:"direction"`
	Ports     []PolicyPortModel `tfsdk:"port"`
}

type PolicyPortModel struct {
	Protocol types.String `tfsdk:"protocol"`
	Port     types.Int32  `tfsdk:"port"`
}

type NetworkPolicyDataSourceModel struct {
	App      types.String                 `tfsdk:"app"`
	Policies []NetworkPolicyListItemModel `tfsdk:"policies"`
}

type NetworkPolicyListItemModel struct {
	ID    types.String      `tfsdk:"id"`
	Name  types.String      `tfsdk:"name"`
	Rules []PolicyRuleModel `tfsdk:"rule"`
}
