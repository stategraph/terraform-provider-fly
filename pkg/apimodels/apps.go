package apimodels

type App struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Organization AppOrg       `json:"organization,omitempty"`
	OrgSlug      string       `json:"-"` // Populated from Organization.Slug
	Network      string       `json:"network,omitempty"`
	Status       string       `json:"status,omitempty"`
	Hostname     string       `json:"hostname,omitempty"`
}

type AppOrg struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type CreateAppRequest struct {
	AppName string `json:"app_name"`
	OrgSlug string `json:"org_slug"`
	Network string `json:"network,omitempty"`
}
