package planner

import (
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
	"setupper/internal/manifest"
)

type Action string

const (
	ActionInstall Action = "Install"
	ActionRemove  Action = "Remove"
)

type Step struct {
	Action   Action            `yaml:"action"`
	Resource manifest.Resource `yaml:"resource"`
}

type ExecutionPlan struct {
	Steps []Step `yaml:"steps"`
}

// GeneratePlan takes selected resources and creates an ordered plan.
// Hardcoded type-level dependency rules:
// - brew (formulas) first
// - cask (apps) second
// - mas (app store) third
func GeneratePlan(installs []manifest.Resource, removes []manifest.Resource) *ExecutionPlan {
	var steps []Step
	
	// Add removes (reverse dependency order safely: mas -> cask -> brew)
	sort.Slice(removes, func(i, j int) bool {
		return typeOrder(removes[i].Type) > typeOrder(removes[j].Type)
	})
	for _, r := range removes {
		steps = append(steps, Step{Action: ActionRemove, Resource: r})
	}

	// Add installs (ordered by dependency: brew -> cask -> mas)
	sort.Slice(installs, func(i, j int) bool {
		if typeOrder(installs[i].Type) == typeOrder(installs[j].Type) {
			return installs[i].Name < installs[j].Name
		}
		return typeOrder(installs[i].Type) < typeOrder(installs[j].Type)
	})
	for _, r := range installs {
		steps = append(steps, Step{Action: ActionInstall, Resource: r})
	}

	return &ExecutionPlan{Steps: steps}
}

func typeOrder(t string) int {
	switch t {
	case "brew":
		return 1
	case "cask":
		return 2
	case "mas":
		return 3
	default:
		return 99
	}
}

func SavePlan(path string, plan *ExecutionPlan) error {
	data, err := yaml.Marshal(plan)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
