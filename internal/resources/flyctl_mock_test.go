package resources_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// flyctlMockResponse defines a canned response for a flyctl command pattern.
type flyctlMockResponse struct {
	Stdout   string
	ExitCode int
}

// createMockFlyctl creates a temporary shell script that acts as a mock flyctl binary.
// The script matches command arguments against patterns and returns canned responses.
// Returns the path to the script; cleanup is handled by t.Cleanup.
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
# Mock flyctl binary for testing
# Join all args into a single string for matching
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

// providerConfigWithFlyctl generates a provider config block with both api_url and flyctl_path.
func providerConfigWithFlyctl(apiURL, flyctlPath string) string {
	return fmt.Sprintf(`
provider "fly" {
  api_token   = "mock-token"
  api_url     = %q
  flyctl_path = %q
}
`, apiURL, flyctlPath)
}
