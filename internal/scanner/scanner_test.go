package scanner_test

import (
	"context"
	"testing"

	"setupper/internal/runner"
	"setupper/internal/scanner"
)

func TestScanner(t *testing.T) {
	fr := runner.NewFakeRunner()

	// Configure fake runner outputs
	fr.Results["brew"] = []byte("git\ngo\n") // note: used for both leaves and cask list in runner.FakeRunner since it keys by binary name
	fr.Results["mas"] = []byte("497799835 Xcode (14.2)\n")
	fr.Results["code"] = []byte("golang.Go\nms-python.python\n")
	fr.Results["cursor"] = []byte("ms-toolsai.jupyter\n")
	fr.Results["npm"] = []byte(`{"dependencies": {"typescript": {"version": "5.3.3"}}}`)
	fr.Results["pnpm"] = []byte(`[{"dependencies": {"sass": {"version": "1.70.0"}}}]`)
	fr.Results["cargo"] = []byte("ripgrep v14.1.0:\n    binary: rg\n")
	fr.Results["pipx"] = []byte(`{"venvs": {"black": {}}}`)
	fr.Results["uv"] = []byte("ruff v0.3.0\n")
	fr.Results["go"] = []byte("golang.org/x/tools/cmd/stringer\n\tpath\tgolang.org/x/tools/cmd/stringer\n")
	fr.Results["git"] = []byte("user.name=Test User\nuser.email=test@example.com\n")

	s := scanner.New(fr)
	obs, err := s.Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	expected := []struct {
		key  string
		name string
	}{
		{"brew:git", "git"},
		{"brew:go", "go"},
		{"cask:git", "git"}, // due to fake runner returning same output for brew
		{"mas:497799835", "Xcode"},
		{"vscode-extension:golang.Go", "golang.Go"},
		{"cursor-extension:ms-toolsai.jupyter", "ms-toolsai.jupyter"},
		{"npm:typescript", "typescript"},
		{"pnpm:sass", "sass"},
		{"cargo:ripgrep", "ripgrep"},
		{"pipx:black", "black"},
		{"uv:ruff", "ruff"},
		{"git-config:user.name", "user.name"},
		{"git-config:user.email", "test@example.com"}, // parsed as user.email=test@example.com
	}

	for _, tc := range expected {
		res, ok := obs.Resources[tc.key]
		if !ok {
			t.Errorf("expected key %s was not found in scan", tc.key)
			continue
		}
		if tc.key == "git-config:user.email" {
			if val, exists := res.Options["value"]; !exists || val != "test@example.com" {
				t.Errorf("expected git-config user.email to be test@example.com, got %v", val)
			}
		}
	}
}
