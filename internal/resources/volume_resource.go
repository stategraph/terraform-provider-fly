package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stategraph/terraform-provider-fly/internal/models"
	"github.com/stategraph/terraform-provider-fly/pkg/apiclient"
	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

var (
	_ resource.Resource                = &volumeResource{}
	_ resource.ResourceWithConfigure   = &volumeResource{}
	_ resource.ResourceWithImportState = &volumeResource{}
)

type volumeResource struct {
	client *apiclient.Client
}

func NewVolumeResource() resource.Resource {
	return &volumeResource{}
}

func (r *volumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (r *volumeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Fly.io volume. Import using app_name/volume_id: `terraform import fly_volume.example my-app/vol_abc123`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the volume.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app": schema.StringAttribute{
				Description: "The name of the app the volume belongs to. Changing this forces a new resource to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the volume. Changing this forces a new resource to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Description: "The region where the volume is created. Changing this forces a new resource to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"size_gb": schema.Int32Attribute{
				Description: "The size of the volume in GB. Can only be increased (extend-only).",
				Required:    true,
			},
			"encrypted": schema.BoolAttribute{
				Description: "Whether the volume is encrypted. Defaults to true. Changing this forces a new resource to be created.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"snapshot_id": schema.StringAttribute{
				Description: "The ID of the snapshot to restore from. Changing this forces a new resource to be created.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_volume_id": schema.StringAttribute{
				Description: "The ID of the source volume to fork from. Changing this forces a new resource to be created.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"snapshot_retention": schema.Int32Attribute{
				Description: "The number of snapshots to retain for the volume.",
				Optional:    true,
				Computed:    true,
			},
			"auto_backup_enabled": schema.BoolAttribute{
				Description: "Whether automatic backups are enabled for the volume.",
				Optional:    true,
				Computed:    true,
			},
			"require_unique_zone": schema.BoolAttribute{
				Description: "Whether the volume requires a unique zone. Changing this forces a new resource to be created.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				Description: "The current state of the volume.",
				Computed:    true,
			},
			"zone": schema.StringAttribute{
				Description: "The zone the volume is in.",
				Computed:    true,
			},
			"attached_machine_id": schema.StringAttribute{
				Description: "The ID of the machine the volume is attached to.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the volume was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *volumeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *volumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, r.client, nil)
	var plan models.VolumeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := apimodels.CreateVolumeRequest{
		Name:   plan.Name.ValueString(),
		Region: plan.Region.ValueString(),
		SizeGB: int(plan.SizeGB.ValueInt32()),
	}

	if !plan.Encrypted.IsNull() && !plan.Encrypted.IsUnknown() {
		v := plan.Encrypted.ValueBool()
		createReq.Encrypted = &v
	}
	if !plan.SnapshotID.IsNull() && !plan.SnapshotID.IsUnknown() {
		createReq.SnapshotID = plan.SnapshotID.ValueString()
	}
	if !plan.SourceVolumeID.IsNull() && !plan.SourceVolumeID.IsUnknown() {
		createReq.SourceVolumeID = plan.SourceVolumeID.ValueString()
	}
	if !plan.SnapshotRetention.IsNull() && !plan.SnapshotRetention.IsUnknown() {
		createReq.SnapshotRetention = int(plan.SnapshotRetention.ValueInt32())
	}
	if !plan.RequireUniqueZone.IsNull() && !plan.RequireUniqueZone.IsUnknown() {
		v := plan.RequireUniqueZone.ValueBool()
		createReq.RequireUniqueZone = &v
	}

	volume, err := r.client.CreateVolume(ctx, plan.App.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating volume", err.Error())
		return
	}

	r.setStateFromVolume(&plan, volume)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *volumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.VolumeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	volume, err := r.client.GetVolume(ctx, state.App.ValueString(), state.ID.ValueString())
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading volume", err.Error())
		return
	}

	r.setStateFromVolume(&state, volume)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *volumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, r.client, nil)
	var plan models.VolumeResourceModel
	var state models.VolumeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Handle size_gb changes (extend-only).
	if plan.SizeGB.ValueInt32() != state.SizeGB.ValueInt32() {
		if plan.SizeGB.ValueInt32() < state.SizeGB.ValueInt32() {
			resp.Diagnostics.AddError(
				"Cannot shrink volume",
				fmt.Sprintf("Volume size can only be increased. Current size: %d GB, requested size: %d GB.",
					state.SizeGB.ValueInt32(), plan.SizeGB.ValueInt32()),
			)
			return
		}

		extendResp, err := r.client.ExtendVolume(ctx, plan.App.ValueString(), state.ID.ValueString(), int(plan.SizeGB.ValueInt32()))
		if err != nil {
			resp.Diagnostics.AddError("Error extending volume", err.Error())
			return
		}

		r.setStateFromVolume(&plan, &extendResp.Volume)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	// Handle snapshot_retention and auto_backup_enabled changes via UpdateVolume.
	needsUpdate := false
	updateReq := apimodels.UpdateVolumeRequest{}

	if !plan.SnapshotRetention.Equal(state.SnapshotRetention) {
		v := int(plan.SnapshotRetention.ValueInt32())
		updateReq.SnapshotRetention = &v
		needsUpdate = true
	}
	if !plan.AutoBackupEnabled.Equal(state.AutoBackupEnabled) {
		v := plan.AutoBackupEnabled.ValueBool()
		updateReq.AutoBackupEnabled = &v
		needsUpdate = true
	}

	if needsUpdate {
		volume, err := r.client.UpdateVolume(ctx, plan.App.ValueString(), state.ID.ValueString(), updateReq)
		if err != nil {
			resp.Diagnostics.AddError("Error updating volume", err.Error())
			return
		}
		r.setStateFromVolume(&plan, volume)
	} else {
		volume, err := r.client.GetVolume(ctx, plan.App.ValueString(), state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error reading volume", err.Error())
			return
		}
		r.setStateFromVolume(&plan, volume)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *volumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer models.FlushDryRunWarnings(&resp.Diagnostics, r.client, nil)
	var state models.VolumeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteVolume(ctx, state.App.ValueString(), state.ID.ValueString())
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting volume", err.Error())
	}
}

func (r *volumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID format: app_name/volume_id, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func (r *volumeResource) setStateFromVolume(state *models.VolumeResourceModel, volume *apimodels.Volume) {
	state.ID = types.StringValue(volume.ID)
	// Preserve app from plan/state — the API doesn't always return it.
	if volume.App != "" {
		state.App = types.StringValue(volume.App)
	}
	state.Name = types.StringValue(volume.Name)
	state.Region = types.StringValue(volume.Region)
	state.SizeGB = types.Int32Value(int32(volume.SizeGB))
	state.Encrypted = types.BoolValue(volume.Encrypted)
	state.State = types.StringValue(volume.State)
	state.Zone = types.StringValue(volume.Zone)
	state.AttachedMachineID = types.StringValue(volume.AttachedMachineID)
	state.CreatedAt = types.StringValue(volume.CreatedAt)
	state.SnapshotRetention = types.Int32Value(int32(volume.SnapshotRetention))
	state.AutoBackupEnabled = types.BoolValue(volume.AutoBackupEnabled)
}
