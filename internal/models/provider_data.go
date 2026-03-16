package models

import (
	"github.com/stategraph/terraform-provider-fly/pkg/apiclient"
	"github.com/stategraph/terraform-provider-fly/pkg/flyctl"
)

// ProviderData holds the configured clients passed to resources and data sources.
type ProviderData struct {
	APIClient *apiclient.Client
	Flyctl    *flyctl.Executor
}
