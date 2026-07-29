package installer_test

import (
	"context"
	"testing"

	"setupper/internal/installer"
	"setupper/internal/manifest"
	"setupper/internal/planner"
	"setupper/internal/runner"
)

func TestExecuteStep_Install(t *testing.T) {
	tests := []struct {
		name         string
		resource     manifest.Resource
		expectedCmd  string
		expectedArgs []string
	}{
		{
			name:         "vscode-extension",
			resource:     manifest.Resource{Type: "vscode-extension", Name: "golang.Go"},
			expectedCmd:  "code",
			expectedArgs: []string{"--install-extension", "golang.Go"},
		},
		{
			name:         "cursor-extension",
			resource:     manifest.Resource{Type: "cursor-extension", Name: "jupyter"},
			expectedCmd:  "cursor",
			expectedArgs: []string{"--install-extension", "jupyter"},
		},
		{
			name:         "npm",
			resource:     manifest.Resource{Type: "npm", Name: "typescript"},
			expectedCmd:  "npm",
			expectedArgs: []string{"install", "-g", "typescript"},
		},
		{
			name:         "pnpm",
			resource:     manifest.Resource{Type: "pnpm", Name: "sass"},
			expectedCmd:  "pnpm",
			expectedArgs: []string{"add", "-g", "sass"},
		},
		{
			name:         "cargo",
			resource:     manifest.Resource{Type: "cargo", Name: "ripgrep"},
			expectedCmd:  "cargo",
			expectedArgs: []string{"install", "ripgrep"},
		},
		{
			name:         "pipx",
			resource:     manifest.Resource{Type: "pipx", Name: "black"},
			expectedCmd:  "pipx",
			expectedArgs: []string{"install", "black"},
		},
		{
			name:         "uv",
			resource:     manifest.Resource{Type: "uv", Name: "ruff"},
			expectedCmd:  "uv",
			expectedArgs: []string{"tool", "install", "ruff"},
		},
		{
			name:         "go",
			resource:     manifest.Resource{Type: "go", Name: "golang.org/x/tools/cmd/stringer"},
			expectedCmd:  "go",
			expectedArgs: []string{"install", "golang.org/x/tools/cmd/stringer@latest"},
		},
		{
			name:         "git-config",
			resource:     manifest.Resource{Type: "git-config", Name: "user.name", Options: map[string]string{"value": "John"}},
			expectedCmd:  "git",
			expectedArgs: []string{"config", "--global", "user.name", "John"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fr := runner.NewFakeRunner()
			step := planner.Step{
				Action:   planner.ActionInstall,
				Resource: tc.resource,
			}
			err := installer.ExecuteStep(context.Background(), fr, step)
			if err != nil {
				t.Fatalf("ExecuteStep failed: %v", err)
			}

			if len(fr.Calls) != 1 {
				t.Fatalf("expected 1 runner call, got %d", len(fr.Calls))
			}

			call := fr.Calls[0]
			if call.Name != tc.expectedCmd {
				t.Errorf("expected command %s, got %s", tc.expectedCmd, call.Name)
			}

			if len(call.Args) != len(tc.expectedArgs) {
				t.Fatalf("expected args len %d, got %d", len(tc.expectedArgs), len(call.Args))
			}

			for i, arg := range call.Args {
				if arg != tc.expectedArgs[i] {
					t.Errorf("arg %d: expected %s, got %s", i, tc.expectedArgs[i], arg)
				}
			}
		})
	}
}

func TestExecuteStep_Remove(t *testing.T) {
	tests := []struct {
		name         string
		resource     manifest.Resource
		expectedCmd  string
		expectedArgs []string
	}{
		{
			name:         "vscode-extension",
			resource:     manifest.Resource{Type: "vscode-extension", Name: "golang.Go"},
			expectedCmd:  "code",
			expectedArgs: []string{"--uninstall-extension", "golang.Go"},
		},
		{
			name:         "cursor-extension",
			resource:     manifest.Resource{Type: "cursor-extension", Name: "jupyter"},
			expectedCmd:  "cursor",
			expectedArgs: []string{"--uninstall-extension", "jupyter"},
		},
		{
			name:         "npm",
			resource:     manifest.Resource{Type: "npm", Name: "typescript"},
			expectedCmd:  "npm",
			expectedArgs: []string{"uninstall", "-g", "typescript"},
		},
		{
			name:         "pnpm",
			resource:     manifest.Resource{Type: "pnpm", Name: "sass"},
			expectedCmd:  "pnpm",
			expectedArgs: []string{"remove", "-g", "sass"},
		},
		{
			name:         "cargo",
			resource:     manifest.Resource{Type: "cargo", Name: "ripgrep"},
			expectedCmd:  "cargo",
			expectedArgs: []string{"uninstall", "ripgrep"},
		},
		{
			name:         "pipx",
			resource:     manifest.Resource{Type: "pipx", Name: "black"},
			expectedCmd:  "pipx",
			expectedArgs: []string{"uninstall", "black"},
		},
		{
			name:         "uv",
			resource:     manifest.Resource{Type: "uv", Name: "ruff"},
			expectedCmd:  "uv",
			expectedArgs: []string{"tool", "uninstall", "ruff"},
		},
		{
			name:         "git-config",
			resource:     manifest.Resource{Type: "git-config", Name: "user.name"},
			expectedCmd:  "git",
			expectedArgs: []string{"config", "--global", "--unset", "user.name"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fr := runner.NewFakeRunner()
			step := planner.Step{
				Action:   planner.ActionRemove,
				Resource: tc.resource,
			}
			err := installer.ExecuteStep(context.Background(), fr, step)
			if err != nil {
				t.Fatalf("ExecuteStep failed: %v", err)
			}

			if len(fr.Calls) != 1 {
				t.Fatalf("expected 1 runner call, got %d", len(fr.Calls))
			}

			call := fr.Calls[0]
			if call.Name != tc.expectedCmd {
				t.Errorf("expected command %s, got %s", tc.expectedCmd, call.Name)
			}

			if len(call.Args) != len(tc.expectedArgs) {
				t.Fatalf("expected args len %d, got %d", len(tc.expectedArgs), len(call.Args))
			}

			for i, arg := range call.Args {
				if arg != tc.expectedArgs[i] {
					t.Errorf("arg %d: expected %s, got %s", i, tc.expectedArgs[i], arg)
				}
			}
		})
	}
}
