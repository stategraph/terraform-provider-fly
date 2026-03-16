package flyctl

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultTimeout = 2 * time.Minute
	createTimeout  = 10 * time.Minute
)

// CommandRunner abstracts command execution for testing.
type CommandRunner interface {
	Run(ctx context.Context, args []string, env []string) (stdout, stderr []byte, exitCode int, err error)
}

// Result holds the output of a flyctl command.
type Result struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// ExecRunner executes commands using os/exec.
type ExecRunner struct{}

func (r *ExecRunner) Run(ctx context.Context, args []string, env []string) (stdout, stderr []byte, exitCode int, err error) {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Env = env

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	stdout = []byte(outBuf.String())
	stderr = []byte(errBuf.String())

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return stdout, stderr, exitErr.ExitCode(), nil
		}
		return stdout, stderr, -1, err
	}
	return stdout, stderr, 0, nil
}

// Executor runs flyctl commands.
type Executor struct {
	binaryPath string
	token      string
	runner     CommandRunner
	mu         sync.Mutex
}

// NewExecutor creates a new flyctl executor.
func NewExecutor(binaryPath, token string) *Executor {
	return &Executor{
		binaryPath: binaryPath,
		token:      token,
		runner:     &ExecRunner{},
	}
}

// NewExecutorWithRunner creates a new flyctl executor with a custom runner (for testing).
func NewExecutorWithRunner(token string, runner CommandRunner) *Executor {
	return &Executor{
		binaryPath: "flyctl",
		token:      token,
		runner:     runner,
	}
}

// RunJSON executes a flyctl command, appends --json, and unmarshals the JSON output into target.
func (e *Executor) RunJSON(ctx context.Context, target any, args ...string) error {
	args = append(args, "--json")
	result, err := e.Run(ctx, args...)
	if err != nil {
		return err
	}
	if target != nil && len(result.Stdout) > 0 {
		if err := json.Unmarshal(result.Stdout, target); err != nil {
			return fmt.Errorf("parsing flyctl JSON output: %w (output: %s)", err, string(result.Stdout))
		}
	}
	return nil
}

// Run executes a flyctl command and returns the result.
func (e *Executor) Run(ctx context.Context, args ...string) (*Result, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	timeout := timeoutForCommand(args)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	fullArgs := append([]string{e.binaryPath}, args...)
	env := e.buildEnv()

	cmdStr := strings.Join(fullArgs, " ")
	tflog.Debug(ctx, "executing flyctl command", map[string]any{"command": cmdStr})

	stdout, stderr, exitCode, err := e.runner.Run(ctx, fullArgs, env)
	if err != nil {
		return nil, fmt.Errorf("executing flyctl: %w", err)
	}

	if exitCode != 0 {
		tflog.Debug(ctx, "flyctl command failed", map[string]any{
			"command":   cmdStr,
			"exit_code": exitCode,
			"stderr":    strings.TrimSpace(string(stderr)),
		})
		return nil, &FlyctlError{
			ExitCode: exitCode,
			Stderr:   strings.TrimSpace(string(stderr)),
			Command:  strings.Join(args, " "),
		}
	}

	tflog.Debug(ctx, "flyctl command succeeded", map[string]any{
		"command":       cmdStr,
		"stdout_length": len(stdout),
	})

	return &Result{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
	}, nil
}

func (e *Executor) buildEnv() []string {
	return []string{
		"FLY_API_TOKEN=" + e.token,
		"HOME=" + homeDir(),
		"PATH=" + pathEnv(),
	}
}

func timeoutForCommand(args []string) time.Duration {
	for _, arg := range args {
		if arg == "create" || arg == "attach" {
			return createTimeout
		}
	}
	return defaultTimeout
}

// FindBinary locates the flyctl binary. Checks the provided path first,
// then FLYCTL_PATH env var, then PATH lookup.
func FindBinary(configPath string) (string, error) {
	if configPath != "" {
		return configPath, nil
	}
	// Try common names
	for _, name := range []string{"flyctl", "fly"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("flyctl binary not found in PATH; set flyctl_path in provider config or FLYCTL_PATH environment variable")
}
