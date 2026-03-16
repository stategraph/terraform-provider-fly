package resources_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

func TestMachineResource_lifecycle(t *testing.T) {
	machine := apimodels.Machine{
		ID:         "mach-abc123",
		Name:       "test-machine",
		State:      "started",
		Region:     "iad",
		InstanceID: "inst-001",
		PrivateIP:  "fdaa::1",
		Config: apimodels.MachineConfig{
			Image: "nginx:latest",
			Env:   map[string]string{"PORT": "8080"},
			Guest: &apimodels.MachineGuest{
				CPUKind:  "shared",
				CPUs:     1,
				MemoryMB: 256,
			},
		},
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		// App endpoints
		case r.Method == "POST" && r.URL.Path == "/apps":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(apimodels.App{ID: "app-1", Name: "machine-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/apps/machine-test-app") && !strings.Contains(r.URL.Path, "/machines"):
			_ = json.NewEncoder(w).Encode(apimodels.App{ID: "app-1", Name: "machine-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "DELETE" && r.URL.Path == "/apps/machine-test-app":
			w.WriteHeader(http.StatusNoContent)

		// Machine create
		case r.Method == "POST" && r.URL.Path == "/apps/machine-test-app/machines":
			_ = json.NewEncoder(w).Encode(machine)

		// Machine get
		case r.Method == "GET" && r.URL.Path == "/apps/machine-test-app/machines/mach-abc123" && r.URL.RawQuery == "":
			_ = json.NewEncoder(w).Encode(machine)

		// Machine wait
		case r.Method == "GET" && r.URL.Path == "/apps/machine-test-app/machines/mach-abc123/wait":
			w.WriteHeader(http.StatusOK)

		// Machine update
		case r.Method == "POST" && r.URL.Path == "/apps/machine-test-app/machines/mach-abc123":
			_ = json.NewEncoder(w).Encode(machine)

		// Machine lease
		case r.Method == "POST" && r.URL.Path == "/apps/machine-test-app/machines/mach-abc123/lease":
			_ = json.NewEncoder(w).Encode(apimodels.Lease{Nonce: "lease-nonce-1", ExpiresAt: 9999999999})
		case r.Method == "DELETE" && r.URL.Path == "/apps/machine-test-app/machines/mach-abc123/lease":
			w.WriteHeader(http.StatusOK)

		// Machine stop
		case r.Method == "POST" && r.URL.Path == "/apps/machine-test-app/machines/mach-abc123/stop":
			w.WriteHeader(http.StatusOK)

		// Machine delete
		case r.Method == "DELETE" && r.URL.Path == "/apps/machine-test-app/machines/mach-abc123":
			w.WriteHeader(http.StatusOK)

		default:
			t.Logf("Unhandled: %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	}))
	defer server.Close()

	config := testMachineConfigWithURL(server.URL)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_machine.test", "id", "mach-abc123"),
					resource.TestCheckResourceAttr("fly_machine.test", "app", "machine-test-app"),
					resource.TestCheckResourceAttr("fly_machine.test", "region", "iad"),
					resource.TestCheckResourceAttr("fly_machine.test", "image", "nginx:latest"),
					resource.TestCheckResourceAttr("fly_machine.test", "state", "started"),
					resource.TestCheckResourceAttr("fly_machine.test", "private_ip", "fdaa::1"),
					resource.TestCheckResourceAttr("fly_machine.test", "instance_id", "inst-001"),
					resource.TestCheckResourceAttr("fly_machine.test", "env.PORT", "8080"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

func TestMachineResource_withServices(t *testing.T) {
	machine := apimodels.Machine{
		ID:         "mach-svc123",
		Name:       "svc-machine",
		State:      "started",
		Region:     "iad",
		InstanceID: "inst-002",
		PrivateIP:  "fdaa::2",
		Config: apimodels.MachineConfig{
			Image: "nginx:latest",
			Guest: &apimodels.MachineGuest{CPUKind: "shared", CPUs: 1, MemoryMB: 256},
			Services: []apimodels.MachineService{
				{
					Protocol:     "tcp",
					InternalPort: 8080,
					Ports: []apimodels.MachinePort{
						{Port: intPtr(80), Handlers: []string{"http"}},
						{Port: intPtr(443), Handlers: []string{"tls", "http"}},
					},
				},
			},
		},
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/apps":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(apimodels.App{ID: "app-svc", Name: "svc-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "GET" && r.URL.Path == "/apps/svc-test-app" && !strings.Contains(r.URL.Path, "/machines"):
			_ = json.NewEncoder(w).Encode(apimodels.App{ID: "app-svc", Name: "svc-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "DELETE" && r.URL.Path == "/apps/svc-test-app":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == "POST" && r.URL.Path == "/apps/svc-test-app/machines":
			_ = json.NewEncoder(w).Encode(machine)
		case r.Method == "GET" && r.URL.Path == "/apps/svc-test-app/machines/mach-svc123" && r.URL.RawQuery == "":
			_ = json.NewEncoder(w).Encode(machine)
		case r.Method == "GET" && strings.Contains(r.URL.Path, "/wait"):
			w.WriteHeader(http.StatusOK)
		case r.Method == "POST" && strings.Contains(r.URL.Path, "/stop"):
			w.WriteHeader(http.StatusOK)
		case r.Method == "DELETE" && r.URL.Path == "/apps/svc-test-app/machines/mach-svc123":
			w.WriteHeader(http.StatusOK)
		default:
			t.Logf("Unhandled: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	}))
	defer server.Close()

	svcConfig := testMachineWithServicesConfigWithURL(server.URL)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: svcConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_machine.test", "service.#", "1"),
					resource.TestCheckResourceAttr("fly_machine.test", "service.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("fly_machine.test", "service.0.internal_port", "8080"),
					resource.TestCheckResourceAttr("fly_machine.test", "service.0.port.#", "2"),
					resource.TestCheckResourceAttr("fly_machine.test", "service.0.port.0.port", "80"),
					resource.TestCheckResourceAttr("fly_machine.test", "service.0.port.1.port", "443"),
				),
			},
			{
				Config:   svcConfig,
				PlanOnly: true,
			},
		},
	})
}

func TestMachineResource_invalidImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/apps":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(apimodels.App{ID: "app-err", Name: "err-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "GET" && r.URL.Path == "/apps/err-test-app":
			_ = json.NewEncoder(w).Encode(apimodels.App{ID: "app-err", Name: "err-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "POST" && strings.Contains(r.URL.Path, "/machines"):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid image reference"})
		case r.Method == "DELETE":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testMachineConfigWithURL_customApp(server.URL, "err-test-app"),
				ExpectError: regexp.MustCompile(`invalid\s+image\s+reference`),
			},
		},
	})
}

func intPtr(i int) *int { return &i }

func testMachineConfigWithURL(apiURL string) string {
	return `
provider "fly" {
  api_token = "mock-token"
  api_url   = "` + apiURL + `"
}

resource "fly_app" "test" {
  name     = "machine-test-app"
  org_slug = "personal"
}

resource "fly_machine" "test" {
  app    = fly_app.test.name
  region = "iad"
  image  = "nginx:latest"

  env = {
    PORT = "8080"
  }

  guest {
    cpu_kind  = "shared"
    cpus      = 1
    memory_mb = 256
  }
}
`
}

func testMachineWithServicesConfigWithURL(apiURL string) string {
	return `
provider "fly" {
  api_token = "mock-token"
  api_url   = "` + apiURL + `"
}

resource "fly_app" "test" {
  name     = "svc-test-app"
  org_slug = "personal"
}

resource "fly_machine" "test" {
  app    = fly_app.test.name
  region = "iad"
  image  = "nginx:latest"

  guest {
    cpu_kind  = "shared"
    cpus      = 1
    memory_mb = 256
  }

  service {
    protocol      = "tcp"
    internal_port = 8080

    port {
      port     = 80
      handlers = ["http"]
    }

    port {
      port     = 443
      handlers = ["tls", "http"]
    }
  }
}
`
}

func testMachineConfigWithURL_customApp(apiURL, appName string) string {
	return `
provider "fly" {
  api_token = "mock-token"
  api_url   = "` + apiURL + `"
}

resource "fly_app" "test" {
  name     = "` + appName + `"
  org_slug = "personal"
}

resource "fly_machine" "test" {
  app    = fly_app.test.name
  region = "iad"
  image  = "bad-image"

  guest {
    cpu_kind  = "shared"
    cpus      = 1
    memory_mb = 256
  }
}
`
}
