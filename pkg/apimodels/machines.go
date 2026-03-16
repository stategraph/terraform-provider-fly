package apimodels

type Machine struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	State      string        `json:"state"`
	Region     string        `json:"region"`
	InstanceID string        `json:"instance_id"`
	PrivateIP  string        `json:"private_ip"`
	Config     MachineConfig `json:"config"`
	CreatedAt  string        `json:"created_at"`
	UpdatedAt  string        `json:"updated_at"`
}

type MachineConfig struct {
	Image       string            `json:"image"`
	Env         map[string]string `json:"env,omitempty"`
	Init        *MachineInit      `json:"init,omitempty"`
	Guest       *MachineGuest     `json:"guest,omitempty"`
	Services    []MachineService  `json:"services,omitempty"`
	Mounts      []MachineMount    `json:"mounts,omitempty"`
	Checks      map[string]Check  `json:"checks,omitempty"`
	Metrics     *MachineMetrics   `json:"metrics,omitempty"`
	Restart     *MachineRestart   `json:"restart,omitempty"`
	DNS         *MachineDNS       `json:"dns,omitempty"`
	Files       []MachineFile     `json:"files,omitempty"`
	Statics     []MachineStatic   `json:"statics,omitempty"`
	StopConfig  *StopConfig       `json:"stop_config,omitempty"`
	AutoDestroy bool              `json:"auto_destroy,omitempty"`
	Schedule    string            `json:"schedule,omitempty"`
	Standbys    []string          `json:"standbys,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type MachineGuest struct {
	CPUKind  string `json:"cpu_kind,omitempty"`
	CPUs     int    `json:"cpus,omitempty"`
	MemoryMB int    `json:"memory_mb,omitempty"`
	GPUKind  string `json:"gpu_kind,omitempty"`
	GPUs     int    `json:"gpus,omitempty"`
}

type MachineInit struct {
	Cmd        []string `json:"cmd,omitempty"`
	Entrypoint []string `json:"entrypoint,omitempty"`
	Exec       []string `json:"exec,omitempty"`
	SwapSizeMB int      `json:"swap_size_mb,omitempty"`
	TTY        bool     `json:"tty,omitempty"`
}

type MachineService struct {
	Protocol            string             `json:"protocol"`
	InternalPort        int                `json:"internal_port"`
	Autostart           *bool              `json:"autostart,omitempty"`
	Autostop            *bool              `json:"autostop,omitempty"`
	MinMachinesRunning  *int               `json:"min_machines_running,omitempty"`
	ForceHTTPS          bool               `json:"force_instance_https,omitempty"`
	Ports               []MachinePort      `json:"ports,omitempty"`
	Concurrency         *ServiceConcurrency `json:"concurrency,omitempty"`
	Checks              []ServiceCheck     `json:"checks,omitempty"`
}

type MachinePort struct {
	Port       *int     `json:"port,omitempty"`
	Handlers   []string `json:"handlers,omitempty"`
	ForceHTTPS bool     `json:"force_https,omitempty"`
}

type ServiceConcurrency struct {
	Type      string `json:"type"`
	HardLimit int    `json:"hard_limit"`
	SoftLimit int    `json:"soft_limit"`
}

type ServiceCheck struct {
	Type     string            `json:"type,omitempty"`
	Port     *int              `json:"port,omitempty"`
	Interval string            `json:"interval,omitempty"`
	Timeout  string            `json:"timeout,omitempty"`
	Path     string            `json:"path,omitempty"`
	Method   string            `json:"method,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
}

type MachineMount struct {
	Volume                 string `json:"volume"`
	Path                   string `json:"path"`
	ExtendThresholdPercent int    `json:"extend_threshold_percent,omitempty"`
	AddSizeGB              int    `json:"add_size_gb,omitempty"`
	SizeGBLimit            int    `json:"size_gb_limit,omitempty"`
}

type Check struct {
	Type     string            `json:"type,omitempty"`
	Port     *int              `json:"port,omitempty"`
	Interval string            `json:"interval,omitempty"`
	Timeout  string            `json:"timeout,omitempty"`
	Path     string            `json:"path,omitempty"`
	Method   string            `json:"method,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
}

type MachineMetrics struct {
	Port int    `json:"port"`
	Path string `json:"path"`
}

type MachineRestart struct {
	Policy     string `json:"policy,omitempty"`
	MaxRetries int    `json:"max_retries,omitempty"`
}

type MachineDNS struct {
	SkipRegistration bool     `json:"skip_registration,omitempty"`
	Nameservers      []string `json:"nameservers,omitempty"`
	Searches         []string `json:"searches,omitempty"`
}

type MachineFile struct {
	GuestPath  string `json:"guest_path"`
	RawValue   string `json:"raw_value,omitempty"`
	SecretName string `json:"secret_name,omitempty"`
}

type MachineStatic struct {
	GuestPath   string `json:"guest_path"`
	URLPrefix   string `json:"url_prefix"`
	TigrisBucket string `json:"tigris_bucket,omitempty"`
}

type StopConfig struct {
	Timeout string `json:"timeout,omitempty"`
	Signal  string `json:"signal,omitempty"`
}

type CreateMachineRequest struct {
	Name       string        `json:"name,omitempty"`
	Region     string        `json:"region"`
	Config     MachineConfig `json:"config"`
	SkipLaunch bool          `json:"skip_launch,omitempty"`
}

type UpdateMachineRequest struct {
	Config MachineConfig `json:"config"`
}

type Lease struct {
	Nonce     string `json:"nonce"`
	ExpiresAt int64  `json:"expires_at"`
	Owner     string `json:"owner"`
}

type LeaseRequest struct {
	TTL int `json:"ttl,omitempty"`
}
