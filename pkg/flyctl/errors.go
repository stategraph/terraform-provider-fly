package flyctl

import (
	"errors"
	"fmt"
	"strings"
)

// FlyctlError represents an error from running a flyctl command.
type FlyctlError struct {
	ExitCode int
	Stderr   string
	Command  string
}

func (e *FlyctlError) Error() string {
	return fmt.Sprintf("flyctl error (exit %d) running %q: %s", e.ExitCode, e.Command, e.Stderr)
}

// IsNotFound returns true if the error indicates a resource was not found.
func IsNotFound(err error) bool {
	var flyErr *FlyctlError
	if errors.As(err, &flyErr) {
		stderr := strings.ToLower(flyErr.Stderr)
		return strings.Contains(stderr, "not found") ||
			strings.Contains(stderr, "could not find") ||
			strings.Contains(stderr, "no such") ||
			strings.Contains(stderr, "does not exist")
	}
	return false
}

// IsAlreadyExists returns true if the error indicates a resource already exists.
func IsAlreadyExists(err error) bool {
	var flyErr *FlyctlError
	if errors.As(err, &flyErr) {
		stderr := strings.ToLower(flyErr.Stderr)
		return strings.Contains(stderr, "already exists") ||
			strings.Contains(stderr, "conflict") ||
			strings.Contains(stderr, "duplicate")
	}
	return false
}
