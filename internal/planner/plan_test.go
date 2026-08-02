package planner_test

import (
	"testing"

	"setupper/internal/manifest"
	"setupper/internal/planner"
)

func TestGeneratePlan(t *testing.T) {
	installs := []manifest.Resource{
		{Type: "cask", Name: "slack"},
		{Type: "brew", Name: "git"},
		{Type: "brew-tap", Name: "homebrew/cask-fonts"},
		{Type: "mas", Name: "Xcode"},
	}

	removes := []manifest.Resource{
		{Type: "brew-tap", Name: "homebrew/cask-versions"},
		{Type: "cask", Name: "firefox"},
	}

	plan := planner.GeneratePlan(installs, removes)

	// Check removes ordering (cask firefox then brew-tap homebrew/cask-versions)
	if len(plan.Steps) != 6 {
		t.Fatalf("expected 6 steps, got %d", len(plan.Steps))
	}

	// First two steps should be removes
	if plan.Steps[0].Action != planner.ActionRemove || plan.Steps[0].Resource.Type != "cask" {
		t.Errorf("step 0 should be Remove cask, got %s %s", plan.Steps[0].Action, plan.Steps[0].Resource.Type)
	}
	if plan.Steps[1].Action != planner.ActionRemove || plan.Steps[1].Resource.Type != "brew-tap" {
		t.Errorf("step 1 should be Remove brew-tap, got %s %s", plan.Steps[1].Action, plan.Steps[1].Resource.Type)
	}

	// Last four steps should be installs ordered by: brew-tap -> brew -> cask -> mas
	expectedInstalls := []string{"brew-tap", "brew", "cask", "mas"}
	for i, expectedType := range expectedInstalls {
		stepIdx := i + 2
		if plan.Steps[stepIdx].Action != planner.ActionInstall {
			t.Errorf("step %d should be Install, got %s", stepIdx, plan.Steps[stepIdx].Action)
		}
		if plan.Steps[stepIdx].Resource.Type != expectedType {
			t.Errorf("step %d should be of type %s, got %s", stepIdx, expectedType, plan.Steps[stepIdx].Resource.Type)
		}
	}
}
