package apiclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

func TestCreateApp(t *testing.T) {
	expected := apimodels.App{
		ID:      "app-123",
		Name:    "test-app",
		OrgSlug: "personal",
		Status:  "pending",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/apps" {
			t.Errorf("expected /apps, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Authorization header with test-token")
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("expected User-Agent header")
		}

		var req apimodels.CreateAppRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		if req.AppName != "test-app" {
			t.Errorf("expected app_name=test-app, got %s", req.AppName)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewClient("test-token", "test", WithBaseURL(server.URL))
	app, err := client.CreateApp(context.Background(), apimodels.CreateAppRequest{
		AppName: "test-app",
		OrgSlug: "personal",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.ID != expected.ID {
		t.Errorf("expected ID=%s, got %s", expected.ID, app.ID)
	}
	if app.Name != expected.Name {
		t.Errorf("expected Name=%s, got %s", expected.Name, app.Name)
	}
}

func TestGetApp_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))
	defer server.Close()

	client := NewClient("test-token", "test", WithBaseURL(server.URL))
	_, err := client.GetApp(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound=true, got false for error: %v", err)
	}
}

func TestGetApp_Success(t *testing.T) {
	expected := apimodels.App{
		ID:     "app-456",
		Name:   "my-app",
		Status: "deployed",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/apps/my-app" {
			t.Errorf("expected /apps/my-app, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewClient("test-token", "test", WithBaseURL(server.URL))
	app, err := client.GetApp(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.Name != expected.Name {
		t.Errorf("expected Name=%s, got %s", expected.Name, app.Name)
	}
}

func TestDeleteApp_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("test-token", "test", WithBaseURL(server.URL))
	err := client.DeleteApp(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRetryOn429(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"error": "rate limited"})
			return
		}
		json.NewEncoder(w).Encode(apimodels.App{ID: "ok", Name: "ok"})
	}))
	defer server.Close()

	client := NewClient("test-token", "test", WithBaseURL(server.URL))
	app, err := client.GetApp(context.Background(), "test")
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if app.Name != "ok" {
		t.Errorf("expected Name=ok, got %s", app.Name)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryOn500(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "internal"})
			return
		}
		json.NewEncoder(w).Encode(apimodels.App{ID: "ok", Name: "ok"})
	}))
	defer server.Close()

	client := NewClient("test-token", "test", WithBaseURL(server.URL))
	_, err := client.GetApp(context.Background(), "test")
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
}

func TestNoRetryOn422(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid"})
	}))
	defer server.Close()

	client := NewClient("test-token", "test", WithBaseURL(server.URL))
	_, err := client.GetApp(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for 422")
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt (no retry), got %d", attempts)
	}
}

func TestCreateMachine(t *testing.T) {
	expected := apimodels.Machine{
		ID:     "mach-123",
		Name:   "test-machine",
		State:  "started",
		Region: "iad",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/apps/test-app/machines" {
			t.Errorf("expected /apps/test-app/machines, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewClient("test-token", "test", WithBaseURL(server.URL))
	machine, err := client.CreateMachine(context.Background(), "test-app", apimodels.CreateMachineRequest{
		Region: "iad",
		Config: apimodels.MachineConfig{Image: "nginx"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if machine.ID != expected.ID {
		t.Errorf("expected ID=%s, got %s", expected.ID, machine.ID)
	}
}

func TestCreateVolume(t *testing.T) {
	expected := apimodels.Volume{
		ID:     "vol-123",
		Name:   "data",
		Region: "iad",
		SizeGB: 10,
		State:  "created",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/apps/test-app/volumes" {
			t.Errorf("expected /apps/test-app/volumes, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewClient("test-token", "test", WithBaseURL(server.URL))
	volume, err := client.CreateVolume(context.Background(), "test-app", apimodels.CreateVolumeRequest{
		Name:   "data",
		Region: "iad",
		SizeGB: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if volume.SizeGB != 10 {
		t.Errorf("expected SizeGB=10, got %d", volume.SizeGB)
	}
}

func TestExtendVolume(t *testing.T) {
	expected := apimodels.ExtendVolumeResponse{
		Volume: apimodels.Volume{
			ID:     "vol-123",
			SizeGB: 20,
		},
		NeedsRestart: true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewClient("test-token", "test", WithBaseURL(server.URL))
	resp, err := client.ExtendVolume(context.Background(), "test-app", "vol-123", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Volume.SizeGB != 20 {
		t.Errorf("expected SizeGB=20, got %d", resp.Volume.SizeGB)
	}
	if !resp.NeedsRestart {
		t.Error("expected NeedsRestart=true")
	}
}

func TestIsConflict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "conflict"})
	}))
	defer server.Close()

	client := NewClient("test-token", "test", WithBaseURL(server.URL))
	_, err := client.GetApp(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsConflict(err) {
		t.Errorf("expected IsConflict=true, got false")
	}
}

func TestMalformedJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient("test-token", "test", WithBaseURL(server.URL))
	_, err := client.GetApp(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestUserAgentHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua != "terraform-provider-fly/1.0.0" {
			t.Errorf("expected User-Agent=terraform-provider-fly/1.0.0, got %s", ua)
		}
		json.NewEncoder(w).Encode(apimodels.App{})
	}))
	defer server.Close()

	client := NewClient("test-token", "1.0.0", WithBaseURL(server.URL))
	_, _ = client.GetApp(context.Background(), "test")
}

