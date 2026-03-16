package apimodels

type VolumeSnapshot struct {
	ID        string `json:"id"`
	Size      int    `json:"size"`
	Digest    string `json:"digest,omitempty"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}
