package apimodels

type Certificate struct {
	ID                      string `json:"id"`
	Hostname                string `json:"hostname"`
	CheckStatus             string `json:"check,omitempty"`
	DNSValidationHostname   string `json:"dns_validation_hostname,omitempty"`
	DNSValidationTarget     string `json:"dns_validation_target,omitempty"`
	DNSValidationInstructions string `json:"dns_validation_instructions,omitempty"`
	Source                  string `json:"source,omitempty"`
	IssuedAt                string `json:"issued_at,omitempty"`
	CertificateAuthority    string `json:"certificate_authority,omitempty"`
	CreatedAt               string `json:"created_at,omitempty"`
}

type AddCertificateRequest struct {
	Hostname string `json:"hostname"`
}

type ACMECertificateRequest struct {
	Hostname string `json:"hostname"`
}

type CustomCertificateRequest struct {
	Hostname   string `json:"hostname"`
	Fullchain  string `json:"fullchain"`
	PrivateKey string `json:"private_key"`
}
