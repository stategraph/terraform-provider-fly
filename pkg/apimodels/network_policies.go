package apimodels

type NetworkPolicy struct {
	ID       string         `json:"id,omitempty"`
	Name     string         `json:"name"`
	Selector PolicySelector `json:"selector"`
	Rules    []PolicyRule   `json:"rules"`
}

type PolicySelector struct {
	All      bool                    `json:"all,omitempty"`
	Machines []PolicyMachineSelector `json:"machines,omitempty"`
	Metadata map[string]string       `json:"metadata,omitempty"`
}

type PolicyMachineSelector struct {
	ID string `json:"id"`
}

type PolicyRule struct {
	Action    string       `json:"action"`
	Direction string       `json:"direction"`
	Ports     []PolicyPort `json:"ports"`
}

type PolicyPort struct {
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
}

type CreateNetworkPolicyRequest struct {
	ID       string         `json:"id,omitempty"`
	Name     string         `json:"name"`
	Selector PolicySelector `json:"selector"`
	Rules    []PolicyRule   `json:"rules"`
}
