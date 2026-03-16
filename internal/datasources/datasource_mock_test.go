package datasources_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

func testProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"fly": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

// flyctlMockResponse defines a canned response for a flyctl command pattern.
type flyctlMockResponse struct {
	Stdout   string
	ExitCode int
}

// createMockFlyctl creates a temporary shell script that acts as a mock flyctl binary.
func createMockFlyctl(t *testing.T, responses map[string]flyctlMockResponse) string {
	t.Helper()

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "flyctl")

	var cases strings.Builder
	for pattern, resp := range responses {
		stdout, _ := json.Marshal(resp.Stdout)
		cases.WriteString(fmt.Sprintf("  *%q*)\n    printf '%%s' %s\n    exit %d\n    ;;\n",
			pattern, string(stdout), resp.ExitCode))
	}

	script := fmt.Sprintf(`#!/bin/bash
ARGS="$*"
case "$ARGS" in
%s  *)
    echo "mock flyctl: unhandled command: $ARGS" >&2
    exit 1
    ;;
esac
`, cases.String())

	err := os.WriteFile(scriptPath, []byte(script), 0755)
	if err != nil {
		t.Fatalf("failed to write mock flyctl: %v", err)
	}

	return scriptPath
}

func testDSProviderConfig(apiURL string) string {
	return fmt.Sprintf(`
provider "fly" {
  api_token = "mock-token"
  api_url   = %q
}
`, apiURL)
}

func testDSProviderConfigWithFlyctl(apiURL, flyctlPath string) string {
	return fmt.Sprintf(`
provider "fly" {
  api_token   = "mock-token"
  api_url     = %q
  flyctl_path = %q
}
`, apiURL, flyctlPath)
}

func TestAppDataSource_read(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/apps/my-ds-app" {
			_ = json.NewEncoder(w).Encode(apimodels.App{
				ID: "app-ds-1", Name: "my-ds-app", Organization: apimodels.AppOrg{Slug: "personal"},
				Status: "deployed", Network: "default", Hostname: "my-ds-app.fly.dev",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfig(server.URL) + `
data "fly_app" "test" {
  name = "my-ds-app"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_app.test", "name", "my-ds-app"),
					resource.TestCheckResourceAttr("data.fly_app.test", "status", "deployed"),
					resource.TestCheckResourceAttr("data.fly_app.test", "org_slug", "personal"),
					resource.TestCheckResourceAttr("data.fly_app.test", "hostname", "my-ds-app.fly.dev"),
				),
			},
		},
	})
}

func TestMachineDataSource_read(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/apps/ds-machine-app/machines/mach-ds-1" {
			_ = json.NewEncoder(w).Encode(apimodels.Machine{
				ID: "mach-ds-1", Name: "web-1", State: "started", Region: "iad",
				InstanceID: "inst-ds-1", PrivateIP: "fdaa::99",
				Config:    apimodels.MachineConfig{Image: "nginx:latest"},
				CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-02T00:00:00Z",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfig(server.URL) + `
data "fly_machine" "test" {
  app = "ds-machine-app"
  id  = "mach-ds-1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_machine.test", "name", "web-1"),
					resource.TestCheckResourceAttr("data.fly_machine.test", "state", "started"),
					resource.TestCheckResourceAttr("data.fly_machine.test", "region", "iad"),
					resource.TestCheckResourceAttr("data.fly_machine.test", "image", "nginx:latest"),
					resource.TestCheckResourceAttr("data.fly_machine.test", "private_ip", "fdaa::99"),
				),
			},
		},
	})
}

func TestVolumeDataSource_read(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/apps/ds-vol-app/volumes/vol-ds-1" {
			_ = json.NewEncoder(w).Encode(apimodels.Volume{
				ID: "vol-ds-1", Name: "data", Region: "iad", SizeGB: 10,
				State: "created", Encrypted: true, Zone: "iad-2",
				CreatedAt: "2024-01-01T00:00:00Z",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfig(server.URL) + `
data "fly_volume" "test" {
  app = "ds-vol-app"
  id  = "vol-ds-1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_volume.test", "name", "data"),
					resource.TestCheckResourceAttr("data.fly_volume.test", "region", "iad"),
					resource.TestCheckResourceAttr("data.fly_volume.test", "size_gb", "10"),
					resource.TestCheckResourceAttr("data.fly_volume.test", "encrypted", "true"),
					resource.TestCheckResourceAttr("data.fly_volume.test", "state", "created"),
				),
			},
		},
	})
}

func TestCertificateDataSource_read(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/apps/ds-cert-app/certificates/example.com" {
			_ = json.NewEncoder(w).Encode(apimodels.Certificate{
				ID: "cert-ds-1", Hostname: "example.com", CheckStatus: "passing",
				Source: "fly", CertificateAuthority: "lets_encrypt",
				DNSValidationHostname: "_acme.example.com",
				DNSValidationTarget:   "example.com.flydns.net",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfig(server.URL) + `
data "fly_certificate" "test" {
  app      = "ds-cert-app"
  hostname = "example.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_certificate.test", "id", "cert-ds-1"),
					resource.TestCheckResourceAttr("data.fly_certificate.test", "hostname", "example.com"),
					resource.TestCheckResourceAttr("data.fly_certificate.test", "check_status", "passing"),
					resource.TestCheckResourceAttr("data.fly_certificate.test", "source", "fly"),
				),
			},
		},
	})
}

func TestIPAddressesDataSource_read(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"ips list -a ds-ip-app --json": {
			Stdout: `[{"id":"ip-1","address":"1.2.3.4","type":"v4","region":"iad","created_at":"2024-01-01T00:00:00Z"},{"id":"ip-2","address":"2001:db8::1","type":"v6","region":"","created_at":"2024-01-02T00:00:00Z"}]`,
		},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfigWithFlyctl("http://localhost:1", flyctlPath) + `
data "fly_ip_addresses" "test" {
  app = "ds-ip-app"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_ip_addresses.test", "ip_addresses.#", "2"),
					resource.TestCheckResourceAttr("data.fly_ip_addresses.test", "ip_addresses.0.address", "1.2.3.4"),
					resource.TestCheckResourceAttr("data.fly_ip_addresses.test", "ip_addresses.0.type", "v4"),
					resource.TestCheckResourceAttr("data.fly_ip_addresses.test", "ip_addresses.1.address", "2001:db8::1"),
					resource.TestCheckResourceAttr("data.fly_ip_addresses.test", "ip_addresses.1.type", "v6"),
				),
			},
		},
	})
}

func TestAppsDataSource_read(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/apps" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"apps": []apimodels.App{
					{ID: "app-1", Name: "app-one", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"},
					{ID: "app-2", Name: "app-two", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"},
				},
				"total_apps": 2,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfig(server.URL) + `
data "fly_apps" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_apps.test", "apps.#", "2"),
					resource.TestCheckResourceAttr("data.fly_apps.test", "apps.0.name", "app-one"),
					resource.TestCheckResourceAttr("data.fly_apps.test", "apps.1.name", "app-two"),
				),
			},
		},
	})
}

func TestMachinesDataSource_read(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/apps/ds-machines-app/machines" {
			_ = json.NewEncoder(w).Encode([]apimodels.Machine{
				{ID: "mach-1", Name: "web-1", State: "started", Region: "iad", Config: apimodels.MachineConfig{Image: "nginx:latest"}, CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z"},
				{ID: "mach-2", Name: "web-2", State: "stopped", Region: "ord", Config: apimodels.MachineConfig{Image: "nginx:latest"}, CreatedAt: "2024-01-02T00:00:00Z", UpdatedAt: "2024-01-02T00:00:00Z"},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfig(server.URL) + `
data "fly_machines" "test" {
  app = "ds-machines-app"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_machines.test", "machines.#", "2"),
					resource.TestCheckResourceAttr("data.fly_machines.test", "machines.0.name", "web-1"),
					resource.TestCheckResourceAttr("data.fly_machines.test", "machines.0.state", "started"),
					resource.TestCheckResourceAttr("data.fly_machines.test", "machines.1.name", "web-2"),
					resource.TestCheckResourceAttr("data.fly_machines.test", "machines.1.state", "stopped"),
				),
			},
		},
	})
}

func TestVolumesDataSource_read(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/apps/ds-vols-app/volumes" {
			_ = json.NewEncoder(w).Encode([]apimodels.Volume{
				{ID: "vol-1", Name: "data", Region: "iad", SizeGB: 10, State: "created", Encrypted: true, CreatedAt: "2024-01-01T00:00:00Z"},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfig(server.URL) + `
data "fly_volumes" "test" {
  app = "ds-vols-app"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_volumes.test", "volumes.#", "1"),
					resource.TestCheckResourceAttr("data.fly_volumes.test", "volumes.0.name", "data"),
					resource.TestCheckResourceAttr("data.fly_volumes.test", "volumes.0.size_gb", "10"),
				),
			},
		},
	})
}

func TestCertificatesDataSource_read(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/apps/ds-certs-app/certificates" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"certificates": []apimodels.Certificate{
					{ID: "cert-1", Hostname: "example.com", CheckStatus: "passing", Source: "fly", CertificateAuthority: "lets_encrypt", CreatedAt: "2024-01-01T00:00:00Z"},
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfig(server.URL) + `
data "fly_certificates" "test" {
  app = "ds-certs-app"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_certificates.test", "certificates.#", "1"),
					resource.TestCheckResourceAttr("data.fly_certificates.test", "certificates.0.hostname", "example.com"),
					resource.TestCheckResourceAttr("data.fly_certificates.test", "certificates.0.source", "fly"),
				),
			},
		},
	})
}

func TestVolumeSnapshotsDataSource_read(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/apps/ds-snap-app/volumes/vol-1/snapshots" {
			_ = json.NewEncoder(w).Encode([]apimodels.VolumeSnapshot{
				{ID: "snap-1", Size: 1024, Status: "complete", CreatedAt: "2024-01-01T00:00:00Z"},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfig(server.URL) + `
data "fly_volume_snapshots" "test" {
  app       = "ds-snap-app"
  volume_id = "vol-1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_volume_snapshots.test", "snapshots.#", "1"),
					resource.TestCheckResourceAttr("data.fly_volume_snapshots.test", "snapshots.0.id", "snap-1"),
					resource.TestCheckResourceAttr("data.fly_volume_snapshots.test", "snapshots.0.status", "complete"),
				),
			},
		},
	})
}

func TestNetworkPoliciesDataSource_read(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/apps/ds-np-app/network_policies" {
			_ = json.NewEncoder(w).Encode([]apimodels.NetworkPolicy{
				{
					ID:       "pol-1",
					Name:     "allow-http",
					Selector: apimodels.PolicySelector{All: true},
					Rules: []apimodels.PolicyRule{
						{
							Action:    "allow",
							Direction: "egress",
							Ports:     []apimodels.PolicyPort{{Protocol: "tcp", Port: 443}},
						},
					},
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfig(server.URL) + `
data "fly_network_policies" "test" {
  app = "ds-np-app"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_network_policies.test", "policies.#", "1"),
					resource.TestCheckResourceAttr("data.fly_network_policies.test", "policies.0.name", "allow-http"),
					resource.TestCheckResourceAttr("data.fly_network_policies.test", "policies.0.rule.#", "1"),
					resource.TestCheckResourceAttr("data.fly_network_policies.test", "policies.0.rule.0.action", "allow"),
					resource.TestCheckResourceAttr("data.fly_network_policies.test", "policies.0.rule.0.direction", "egress"),
				),
			},
		},
	})
}

func TestOrganizationDataSource_read(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"orgs show personal --json": {
			Stdout: `{"id":"org-1","name":"Personal","slug":"personal","type":"PERSONAL","paid_plan":true}`,
		},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfigWithFlyctl("http://localhost:1", flyctlPath) + `
data "fly_organization" "test" {
  slug = "personal"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_organization.test", "id", "org-1"),
					resource.TestCheckResourceAttr("data.fly_organization.test", "name", "Personal"),
					resource.TestCheckResourceAttr("data.fly_organization.test", "slug", "personal"),
					resource.TestCheckResourceAttr("data.fly_organization.test", "type", "PERSONAL"),
					resource.TestCheckResourceAttr("data.fly_organization.test", "paid_plan", "true"),
				),
			},
		},
	})
}

func TestRegionsDataSource_read(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"platform regions --json": {
			Stdout: `[{"code":"iad","name":"Ashburn, Virginia (US)","gateway":true},{"code":"ord","name":"Chicago, Illinois (US)","gateway":false}]`,
		},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfigWithFlyctl("http://localhost:1", flyctlPath) + `
data "fly_regions" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_regions.test", "regions.#", "2"),
					resource.TestCheckResourceAttr("data.fly_regions.test", "regions.0.code", "iad"),
					resource.TestCheckResourceAttr("data.fly_regions.test", "regions.0.gateway", "true"),
					resource.TestCheckResourceAttr("data.fly_regions.test", "regions.1.code", "ord"),
				),
			},
		},
	})
}

func TestMPGClustersDataSource_read(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"mpg list --json": {
			Stdout: `[{"id":"mpg-1","name":"pg-cluster-1","status":"running","region":"iad"},{"id":"mpg-2","name":"pg-cluster-2","status":"stopped","region":"ord"}]`,
		},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfigWithFlyctl("http://localhost:1", flyctlPath) + `
data "fly_mpg_clusters" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_mpg_clusters.test", "clusters.#", "2"),
					resource.TestCheckResourceAttr("data.fly_mpg_clusters.test", "clusters.0.id", "mpg-1"),
					resource.TestCheckResourceAttr("data.fly_mpg_clusters.test", "clusters.0.name", "pg-cluster-1"),
					resource.TestCheckResourceAttr("data.fly_mpg_clusters.test", "clusters.0.status", "running"),
					resource.TestCheckResourceAttr("data.fly_mpg_clusters.test", "clusters.0.region", "iad"),
					resource.TestCheckResourceAttr("data.fly_mpg_clusters.test", "clusters.1.name", "pg-cluster-2"),
					resource.TestCheckResourceAttr("data.fly_mpg_clusters.test", "clusters.1.status", "stopped"),
				),
			},
		},
	})
}

func TestRedisInstancesDataSource_read(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"redis list --json": {
			Stdout: `[{"id":"redis-1","name":"my-redis","status":"running","plan":"free","region":"iad"},{"id":"redis-2","name":"prod-redis","status":"running","plan":"pro","region":"ord"}]`,
		},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfigWithFlyctl("http://localhost:1", flyctlPath) + `
data "fly_redis_instances" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_redis_instances.test", "instances.#", "2"),
					resource.TestCheckResourceAttr("data.fly_redis_instances.test", "instances.0.id", "redis-1"),
					resource.TestCheckResourceAttr("data.fly_redis_instances.test", "instances.0.name", "my-redis"),
					resource.TestCheckResourceAttr("data.fly_redis_instances.test", "instances.0.plan", "free"),
					resource.TestCheckResourceAttr("data.fly_redis_instances.test", "instances.0.region", "iad"),
					resource.TestCheckResourceAttr("data.fly_redis_instances.test", "instances.1.name", "prod-redis"),
					resource.TestCheckResourceAttr("data.fly_redis_instances.test", "instances.1.plan", "pro"),
				),
			},
		},
	})
}

func TestTigrisBucketsDataSource_read(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"storage list --json": {
			Stdout: `[{"id":"bucket-1","name":"my-bucket","public":false},{"id":"bucket-2","name":"public-assets","public":true}]`,
		},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfigWithFlyctl("http://localhost:1", flyctlPath) + `
data "fly_tigris_buckets" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_tigris_buckets.test", "buckets.#", "2"),
					resource.TestCheckResourceAttr("data.fly_tigris_buckets.test", "buckets.0.id", "bucket-1"),
					resource.TestCheckResourceAttr("data.fly_tigris_buckets.test", "buckets.0.name", "my-bucket"),
					resource.TestCheckResourceAttr("data.fly_tigris_buckets.test", "buckets.0.public", "false"),
					resource.TestCheckResourceAttr("data.fly_tigris_buckets.test", "buckets.1.name", "public-assets"),
					resource.TestCheckResourceAttr("data.fly_tigris_buckets.test", "buckets.1.public", "true"),
				),
			},
		},
	})
}

func TestTokensDataSource_read(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"tokens list --app my-app --json": {
			Stdout: `[{"id":"tok-1","name":"deploy-token","type":"deploy","created_at":"2024-01-01T00:00:00Z"},{"id":"tok-2","name":"read-token","type":"read","created_at":"2024-02-01T00:00:00Z"}]`,
		},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfigWithFlyctl("http://localhost:1", flyctlPath) + `
data "fly_tokens" "test" {
  app = "my-app"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_tokens.test", "tokens.#", "2"),
					resource.TestCheckResourceAttr("data.fly_tokens.test", "tokens.0.id", "tok-1"),
					resource.TestCheckResourceAttr("data.fly_tokens.test", "tokens.0.name", "deploy-token"),
					resource.TestCheckResourceAttr("data.fly_tokens.test", "tokens.0.type", "deploy"),
					resource.TestCheckResourceAttr("data.fly_tokens.test", "tokens.1.name", "read-token"),
					resource.TestCheckResourceAttr("data.fly_tokens.test", "tokens.1.type", "read"),
				),
			},
		},
	})
}

func TestOIDCTokenDataSource_read(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/tokens/oidc" {
			_ = json.NewEncoder(w).Encode(map[string]string{
				"token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.mock-token",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDSProviderConfig(server.URL) + `
data "fly_oidc_token" "test" {
  aud = "https://example.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_oidc_token.test", "token", "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.mock-token"),
				),
			},
		},
	})
}
