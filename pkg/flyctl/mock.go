package flyctl

import (
	"context"
	"fmt"
	"strings"
)

// MockResponse represents a canned response for a mock command.
type MockResponse struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// MockRunner is a CommandRunner for testing that maps command patterns to canned responses.
type MockRunner struct {
	// Responses maps a command key to a response. The key is matched against
	// the joined args string (excluding the binary path).
	Responses map[string]MockResponse
	// Calls records all commands that were executed.
	Calls [][]string
}

// NewMockRunner creates a MockRunner with the given responses.
func NewMockRunner(responses map[string]MockResponse) *MockRunner {
	return &MockRunner{
		Responses: responses,
	}
}

func (m *MockRunner) Run(_ context.Context, args []string, _ []string) (stdout, stderr []byte, exitCode int, err error) {
	// Strip the binary path from args for matching.
	cmdArgs := args[1:]
	m.Calls = append(m.Calls, cmdArgs)

	key := strings.Join(cmdArgs, " ")

	// Try exact match first.
	if resp, ok := m.Responses[key]; ok {
		return []byte(resp.Stdout), []byte(resp.Stderr), resp.ExitCode, nil
	}

	// Try prefix match (longest first).
	var bestMatch string
	var bestResp MockResponse
	for pattern, resp := range m.Responses {
		if strings.HasPrefix(key, pattern) && len(pattern) > len(bestMatch) {
			bestMatch = pattern
			bestResp = resp
		}
	}
	if bestMatch != "" {
		return []byte(bestResp.Stdout), []byte(bestResp.Stderr), bestResp.ExitCode, nil
	}

	return nil, []byte(fmt.Sprintf("mock: no response for command %q", key)), 1, nil
}
