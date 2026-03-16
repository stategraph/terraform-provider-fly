package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stategraph/terraform-provider-fly/internal/datasources"
	"github.com/stategraph/terraform-provider-fly/internal/models"
	"github.com/stategraph/terraform-provider-fly/internal/resources"
	"github.com/stategraph/terraform-provider-fly/pkg/apiclient"
	"github.com/stategraph/terraform-provider-fly/pkg/flyctl"
)

var _ provider.Provider = &FlyProvider{}

type FlyProvider struct {
	version string
}

type FlyProviderModel struct {
	APIToken   types.String `tfsdk:"api_token"`
	APIURL     types.String `tfsdk:"api_url"`
	OrgSlug    types.String `tfsdk:"org_slug"`
	FlyctlPath types.String `tfsdk:"flyctl_path"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &FlyProvider{version: version}
	}
}

func (p *FlyProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "fly"
	resp.Version = p.version
}

func (p *FlyProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing Fly.io infrastructure.",
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				Description: "Fly.io API token. Can also be set via FLY_API_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_url": schema.StringAttribute{
				Description: "Fly.io Machines API base URL. Defaults to https://api.machines.dev/v1. Can also be set via FLY_API_URL.",
				Optional:    true,
			},
			"org_slug": schema.StringAttribute{
				Description: "Default Fly.io organization slug. Can also be set via FLY_ORG.",
				Optional:    true,
			},
			"flyctl_path": schema.StringAttribute{
				Description: "Path to the flyctl binary. Can also be set via FLYCTL_PATH. If unset, searches PATH for flyctl or fly.",
				Optional:    true,
			},
		},
	}
}

func (p *FlyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config FlyProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token := config.APIToken.ValueString()
	if token == "" {
		token = os.Getenv("FLY_API_TOKEN")
	}
	if token == "" {
		resp.Diagnostics.AddError(
			"Missing API Token",
			"The Fly.io API token must be set in the provider configuration block or via the FLY_API_TOKEN environment variable.",
		)
		return
	}

	apiURL := config.APIURL.ValueString()
	if apiURL == "" {
		apiURL = os.Getenv("FLY_API_URL")
	}

	var opts []apiclient.ClientOption
	if apiURL != "" {
		opts = append(opts, apiclient.WithBaseURL(apiURL))
	}

	client := apiclient.NewClient(token, p.version, opts...)

	// Resolve flyctl binary path.
	flyctlPath := config.FlyctlPath.ValueString()
	if flyctlPath == "" {
		flyctlPath = os.Getenv("FLYCTL_PATH")
	}
	binaryPath, err := flyctl.FindBinary(flyctlPath)
	if err != nil {
		// flyctl is optional — only warn, don't fail.
		// Resources that need it will fail at runtime with a clear error.
		resp.Diagnostics.AddWarning(
			"flyctl not found",
			"The flyctl binary was not found. Resources that require flyctl (managed Postgres, Redis, extensions, etc.) will not work. "+
				"Set flyctl_path in the provider configuration or install flyctl in your PATH.",
		)
		binaryPath = ""
	}

	var executor *flyctl.Executor
	if binaryPath != "" {
		executor = flyctl.NewExecutor(binaryPath, token)
	}

	pd := &models.ProviderData{
		APIClient: client,
		Flyctl:    executor,
	}

	resp.DataSourceData = pd
	resp.ResourceData = pd
}

func (p *FlyProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Layer 1: REST API resources
		resources.NewAppResource,
		resources.NewMachineResource,
		resources.NewVolumeResource,
		resources.NewCertificateResource,
		resources.NewSecretResource,
		resources.NewNetworkPolicyResource,
		resources.NewVolumeSnapshotResource,

		// Layer 2: flyctl-based resources (migrated from GraphQL)
		resources.NewIPAddressResource,
		resources.NewEgressIPResource,
		resources.NewWireGuardPeerResource,
		resources.NewWireGuardTokenResource,

		// Layer 2: flyctl-based resources (new)
		resources.NewMPGClusterResource,
		resources.NewMPGDatabaseResource,
		resources.NewMPGUserResource,
		resources.NewMPGAttachmentResource,
		resources.NewPostgresClusterResource,
		resources.NewPostgresAttachmentResource,
		resources.NewRedisResource,
		resources.NewTigrisBucketResource,
		resources.NewTokenResource,
		resources.NewOrgResource,
		resources.NewOrgMemberResource,
		resources.NewLiteFSClusterResource,

		// Extensions
		resources.NewExtMySQLResource,
		resources.NewExtKubernetesResource,
		resources.NewExtSentryResource,
		resources.NewExtArcjetResource,
		resources.NewExtWafrisResource,
		resources.NewExtVectorResource,
	}
}

func (p *FlyProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Layer 1: REST API data sources
		datasources.NewAppDataSource,
		datasources.NewMachineDataSource,
		datasources.NewVolumeDataSource,
		datasources.NewCertificateDataSource,
		datasources.NewAppsDataSource,
		datasources.NewMachinesDataSource,
		datasources.NewVolumesDataSource,
		datasources.NewCertificatesDataSource,
		datasources.NewVolumeSnapshotsDataSource,
		datasources.NewNetworkPoliciesDataSource,
		datasources.NewOIDCTokenDataSource,

		// Layer 2: flyctl-based data sources
		datasources.NewIPAddressesDataSource,
		datasources.NewOrganizationDataSource,
		datasources.NewRegionsDataSource,
		datasources.NewMPGClustersDataSource,
		datasources.NewRedisInstancesDataSource,
		datasources.NewTigrisBucketsDataSource,
		datasources.NewTokensDataSource,
	}
}
