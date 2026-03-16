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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stategraph/terraform-provider-fly/internal/models"
	"github.com/stategraph/terraform-provider-fly/pkg/apiclient"
	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

var (
	_ resource.Resource                = &machineResource{}
	_ resource.ResourceWithConfigure   = &machineResource{}
	_ resource.ResourceWithImportState = &machineResource{}
)

type machineResource struct {
	client *apiclient.Client
}

func NewMachineResource() resource.Resource {
	return &machineResource{}
}

func (r *machineResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine"
}

func (r *machineResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Fly.io machine. Import using app_name/machine_id: `terraform import fly_machine.example my-app/machine-id`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the machine.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app": schema.StringAttribute{
				Description: "The name of the Fly.io application this machine belongs to. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the machine. Changing this forces a new resource.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.StringAttribute{
				Description: "The region to deploy the machine in (e.g., 'ord', 'iad'). Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image": schema.StringAttribute{
				Description: "The Docker image to run on the machine.",
				Required:    true,
			},
			"env": schema.MapAttribute{
				Description: "Environment variables to set on the machine.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"auto_destroy": schema.BoolAttribute{
				Description: "Whether to automatically destroy the machine when it exits.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"metadata": schema.MapAttribute{
				Description: "Metadata key-value pairs for the machine.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"desired_status": schema.StringAttribute{
				Description: "The desired status of the machine: 'started', 'stopped', or 'suspended'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("started"),
			},
			"cordoned": schema.BoolAttribute{
				Description: "Whether the machine is cordoned (will not accept new connections).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"cmd": schema.ListAttribute{
				Description: "The command to run on the machine.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"entrypoint": schema.ListAttribute{
				Description: "The entrypoint for the machine.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"skip_launch": schema.BoolAttribute{
				Description: "If true, the machine will be created but not started.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"schedule": schema.StringAttribute{
				Description: "Cron-like schedule for the machine.",
				Optional:    true,
			},
			"instance_id": schema.StringAttribute{
				Description: "The instance ID of the machine.",
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "The current state of the machine.",
				Computed:    true,
			},
			"private_ip": schema.StringAttribute{
				Description: "The private IPv6 address of the machine.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the machine was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the machine was last updated.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"guest": schema.SingleNestedBlock{
				Description: "Guest VM resource configuration.",
				Attributes: map[string]schema.Attribute{
					"cpu_kind": schema.StringAttribute{
						Description: "The kind of CPU to use (e.g., 'shared', 'performance').",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("shared"),
					},
					"cpus": schema.Int32Attribute{
						Description: "Number of vCPUs.",
						Optional:    true,
						Computed:    true,
					},
					"memory_mb": schema.Int32Attribute{
						Description: "Memory in megabytes.",
						Optional:    true,
						Computed:    true,
					},
					"gpu_kind": schema.StringAttribute{
						Description: "The kind of GPU to attach.",
						Optional:    true,
					},
					"gpus": schema.Int32Attribute{
						Description: "Number of GPUs.",
						Optional:    true,
					},
				},
			},
			"service": schema.ListNestedBlock{
				Description: "Services exposed by the machine.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"protocol": schema.StringAttribute{
							Description: "The protocol for the service (e.g., 'tcp', 'udp').",
							Required:    true,
						},
						"internal_port": schema.Int32Attribute{
							Description: "The port the service listens on inside the machine.",
							Required:    true,
						},
						"autostart": schema.BoolAttribute{
							Description: "Whether the machine should be started automatically when a request is received.",
							Optional:    true,
						},
						"autostop": schema.BoolAttribute{
							Description: "Whether the machine should be stopped automatically when idle.",
							Optional:    true,
						},
						"min_machines_running": schema.Int32Attribute{
							Description: "Minimum number of machines to keep running.",
							Optional:    true,
						},
						"force_https": schema.BoolAttribute{
							Description: "Whether to force HTTPS on the service.",
							Optional:    true,
							Computed:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"port": schema.ListNestedBlock{
							Description: "Port mappings for the service.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"port": schema.Int32Attribute{
										Description: "The external port number.",
										Optional:    true,
									},
									"handlers": schema.ListAttribute{
										Description: "Protocol handlers (e.g., 'http', 'tls').",
										Optional:    true,
										ElementType: types.StringType,
									},
									"force_https": schema.BoolAttribute{
										Description: "Whether to force HTTPS for this port.",
										Optional:    true,
										Computed:    true,
									},
								},
							},
						},
						"concurrency": schema.SingleNestedBlock{
							Description: "Concurrency limits for the service.",
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Description: "The concurrency type (e.g., 'connections', 'requests').",
									Optional:    true,
								},
								"hard_limit": schema.Int32Attribute{
									Description: "The hard concurrency limit.",
									Optional:    true,
								},
								"soft_limit": schema.Int32Attribute{
									Description: "The soft concurrency limit.",
									Optional:    true,
								},
							},
						},
					},
				},
			},
			"mount": schema.SingleNestedBlock{
				Description: "Volume mount configuration.",
				Attributes: map[string]schema.Attribute{
					"volume": schema.StringAttribute{
						Description: "The volume ID to mount. Changing this forces a new resource.",
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"path": schema.StringAttribute{
						Description: "The path inside the machine to mount the volume.",
						Optional:    true,
					},
					"extend_threshold_percent": schema.Int32Attribute{
						Description: "The threshold percentage at which to extend the volume.",
						Optional:    true,
					},
					"add_size_gb": schema.Int32Attribute{
						Description: "The size in GB to add when extending the volume.",
						Optional:    true,
					},
					"size_gb_limit": schema.Int32Attribute{
						Description: "The maximum size in GB the volume can be extended to.",
						Optional:    true,
					},
				},
			},
			"metrics": schema.SingleNestedBlock{
				Description: "Metrics endpoint configuration.",
				Attributes: map[string]schema.Attribute{
					"port": schema.Int32Attribute{
						Description: "The port to expose metrics on.",
						Optional:    true,
					},
					"path": schema.StringAttribute{
						Description: "The path to expose metrics at.",
						Optional:    true,
					},
				},
			},
			"restart": schema.SingleNestedBlock{
				Description: "Restart policy configuration.",
				Attributes: map[string]schema.Attribute{
					"policy": schema.StringAttribute{
						Description: "The restart policy (e.g., 'always', 'on-failure', 'no').",
						Optional:    true,
					},
					"max_retries": schema.Int32Attribute{
						Description: "Maximum number of restart retries.",
						Optional:    true,
					},
				},
			},
			"check": schema.ListNestedBlock{
				Description: "Health check configuration.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the health check.",
							Required:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of health check (e.g., 'tcp', 'http').",
							Required:    true,
						},
						"port": schema.Int32Attribute{
							Description: "The port to check.",
							Optional:    true,
						},
						"interval": schema.StringAttribute{
							Description: "How often to run the check (e.g., '10s', '1m').",
							Optional:    true,
						},
						"timeout": schema.StringAttribute{
							Description: "Timeout for the check (e.g., '2s').",
							Optional:    true,
						},
						"path": schema.StringAttribute{
							Description: "The HTTP path for HTTP checks.",
							Optional:    true,
						},
						"method": schema.StringAttribute{
							Description: "The HTTP method for HTTP checks.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (r *machineResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *machineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.MachineResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq, diags := machineModelToCreateRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	machine, err := r.client.CreateMachine(ctx, plan.App.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating machine", err.Error())
		return
	}

	// If not skipping launch, wait for the machine to be started.
	if !plan.SkipLaunch.ValueBool() {
		err = r.client.WaitForMachine(ctx, plan.App.ValueString(), machine.ID, "started", 60)
		if err != nil {
			resp.Diagnostics.AddError("Error waiting for machine to start", err.Error())
			return
		}

		// If desired_status is "stopped", stop the machine after creation.
		switch plan.DesiredStatus.ValueString() {
		case "stopped":
			err = r.client.StopMachine(ctx, plan.App.ValueString(), machine.ID)
			if err != nil {
				resp.Diagnostics.AddError("Error stopping machine after creation", err.Error())
				return
			}
			err = r.client.WaitForMachine(ctx, plan.App.ValueString(), machine.ID, "stopped", 60)
			if err != nil {
				resp.Diagnostics.AddError("Error waiting for machine to stop", err.Error())
				return
			}
		case "suspended":
			err = r.client.SuspendMachine(ctx, plan.App.ValueString(), machine.ID)
			if err != nil {
				resp.Diagnostics.AddError("Error suspending machine after creation", err.Error())
				return
			}
			err = r.client.WaitForMachine(ctx, plan.App.ValueString(), machine.ID, "suspended", 60)
			if err != nil {
				resp.Diagnostics.AddError("Error waiting for machine to suspend", err.Error())
				return
			}
		}
	}

	// Re-read the machine to get the final state.
	machine, err = r.client.GetMachine(ctx, plan.App.ValueString(), machine.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading machine after creation", err.Error())
		return
	}

	state, diags := machineToModel(ctx, machine, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *machineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.MachineResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	machine, err := r.client.GetMachine(ctx, state.App.ValueString(), state.ID.ValueString())
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading machine", err.Error())
		return
	}

	newState, diags := machineToModel(ctx, machine, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *machineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.MachineResourceModel
	var state models.MachineResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appName := plan.App.ValueString()
	machineID := state.ID.ValueString()

	// Acquire a lease on the machine before updating.
	lease, err := r.client.AcquireLease(ctx, appName, machineID, 60)
	if err != nil {
		resp.Diagnostics.AddError("Error acquiring lease on machine", err.Error())
		return
	}
	defer func() {
		_ = r.client.ReleaseLease(ctx, appName, machineID, lease.Nonce)
	}()

	// Build the update request from the plan.
	config, diags := machineModelToConfig(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := apimodels.UpdateMachineRequest{
		Config: config,
	}

	_, err = r.client.UpdateMachine(ctx, appName, machineID, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating machine", err.Error())
		return
	}

	// Handle desired_status changes.
	desiredStatus := plan.DesiredStatus.ValueString()
	previousDesiredStatus := state.DesiredStatus.ValueString()

	if desiredStatus != previousDesiredStatus {
		switch desiredStatus {
		case "started":
			err = r.client.StartMachine(ctx, appName, machineID)
			if err != nil {
				resp.Diagnostics.AddError("Error starting machine", err.Error())
				return
			}
		case "stopped":
			err = r.client.StopMachine(ctx, appName, machineID)
			if err != nil {
				resp.Diagnostics.AddError("Error stopping machine", err.Error())
				return
			}
		case "suspended":
			err = r.client.SuspendMachine(ctx, appName, machineID)
			if err != nil {
				resp.Diagnostics.AddError("Error suspending machine", err.Error())
				return
			}
		}
	}

	// Handle cordon/uncordon changes.
	cordoned := plan.Cordoned.ValueBool()
	previousCordoned := state.Cordoned.ValueBool()
	if cordoned != previousCordoned {
		if cordoned {
			err = r.client.CordonMachine(ctx, appName, machineID)
		} else {
			err = r.client.UncordonMachine(ctx, appName, machineID)
		}
		if err != nil {
			resp.Diagnostics.AddError("Error updating cordon status", err.Error())
			return
		}
	}

	// Wait for the machine to reach the desired state.
	waitState := desiredStatus
	if waitState == "" {
		waitState = "started"
	}
	err = r.client.WaitForMachine(ctx, appName, machineID, waitState, 60)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error waiting for machine to reach state '%s'", waitState),
			err.Error(),
		)
		return
	}

	// Re-read the machine to get the final state.
	machine, err := r.client.GetMachine(ctx, appName, machineID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading machine after update", err.Error())
		return
	}

	newState, diags := machineToModel(ctx, machine, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *machineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.MachineResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appName := state.App.ValueString()
	machineID := state.ID.ValueString()

	// Stop the machine if it is running.
	if state.State.ValueString() == "started" {
		err := r.client.StopMachine(ctx, appName, machineID)
		if err != nil && !apiclient.IsNotFound(err) {
			resp.Diagnostics.AddError("Error stopping machine before deletion", err.Error())
			return
		}
		if err == nil {
			_ = r.client.WaitForMachine(ctx, appName, machineID, "stopped", 60)
		}
	}

	err := r.client.DeleteMachine(ctx, appName, machineID)
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting machine", err.Error())
	}
}

func (r *machineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID in the format 'app_name/machine_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// --- Conversion helpers ---

// machineModelToCreateRequest converts a TF resource model into an API create request.
func machineModelToCreateRequest(ctx context.Context, m *models.MachineResourceModel) (apimodels.CreateMachineRequest, diag.Diagnostics) {
	config, diags := machineModelToConfig(ctx, m)

	req := apimodels.CreateMachineRequest{
		Name:       m.Name.ValueString(),
		Region:     m.Region.ValueString(),
		Config:     config,
		SkipLaunch: m.SkipLaunch.ValueBool(),
	}
	return req, diags
}

// machineModelToConfig converts a TF resource model into an API machine config.
func machineModelToConfig(ctx context.Context, m *models.MachineResourceModel) (apimodels.MachineConfig, diag.Diagnostics) {
	var allDiags diag.Diagnostics

	config := apimodels.MachineConfig{
		Image:       m.Image.ValueString(),
		AutoDestroy: m.AutoDestroy.ValueBool(),
		Schedule:    m.Schedule.ValueString(),
	}

	// Env
	if !m.Env.IsNull() && !m.Env.IsUnknown() {
		envMap := make(map[string]string)
		diags := m.Env.ElementsAs(ctx, &envMap, false)
		allDiags.Append(diags...)
		config.Env = envMap
	}

	// Metadata
	if !m.Metadata.IsNull() && !m.Metadata.IsUnknown() {
		metaMap := make(map[string]string)
		diags := m.Metadata.ElementsAs(ctx, &metaMap, false)
		allDiags.Append(diags...)
		config.Metadata = metaMap
	}

	// Init (cmd, entrypoint)
	var hasInit bool
	var init apimodels.MachineInit

	if !m.Cmd.IsNull() && !m.Cmd.IsUnknown() {
		var cmd []string
		diags := m.Cmd.ElementsAs(ctx, &cmd, false)
		allDiags.Append(diags...)
		init.Cmd = cmd
		hasInit = true
	}

	if !m.Entrypoint.IsNull() && !m.Entrypoint.IsUnknown() {
		var entrypoint []string
		diags := m.Entrypoint.ElementsAs(ctx, &entrypoint, false)
		allDiags.Append(diags...)
		init.Entrypoint = entrypoint
		hasInit = true
	}

	if hasInit {
		config.Init = &init
	}

	// Guest
	if m.Guest != nil {
		guest := &apimodels.MachineGuest{
			CPUKind:  m.Guest.CPUKind.ValueString(),
			CPUs:     int(m.Guest.CPUs.ValueInt32()),
			MemoryMB: int(m.Guest.MemoryMB.ValueInt32()),
		}
		if !m.Guest.GPUKind.IsNull() && !m.Guest.GPUKind.IsUnknown() {
			guest.GPUKind = m.Guest.GPUKind.ValueString()
		}
		if !m.Guest.GPUs.IsNull() && !m.Guest.GPUs.IsUnknown() {
			guest.GPUs = int(m.Guest.GPUs.ValueInt32())
		}
		config.Guest = guest
	}

	// Services
	if len(m.Services) > 0 {
		services := make([]apimodels.MachineService, len(m.Services))
		for i, svc := range m.Services {
			s := apimodels.MachineService{
				Protocol:     svc.Protocol.ValueString(),
				InternalPort: int(svc.InternalPort.ValueInt32()),
				ForceHTTPS:   svc.ForceHTTPS.ValueBool(),
			}

			if !svc.Autostart.IsNull() && !svc.Autostart.IsUnknown() {
				v := svc.Autostart.ValueBool()
				s.Autostart = &v
			}
			if !svc.Autostop.IsNull() && !svc.Autostop.IsUnknown() {
				v := svc.Autostop.ValueBool()
				s.Autostop = &v
			}
			if !svc.MinMachinesRunning.IsNull() && !svc.MinMachinesRunning.IsUnknown() {
				v := int(svc.MinMachinesRunning.ValueInt32())
				s.MinMachinesRunning = &v
			}

			// Ports
			if len(svc.Ports) > 0 {
				ports := make([]apimodels.MachinePort, len(svc.Ports))
				for j, p := range svc.Ports {
					mp := apimodels.MachinePort{
						ForceHTTPS: p.ForceHTTPS.ValueBool(),
					}
					if !p.Port.IsNull() && !p.Port.IsUnknown() {
						v := int(p.Port.ValueInt32())
						mp.Port = &v
					}
					if !p.Handlers.IsNull() && !p.Handlers.IsUnknown() {
						var handlers []string
						diags := p.Handlers.ElementsAs(ctx, &handlers, false)
						allDiags.Append(diags...)
						mp.Handlers = handlers
					}
					ports[j] = mp
				}
				s.Ports = ports
			}

			// Concurrency
			if svc.Concurrency != nil {
				s.Concurrency = &apimodels.ServiceConcurrency{
					Type:      svc.Concurrency.Type.ValueString(),
					HardLimit: int(svc.Concurrency.HardLimit.ValueInt32()),
					SoftLimit: int(svc.Concurrency.SoftLimit.ValueInt32()),
				}
			}

			services[i] = s
		}
		config.Services = services
	}

	// Mount
	if m.Mount != nil {
		mount := apimodels.MachineMount{
			Volume: m.Mount.Volume.ValueString(),
			Path:   m.Mount.Path.ValueString(),
		}
		if !m.Mount.ExtendThresholdPercent.IsNull() && !m.Mount.ExtendThresholdPercent.IsUnknown() {
			mount.ExtendThresholdPercent = int(m.Mount.ExtendThresholdPercent.ValueInt32())
		}
		if !m.Mount.AddSizeGB.IsNull() && !m.Mount.AddSizeGB.IsUnknown() {
			mount.AddSizeGB = int(m.Mount.AddSizeGB.ValueInt32())
		}
		if !m.Mount.SizeGBLimit.IsNull() && !m.Mount.SizeGBLimit.IsUnknown() {
			mount.SizeGBLimit = int(m.Mount.SizeGBLimit.ValueInt32())
		}
		config.Mounts = []apimodels.MachineMount{mount}
	}

	// Metrics
	if m.Metrics != nil {
		config.Metrics = &apimodels.MachineMetrics{
			Port: int(m.Metrics.Port.ValueInt32()),
			Path: m.Metrics.Path.ValueString(),
		}
	}

	// Restart
	if m.Restart != nil {
		config.Restart = &apimodels.MachineRestart{
			Policy:     m.Restart.Policy.ValueString(),
			MaxRetries: int(m.Restart.MaxRetries.ValueInt32()),
		}
	}

	// Checks
	if len(m.Checks) > 0 {
		checks := make(map[string]apimodels.Check)
		for _, c := range m.Checks {
			check := apimodels.Check{
				Type:     c.Type.ValueString(),
				Interval: c.Interval.ValueString(),
				Timeout:  c.Timeout.ValueString(),
				Path:     c.Path.ValueString(),
				Method:   c.Method.ValueString(),
			}
			if !c.Port.IsNull() && !c.Port.IsUnknown() {
				v := int(c.Port.ValueInt32())
				check.Port = &v
			}
			checks[c.Name.ValueString()] = check
		}
		config.Checks = checks
	}

	return config, allDiags
}

// machineToModel converts an API machine response to a TF resource model,
// preserving user-configured values from the existing model.
func machineToModel(ctx context.Context, machine *apimodels.Machine, existing *models.MachineResourceModel) (models.MachineResourceModel, diag.Diagnostics) {
	var allDiags diag.Diagnostics

	m := models.MachineResourceModel{
		ID:        types.StringValue(machine.ID),
		App:       existing.App,
		Name:      types.StringValue(machine.Name),
		Region:    types.StringValue(machine.Region),
		Image:     types.StringValue(machine.Config.Image),
		Schedule:  existing.Schedule,
		AutoDestroy: types.BoolValue(machine.Config.AutoDestroy),

		// Preserve user-configured values.
		DesiredStatus: existing.DesiredStatus,
		Cordoned:      existing.Cordoned,
		SkipLaunch:    existing.SkipLaunch,
		Cmd:           existing.Cmd,
		Entrypoint:    existing.Entrypoint,

		// Computed values from API.
		InstanceID: types.StringValue(machine.InstanceID),
		State:      types.StringValue(machine.State),
		PrivateIP:  types.StringValue(machine.PrivateIP),
		CreatedAt:  types.StringValue(machine.CreatedAt),
		UpdatedAt:  types.StringValue(machine.UpdatedAt),
	}

	// Env: update from API response.
	if len(machine.Config.Env) > 0 {
		envMap, diags := types.MapValueFrom(ctx, types.StringType, machine.Config.Env)
		allDiags.Append(diags...)
		m.Env = envMap
	} else if existing.Env.IsNull() {
		m.Env = types.MapNull(types.StringType)
	} else {
		m.Env = existing.Env
	}

	// Metadata: update from API response.
	if len(machine.Config.Metadata) > 0 {
		metaMap, diags := types.MapValueFrom(ctx, types.StringType, machine.Config.Metadata)
		allDiags.Append(diags...)
		m.Metadata = metaMap
	} else if existing.Metadata.IsNull() {
		m.Metadata = types.MapNull(types.StringType)
	} else {
		m.Metadata = existing.Metadata
	}

	// Schedule: if the API returns a schedule, use it.
	if machine.Config.Schedule != "" {
		m.Schedule = types.StringValue(machine.Config.Schedule)
	} else if existing.Schedule.IsNull() || existing.Schedule.IsUnknown() {
		m.Schedule = types.StringNull()
	}

	// Guest
	if machine.Config.Guest != nil {
		g := machine.Config.Guest
		m.Guest = &models.GuestModel{
			CPUKind:  types.StringValue(g.CPUKind),
			CPUs:     types.Int32Value(int32(g.CPUs)),
			MemoryMB: types.Int32Value(int32(g.MemoryMB)),
		}
		if g.GPUKind != "" {
			m.Guest.GPUKind = types.StringValue(g.GPUKind)
		} else {
			m.Guest.GPUKind = types.StringNull()
		}
		if g.GPUs > 0 {
			m.Guest.GPUs = types.Int32Value(int32(g.GPUs))
		} else {
			m.Guest.GPUs = types.Int32Null()
		}
	} else {
		m.Guest = existing.Guest
	}

	// Services
	if len(machine.Config.Services) > 0 {
		services := make([]models.ServiceModel, len(machine.Config.Services))
		for i, svc := range machine.Config.Services {
			s := models.ServiceModel{
				Protocol:     types.StringValue(svc.Protocol),
				InternalPort: types.Int32Value(int32(svc.InternalPort)),
				ForceHTTPS:   types.BoolValue(svc.ForceHTTPS),
			}

			if svc.Autostart != nil {
				s.Autostart = types.BoolValue(*svc.Autostart)
			} else {
				s.Autostart = types.BoolNull()
			}
			if svc.Autostop != nil {
				s.Autostop = types.BoolValue(*svc.Autostop)
			} else {
				s.Autostop = types.BoolNull()
			}
			if svc.MinMachinesRunning != nil {
				s.MinMachinesRunning = types.Int32Value(int32(*svc.MinMachinesRunning))
			} else {
				s.MinMachinesRunning = types.Int32Null()
			}

			// Ports
			if len(svc.Ports) > 0 {
				ports := make([]models.PortModel, len(svc.Ports))
				for j, p := range svc.Ports {
					pm := models.PortModel{
						ForceHTTPS: types.BoolValue(p.ForceHTTPS),
					}
					if p.Port != nil {
						pm.Port = types.Int32Value(int32(*p.Port))
					} else {
						pm.Port = types.Int32Null()
					}
					if len(p.Handlers) > 0 {
						handlers, diags := types.ListValueFrom(ctx, types.StringType, p.Handlers)
						allDiags.Append(diags...)
						pm.Handlers = handlers
					} else {
						pm.Handlers = types.ListNull(types.StringType)
					}
					ports[j] = pm
				}
				s.Ports = ports
			}

			// Concurrency
			if svc.Concurrency != nil {
				s.Concurrency = &models.ConcurrencyModel{
					Type:      types.StringValue(svc.Concurrency.Type),
					HardLimit: types.Int32Value(int32(svc.Concurrency.HardLimit)),
					SoftLimit: types.Int32Value(int32(svc.Concurrency.SoftLimit)),
				}
			}

			services[i] = s
		}
		m.Services = services
	} else {
		m.Services = existing.Services
	}

	// Mount
	if len(machine.Config.Mounts) > 0 {
		mount := machine.Config.Mounts[0]
		m.Mount = &models.MountModel{
			Volume: types.StringValue(mount.Volume),
			Path:   types.StringValue(mount.Path),
		}
		if mount.ExtendThresholdPercent > 0 {
			m.Mount.ExtendThresholdPercent = types.Int32Value(int32(mount.ExtendThresholdPercent))
		} else {
			m.Mount.ExtendThresholdPercent = types.Int32Null()
		}
		if mount.AddSizeGB > 0 {
			m.Mount.AddSizeGB = types.Int32Value(int32(mount.AddSizeGB))
		} else {
			m.Mount.AddSizeGB = types.Int32Null()
		}
		if mount.SizeGBLimit > 0 {
			m.Mount.SizeGBLimit = types.Int32Value(int32(mount.SizeGBLimit))
		} else {
			m.Mount.SizeGBLimit = types.Int32Null()
		}
	} else {
		m.Mount = existing.Mount
	}

	// Metrics
	if machine.Config.Metrics != nil {
		m.Metrics = &models.MetricsModel{
			Port: types.Int32Value(int32(machine.Config.Metrics.Port)),
			Path: types.StringValue(machine.Config.Metrics.Path),
		}
	} else {
		m.Metrics = existing.Metrics
	}

	// Restart
	if machine.Config.Restart != nil {
		m.Restart = &models.RestartModel{
			Policy:     types.StringValue(machine.Config.Restart.Policy),
			MaxRetries: types.Int32Value(int32(machine.Config.Restart.MaxRetries)),
		}
	} else {
		m.Restart = existing.Restart
	}

	// Checks
	if len(machine.Config.Checks) > 0 {
		checks := make([]models.MachineCheckModel, 0, len(machine.Config.Checks))
		for name, c := range machine.Config.Checks {
			check := models.MachineCheckModel{
				Name:     types.StringValue(name),
				Type:     types.StringValue(c.Type),
				Interval: types.StringValue(c.Interval),
				Timeout:  types.StringValue(c.Timeout),
			}
			if c.Port != nil {
				check.Port = types.Int32Value(int32(*c.Port))
			} else {
				check.Port = types.Int32Null()
			}
			if c.Path != "" {
				check.Path = types.StringValue(c.Path)
			} else {
				check.Path = types.StringNull()
			}
			if c.Method != "" {
				check.Method = types.StringValue(c.Method)
			} else {
				check.Method = types.StringNull()
			}
			checks = append(checks, check)
		}
		m.Checks = checks
	} else {
		m.Checks = existing.Checks
	}

	return m, allDiags
}
