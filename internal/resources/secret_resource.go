package resources

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
)

var (
	_ resource.Resource                = &secretResource{}
	_ resource.ResourceWithConfigure   = &secretResource{}
	_ resource.ResourceWithImportState = &secretResource{}
)

type secretResource struct {
	client *apiclient.Client
}

func NewSecretResource() resource.Resource {
	return &secretResource{}
}

func (r *secretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *secretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a secret for a Fly.io application. Import using app_name/SECRET_KEY: `terraform import fly_secret.example my-app/DATABASE_URL`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Synthetic identifier in the format app/key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app": schema.StringAttribute{
				Description: "The name of the Fly.io application. Changing this forces a new resource to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				Description: "The secret key name. Changing this forces a new resource to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Description: "The secret value. The API never returns this value. Write-only: not persisted in state.",
				Required:    true,
				Sensitive:   true,
				WriteOnly:   true,
			},
			"digest": schema.StringAttribute{
				Description: "SHA256 hex digest of the secret value, used for drift detection.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the secret was created.",
				Computed:    true,
			},
		},
	}
}

func (r *secretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *secretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.SecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appName := plan.App.ValueString()
	key := plan.Key.ValueString()
	value := plan.Value.ValueString()

	secret, err := r.client.SetSecret(ctx, appName, key, value)
	if err != nil {
		resp.Diagnostics.AddError("Error creating secret", err.Error())
		return
	}

	plan.ID = types.StringValue(appName + "/" + key)
	plan.Digest = types.StringValue(sha256Hex(value))

	// The set response returns the digest; fetch the list for created_at.
	_ = secret // digest from set response could be used, but we compute our own
	secrets, err := r.client.ListSecrets(ctx, appName)
	if err != nil {
		// Non-fatal: set created_at to empty rather than failing.
		plan.CreatedAt = types.StringValue("")
	} else {
		plan.CreatedAt = types.StringValue("")
		for _, s := range secrets {
			if s.EffectiveName() == key {
				plan.CreatedAt = types.StringValue(s.CreatedAt)
				break
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *secretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.SecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appName := state.App.ValueString()
	key := state.Key.ValueString()

	secrets, err := r.client.ListSecrets(ctx, appName)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading secrets", err.Error())
		return
	}

	var found bool
	for _, s := range secrets {
		if s.EffectiveName() == key {
			state.ID = types.StringValue(appName + "/" + key)
			state.Digest = types.StringValue(s.Digest)
			state.CreatedAt = types.StringValue(s.CreatedAt)
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *secretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.SecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appName := plan.App.ValueString()
	key := plan.Key.ValueString()
	value := plan.Value.ValueString()

	_, err := r.client.SetSecret(ctx, appName, key, value)
	if err != nil {
		resp.Diagnostics.AddError("Error updating secret", err.Error())
		return
	}

	plan.ID = types.StringValue(appName + "/" + key)
	plan.Digest = types.StringValue(sha256Hex(value))

	secrets, err := r.client.ListSecrets(ctx, appName)
	if err != nil {
		plan.CreatedAt = types.StringValue("")
	} else {
		plan.CreatedAt = types.StringValue("")
		for _, s := range secrets {
			if s.EffectiveName() == key {
				plan.CreatedAt = types.StringValue(s.CreatedAt)
				break
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *secretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.SecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSecret(ctx, state.App.ValueString(), state.Key.ValueString())
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting secret", err.Error())
	}
}

func (r *secretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in the format 'app_name/secret_key', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
