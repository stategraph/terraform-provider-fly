package apimodels

type Secret struct {
	Name      string `json:"name"`
	Label     string `json:"label,omitempty"` // Alias for name in some contexts
	Digest    string `json:"digest,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	Type      string `json:"type,omitempty"`
}

// EffectiveName returns the name, falling back to label for compatibility.
func (s Secret) EffectiveName() string {
	if s.Name != "" {
		return s.Name
	}
	return s.Label
}

type SetSecretRequest struct {
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

// Deprecated: batch endpoint doesn't work reliably.
type SetSecretsRequest struct {
	Secrets []struct {
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"secrets"`
}
