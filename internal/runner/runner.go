package runner

import (
	"context"
	"os/exec"
)

// Runner represents a mechanism to run arbitrary commands
type Runner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// SubprocessRunner is a real implementation of Runner that shells out to arbitrary commands
type SubprocessRunner struct{}

// NewSubprocessRunner creates a new SubprocessRunner
func NewSubprocessRunner() *SubprocessRunner {
	return &SubprocessRunner{}
}

// Run executes a command and returns its combined output
func (r *SubprocessRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}
