package exporter_test

import (
	"bytes"
	"strings"
	"testing"

	"setupper/internal/exporter"
	"setupper/internal/manifest"
)

func TestShellScriptExporter(t *testing.T) {
	exp, err := exporter.Get("shell")
	if err != nil {
		t.Fatalf("failed to get shell exporter: %v", err)
	}

	des := &manifest.DesiredManifest{
		Resources: map[string]manifest.Resource{
			"brew:go":     {Type: "brew", Name: "go"},
			"brew:git":    {Type: "brew", Name: "git"},
			"cask:vscode": {Type: "cask", Name: "visual-studio-code"},
			"mas:xcode":   {Type: "mas", Name: "Xcode", Options: map[string]string{"id": "497799835"}},
		},
	}

	var buf bytes.Buffer
	if err := exp.Export(des, &buf); err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	output := buf.String()

	// Check shebang
	if !strings.HasPrefix(output, "#!/bin/bash") {
		t.Errorf("expected shebang at prefix, got: %s", output)
	}

	// Check sorted formulas
	gitIdx := strings.Index(output, `brew list --formula "git"`)
	goIdx := strings.Index(output, `brew list --formula "go"`)
	if gitIdx == -1 || goIdx == -1 {
		t.Error("missing git or go formulas in output")
	}
	if gitIdx > goIdx {
		t.Error("formulas are not sorted alphabetically")
	}

	// Check casks
	if !strings.Contains(output, `brew list --cask "visual-studio-code"`) {
		t.Error("missing vscode cask in output")
	}

	// Check mas apps
	if !strings.Contains(output, "mas install 497799835") {
		t.Error("missing xcode mas install in output")
	}
}

func TestBrewfileExporter(t *testing.T) {
	exp, err := exporter.Get("brewfile")
	if err != nil {
		t.Fatalf("failed to get brewfile exporter: %v", err)
	}

	des := &manifest.DesiredManifest{
		Resources: map[string]manifest.Resource{
			"brew:go":     {Type: "brew", Name: "go"},
			"brew:git":    {Type: "brew", Name: "git"},
			"cask:vscode": {Type: "cask", Name: "visual-studio-code"},
			"mas:xcode":   {Type: "mas", Name: "Xcode", Options: map[string]string{"id": "497799835"}},
		},
	}

	var buf bytes.Buffer
	if err := exp.Export(des, &buf); err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	output := buf.String()

	// Check formulas
	if !strings.Contains(output, `brew "git"`) || !strings.Contains(output, `brew "go"`) {
		t.Error("missing git or go in Brewfile output")
	}

	// Check casks
	if !strings.Contains(output, `cask "visual-studio-code"`) {
		t.Error("missing vscode in Brewfile output")
	}

	// Check mas
	if !strings.Contains(output, `mas "Xcode", id: 497799835`) {
		t.Error("missing xcode in Brewfile output")
	}
}

func TestGetUnknownExporter(t *testing.T) {
	_, err := exporter.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent exporter, got nil")
	}
}
