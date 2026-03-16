package apimodels

type Volume struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	App               string `json:"app,omitempty"`
	Region            string `json:"region"`
	SizeGB            int    `json:"size_gb"`
	State             string `json:"state"`
	Zone              string `json:"zone,omitempty"`
	Encrypted         bool   `json:"encrypted"`
	AttachedMachineID string `json:"attached_machine_id,omitempty"`
	CreatedAt         string `json:"created_at"`
	SnapshotRetention int    `json:"snapshot_retention,omitempty"`
	AutoBackupEnabled bool   `json:"auto_backup_enabled,omitempty"`
}

type CreateVolumeRequest struct {
	Name              string `json:"name"`
	Region            string `json:"region"`
	SizeGB            int    `json:"size_gb"`
	Encrypted         *bool  `json:"encrypted,omitempty"`
	SnapshotID        string `json:"snapshot_id,omitempty"`
	SourceVolumeID    string `json:"source_volume_id,omitempty"`
	SnapshotRetention int    `json:"snapshot_retention,omitempty"`
	RequireUniqueZone *bool  `json:"require_unique_zone,omitempty"`
}

type ExtendVolumeRequest struct {
	SizeGB int `json:"size_gb"`
}

type ExtendVolumeResponse struct {
	Volume       Volume `json:"volume"`
	NeedsRestart bool   `json:"needs_restart"`
}

type UpdateVolumeRequest struct {
	AutoBackupEnabled *bool `json:"auto_backup_enabled,omitempty"`
	SnapshotRetention *int  `json:"snapshot_retention,omitempty"`
}
