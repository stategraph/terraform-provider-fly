package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stategraph/terraform-provider-fly/internal/models"
	"github.com/stategraph/terraform-provider-fly/pkg/apiclient"
	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

var (
	_ resource.Resource                = &certificateResource{}
	_ resource.ResourceWithConfigure   = &certificateResource{}
	_ resource.ResourceWithImportState = &certificateResource{}
)

type certificateResource struct {
	client *apiclient.Client
}

func NewCertificateResource() resource.Resource {
	return &certificateResource{}
}

func (r *certificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (r *certificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TLS certificate for a Fly.io application. Import using app_name/hostname: `terraform import fly_certificate.example my-app/example.com`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the certificate.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app": schema.StringAttribute{
				Description: "The name of the application to add the certificate to. Changing this forces a new resource to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hostname": schema.StringAttribute{
				Description: "The hostname for the certificate. Changing this forces a new resource to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"check_status": schema.StringAttribute{
				Description: "The validation check status of the certificate.",
				Computed:    true,
			},
			"dns_validation_hostname": schema.StringAttribute{
				Description: "The hostname to use for DNS validation.",
				Computed:    true,
			},
			"dns_validation_target": schema.StringAttribute{
				Description: "The target to use for DNS validation.",
				Computed:    true,
			},
			"source": schema.StringAttribute{
				Description: "The source of the certificate.",
				Computed:    true,
			},
			"issued_at": schema.StringAttribute{
				Description: "The timestamp when the certificate was issued.",
				Computed:    true,
			},
			"certificate_authority": schema.StringAttribute{
				Description: "The certificate authority that issued the certificate.",
				Computed:    true,
			},
		},
	}
}

func (r *certificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = pd.APIClient
}

func (r *certificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.CertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addReq := apimodels.AddCertificateRequest{
		Hostname: plan.Hostname.ValueString(),
	}

	cert, err := r.client.AddCertificate(ctx, plan.App.ValueString(), addReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating certificate", err.Error())
		return
	}

	plan.ID = types.StringValue(cert.ID)
	plan.Hostname = types.StringValue(cert.Hostname)
	plan.CheckStatus = types.StringValue(cert.CheckStatus)
	plan.DNSValidationHostname = types.StringValue(cert.DNSValidationHostname)
	plan.DNSValidationTarget = types.StringValue(cert.DNSValidationTarget)
	plan.Source = types.StringValue(cert.Source)
	plan.IssuedAt = types.StringValue(cert.IssuedAt)
	plan.CertificateAuthority = types.StringValue(cert.CertificateAuthority)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *certificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.CertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cert, err := r.client.GetCertificate(ctx, state.App.ValueString(), state.Hostname.ValueString())
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading certificate", err.Error())
		return
	}

	state.ID = types.StringValue(cert.ID)
	state.Hostname = types.StringValue(cert.Hostname)
	state.CheckStatus = types.StringValue(cert.CheckStatus)
	state.DNSValidationHostname = types.StringValue(cert.DNSValidationHostname)
	state.DNSValidationTarget = types.StringValue(cert.DNSValidationTarget)
	state.Source = types.StringValue(cert.Source)
	state.IssuedAt = types.StringValue(cert.IssuedAt)
	state.CertificateAuthority = types.StringValue(cert.CertificateAuthority)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *certificateResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"All attributes of fly_certificate require replacement. Update should never be called.",
	)
}

func (r *certificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.CertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCertificate(ctx, state.App.ValueString(), state.Hostname.ValueString())
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting certificate", err.Error())
	}
}

func (r *certificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID in the format 'app_name/hostname', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("hostname"), parts[1])...)
}
