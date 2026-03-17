package models

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stategraph/terraform-provider-fly/pkg/apiclient"
	"github.com/stategraph/terraform-provider-fly/pkg/flyctl"
)

// ProviderData holds the configured clients passed to resources and data sources.
type ProviderData struct {
	APIClient *apiclient.Client
	Flyctl    *flyctl.Executor
	DryRun    bool
}

// FlushDryRunWarnings drains accumulated dry-run messages from the API client
// and flyctl executor, adding each as a Terraform warning diagnostic.
func FlushDryRunWarnings(diags *diag.Diagnostics, client *apiclient.Client, flyctl *flyctl.Executor) {
	if client != nil {
		for i, msg := range client.FlushDryRunMessages() {
			diags.AddWarning(fmt.Sprintf("Dry Run [%d]", i+1), msg)
		}
	}
	if flyctl != nil {
		for i, msg := range flyctl.FlushDryRunMessages() {
			diags.AddWarning(fmt.Sprintf("Dry Run [%d]", i+1), msg)
		}
	}
}
