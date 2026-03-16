package flyctl

import (
	"context"
	"encoding/json"
	"testing"
)

func TestExecutor_RunJSON(t *testing.T) {
	type result struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}

	mock := NewMockRunner(map[string]MockResponse{
		"orgs show personal --json": {
			Stdout: `{"name":"Personal","status":"active"}`,
		},
	})

	exec := NewExecutorWithRunner("test-token", mock)

	var got result
	err := exec.RunJSON(context.Background(), &got, "orgs", "show", "personal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Personal" {
		t.Errorf("got name %q, want %q", got.Name, "Personal")
	}
	if got.Status != "active" {
		t.Errorf("got status %q, want %q", got.Status, "active")
	}
}

func TestExecutor_RunJSON_error(t *testing.T) {
	mock := NewMockRunner(map[string]MockResponse{
		"orgs show nonexistent --json": {
			Stderr:   "Error: organization not found",
			ExitCode: 1,
		},
	})

	exec := NewExecutorWithRunner("test-token", mock)

	var got map[string]any
	err := exec.RunJSON(context.Background(), &got, "orgs", "show", "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

func TestExecutor_Run(t *testing.T) {
	mock := NewMockRunner(map[string]MockResponse{
		"ips release 1.2.3.4 -a my-app --yes": {
			Stdout: "Released 1.2.3.4\n",
		},
	})

	exec := NewExecutorWithRunner("test-token", mock)

	result, err := exec.Run(context.Background(), "ips", "release", "1.2.3.4", "-a", "my-app", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result.Stdout) != "Released 1.2.3.4\n" {
		t.Errorf("unexpected stdout: %s", result.Stdout)
	}
}

func TestExecutor_RunJSON_list(t *testing.T) {
	ips := []map[string]string{
		{"id": "ip-1", "address": "1.2.3.4", "type": "v4"},
		{"id": "ip-2", "address": "2001:db8::1", "type": "v6"},
	}
	data, _ := json.Marshal(ips)

	mock := NewMockRunner(map[string]MockResponse{
		"ips list -a my-app --json": {
			Stdout: string(data),
		},
	})

	exec := NewExecutorWithRunner("test-token", mock)

	var got []map[string]string
	err := exec.RunJSON(context.Background(), &got, "ips", "list", "-a", "my-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d items, want 2", len(got))
	}
	if got[0]["address"] != "1.2.3.4" {
		t.Errorf("got address %q, want %q", got[0]["address"], "1.2.3.4")
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		stderr string
		want   bool
	}{
		{"Error: organization not found", true},
		{"Error: could not find app", true},
		{"Error: no such resource", true},
		{"Error: resource does not exist", true},
		{"Error: permission denied", false},
		{"Error: timeout", false},
	}

	for _, tt := range tests {
		err := &FlyctlError{ExitCode: 1, Stderr: tt.stderr, Command: "test"}
		if got := IsNotFound(err); got != tt.want {
			t.Errorf("IsNotFound(%q) = %v, want %v", tt.stderr, got, tt.want)
		}
	}
}

func TestIsAlreadyExists(t *testing.T) {
	tests := []struct {
		stderr string
		want   bool
	}{
		{"Error: app already exists", true},
		{"Error: conflict", true},
		{"Error: duplicate entry", true},
		{"Error: not found", false},
	}

	for _, tt := range tests {
		err := &FlyctlError{ExitCode: 1, Stderr: tt.stderr, Command: "test"}
		if got := IsAlreadyExists(err); got != tt.want {
			t.Errorf("IsAlreadyExists(%q) = %v, want %v", tt.stderr, got, tt.want)
		}
	}
}

func TestMockRunner_calls(t *testing.T) {
	mock := NewMockRunner(map[string]MockResponse{
		"ips list --json": {Stdout: "[]"},
	})

	exec := NewExecutorWithRunner("token", mock)
	var out []any
	_ = exec.RunJSON(context.Background(), &out, "ips", "list")

	if len(mock.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.Calls))
	}
}
