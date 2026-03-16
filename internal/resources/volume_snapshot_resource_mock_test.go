package resources_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

func TestVolumeSnapshotResource_lifecycle(t *testing.T) {
	snapshot := apimodels.VolumeSnapshot{
		ID:        "snap-abc123",
		Size:      1073741824,
		Digest:    "sha256:abc",
		Status:    "complete",
		CreatedAt: "2024-01-01T00:00:00Z",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/apps":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(apimodels.App{ID: "app-snap", Name: "snap-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "GET" && r.URL.Path == "/apps/snap-test-app" && !strings.Contains(r.URL.Path, "/volumes"):
			json.NewEncoder(w).Encode(apimodels.App{ID: "app-snap", Name: "snap-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "DELETE" && r.URL.Path == "/apps/snap-test-app":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/snapshots"):
			json.NewEncoder(w).Encode(snapshot)
		case r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/snapshots"):
			json.NewEncoder(w).Encode([]apimodels.VolumeSnapshot{snapshot})
		default:
			t.Logf("Unhandled: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	}))
	defer server.Close()

	config := `
provider "fly" {
  api_token = "mock-token"
  api_url   = "` + server.URL + `"
}

resource "fly_volume_snapshot" "test" {
  app       = "snap-test-app"
  volume_id = "vol-123"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_volume_snapshot.test", "id", "snap-abc123"),
					resource.TestCheckResourceAttr("fly_volume_snapshot.test", "status", "complete"),
					resource.TestCheckResourceAttr("fly_volume_snapshot.test", "digest", "sha256:abc"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
