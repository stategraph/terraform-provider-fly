package resources_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

func TestVolumeResource_lifecycle(t *testing.T) {
	volume := apimodels.Volume{
		ID:        "vol-test-123",
		Name:      "testdata",
		Region:    "iad",
		SizeGB:    1,
		State:     "created",
		Encrypted: true,
		Zone:      "iad-1",
		CreatedAt: "2024-01-01T00:00:00Z",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/apps":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(apimodels.App{ID: "app-vol", Name: "vol-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "GET" && r.URL.Path == "/apps/vol-test-app" && !strings.Contains(r.URL.Path, "/volumes"):
			json.NewEncoder(w).Encode(apimodels.App{ID: "app-vol", Name: "vol-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "DELETE" && r.URL.Path == "/apps/vol-test-app":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == "POST" && r.URL.Path == "/apps/vol-test-app/volumes":
			json.NewEncoder(w).Encode(volume)
		case r.Method == "GET" && r.URL.Path == "/apps/vol-test-app/volumes/vol-test-123":
			json.NewEncoder(w).Encode(volume)
		case r.Method == "DELETE" && r.URL.Path == "/apps/vol-test-app/volumes/vol-test-123":
			w.WriteHeader(http.StatusOK)
		default:
			t.Logf("Unhandled: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	}))
	defer server.Close()

	config := testVolumeConfigWithURL(server.URL, 1)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_volume.test", "id", "vol-test-123"),
					resource.TestCheckResourceAttr("fly_volume.test", "name", "testdata"),
					resource.TestCheckResourceAttr("fly_volume.test", "region", "iad"),
					resource.TestCheckResourceAttr("fly_volume.test", "size_gb", "1"),
					resource.TestCheckResourceAttr("fly_volume.test", "encrypted", "true"),
					resource.TestCheckResourceAttr("fly_volume.test", "state", "created"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

func TestVolumeResource_extend(t *testing.T) {
	currentSize := 1

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vol := apimodels.Volume{
			ID: "vol-ext-123", Name: "testdata", Region: "iad",
			SizeGB: currentSize, State: "created", Encrypted: true,
			CreatedAt: "2024-01-01T00:00:00Z",
		}
		switch {
		case r.Method == "POST" && r.URL.Path == "/apps":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(apimodels.App{ID: "app-ext", Name: "ext-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "GET" && r.URL.Path == "/apps/ext-test-app" && !strings.Contains(r.URL.Path, "/volumes"):
			json.NewEncoder(w).Encode(apimodels.App{ID: "app-ext", Name: "ext-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "DELETE" && r.URL.Path == "/apps/ext-test-app":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == "POST" && r.URL.Path == "/apps/ext-test-app/volumes":
			json.NewEncoder(w).Encode(vol)
		case r.Method == "GET" && r.URL.Path == "/apps/ext-test-app/volumes/vol-ext-123":
			json.NewEncoder(w).Encode(vol)
		case r.Method == "PUT" && strings.Contains(r.URL.Path, "/extend"):
			currentSize = 2
			json.NewEncoder(w).Encode(apimodels.ExtendVolumeResponse{
				Volume:       apimodels.Volume{ID: "vol-ext-123", Name: "testdata", Region: "iad", SizeGB: 2, State: "created", Encrypted: true, CreatedAt: "2024-01-01T00:00:00Z"},
				NeedsRestart: false,
			})
		case r.Method == "DELETE" && r.URL.Path == "/apps/ext-test-app/volumes/vol-ext-123":
			w.WriteHeader(http.StatusOK)
		default:
			t.Logf("Unhandled: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testVolumeExtendConfigWithURL(server.URL, 1),
				Check:  resource.TestCheckResourceAttr("fly_volume.test", "size_gb", "1"),
			},
			{
				Config: testVolumeExtendConfigWithURL(server.URL, 2),
				Check:  resource.TestCheckResourceAttr("fly_volume.test", "size_gb", "2"),
			},
		},
	})
}

func testVolumeConfigWithURL(apiURL string, sizeGB int) string {
	return testVolumeConfigGeneric(apiURL, "vol-test-app", sizeGB)
}

func testVolumeExtendConfigWithURL(apiURL string, sizeGB int) string {
	return testVolumeConfigGeneric(apiURL, "ext-test-app", sizeGB)
}

func testVolumeConfigGeneric(apiURL, appName string, sizeGB int) string {
	return `
provider "fly" {
  api_token = "mock-token"
  api_url   = "` + apiURL + `"
}

resource "fly_app" "test" {
  name     = "` + appName + `"
  org_slug = "personal"
}

resource "fly_volume" "test" {
  app     = fly_app.test.name
  name    = "testdata"
  region  = "iad"
  size_gb = ` + fmt.Sprintf("%d", sizeGB) + `
}
`
}
