package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type MachineResourceModel struct {
	ID            types.String `tfsdk:"id"`
	App           types.String `tfsdk:"app"`
	Name          types.String `tfsdk:"name"`
	Region        types.String `tfsdk:"region"`
	Image         types.String `tfsdk:"image"`
	Env           types.Map    `tfsdk:"env"`
	AutoDestroy   types.Bool   `tfsdk:"auto_destroy"`
	Metadata      types.Map    `tfsdk:"metadata"`
	DesiredStatus types.String `tfsdk:"desired_status"`
	Cordoned      types.Bool   `tfsdk:"cordoned"`
	Cmd           types.List   `tfsdk:"cmd"`
	Entrypoint    types.List   `tfsdk:"entrypoint"`
	SkipLaunch    types.Bool   `tfsdk:"skip_launch"`
	Schedule      types.String `tfsdk:"schedule"`

	// Nested blocks
	Guest    *GuestModel     `tfsdk:"guest"`
	Services []ServiceModel  `tfsdk:"service"`
	Mount    *MountModel     `tfsdk:"mount"`
	Metrics  *MetricsModel   `tfsdk:"metrics"`
	Restart  *RestartModel   `tfsdk:"restart"`
	Checks   []MachineCheckModel `tfsdk:"check"`

	// Computed
	InstanceID types.String `tfsdk:"instance_id"`
	State      types.String `tfsdk:"state"`
	PrivateIP  types.String `tfsdk:"private_ip"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

type GuestModel struct {
	CPUKind  types.String `tfsdk:"cpu_kind"`
	CPUs     types.Int32  `tfsdk:"cpus"`
	MemoryMB types.Int32  `tfsdk:"memory_mb"`
	GPUKind  types.String `tfsdk:"gpu_kind"`
	GPUs     types.Int32  `tfsdk:"gpus"`
}

type ServiceModel struct {
	Protocol           types.String       `tfsdk:"protocol"`
	InternalPort       types.Int32        `tfsdk:"internal_port"`
	Autostart          types.Bool         `tfsdk:"autostart"`
	Autostop           types.Bool         `tfsdk:"autostop"`
	MinMachinesRunning types.Int32        `tfsdk:"min_machines_running"`
	ForceHTTPS         types.Bool         `tfsdk:"force_https"`
	Ports              []PortModel        `tfsdk:"port"`
	Concurrency        *ConcurrencyModel  `tfsdk:"concurrency"`
}

type PortModel struct {
	Port       types.Int32 `tfsdk:"port"`
	Handlers   types.List  `tfsdk:"handlers"`
	ForceHTTPS types.Bool  `tfsdk:"force_https"`
}

type ConcurrencyModel struct {
	Type      types.String `tfsdk:"type"`
	HardLimit types.Int32  `tfsdk:"hard_limit"`
	SoftLimit types.Int32  `tfsdk:"soft_limit"`
}

type MountModel struct {
	Volume                 types.String `tfsdk:"volume"`
	Path                   types.String `tfsdk:"path"`
	ExtendThresholdPercent types.Int32  `tfsdk:"extend_threshold_percent"`
	AddSizeGB              types.Int32  `tfsdk:"add_size_gb"`
	SizeGBLimit            types.Int32  `tfsdk:"size_gb_limit"`
}

type MetricsModel struct {
	Port types.Int32  `tfsdk:"port"`
	Path types.String `tfsdk:"path"`
}

type RestartModel struct {
	Policy     types.String `tfsdk:"policy"`
	MaxRetries types.Int32  `tfsdk:"max_retries"`
}

type MachineCheckModel struct {
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Port     types.Int32  `tfsdk:"port"`
	Interval types.String `tfsdk:"interval"`
	Timeout  types.String `tfsdk:"timeout"`
	Path     types.String `tfsdk:"path"`
	Method   types.String `tfsdk:"method"`
}

type MachineDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	App        types.String `tfsdk:"app"`
	Name       types.String `tfsdk:"name"`
	Region     types.String `tfsdk:"region"`
	Image      types.String `tfsdk:"image"`
	State      types.String `tfsdk:"state"`
	PrivateIP  types.String `tfsdk:"private_ip"`
	InstanceID types.String `tfsdk:"instance_id"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}
