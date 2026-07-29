package installer

import (
	"context"
	"fmt"
	"setupper/internal/manifest"
	"setupper/internal/planner"
	"setupper/internal/runner"
)

type Result struct {
	Step  planner.Step
	Error error
}

type ApplyOptions struct {
	FailFast bool
}

func ExecuteStep(ctx context.Context, r runner.Runner, step planner.Step) error {
	switch step.Action {
	case planner.ActionInstall:
		return runInstall(ctx, r, step.Resource)
	case planner.ActionRemove:
		return runRemove(ctx, r, step.Resource)
	}
	return fmt.Errorf("unknown action: %s", step.Action)
}

func runInstall(ctx context.Context, r runner.Runner, res manifest.Resource) error {
	switch res.Type {
	case "brew":
		_, err := r.Run(ctx, "brew", "install", res.Name)
		return err
	case "cask":
		_, err := r.Run(ctx, "brew", "install", "--cask", res.Name)
		return err
	case "mas":
		id := res.Options["id"]
		if id == "" {
			return fmt.Errorf("missing mas id for %s", res.Name)
		}
		_, err := r.Run(ctx, "mas", "install", id)
		return err
	default:
		return fmt.Errorf("unsupported resource type: %s", res.Type)
	}
}

func runRemove(ctx context.Context, r runner.Runner, res manifest.Resource) error {
	switch res.Type {
	case "brew", "cask":
		_, err := r.Run(ctx, "brew", "uninstall", res.Name)
		return err
	case "mas":
		return fmt.Errorf("automatic removal of Mac App Store apps is not supported")
	default:
		return fmt.Errorf("unsupported resource type: %s", res.Type)
	}
}
