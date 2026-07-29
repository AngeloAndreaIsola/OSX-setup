package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"setupper/internal/configurator"
	"setupper/internal/installer"
	"setupper/internal/paths"
	"setupper/internal/planner"
	"setupper/internal/runner"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var failFast bool

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Execute the plan to reach desired state",
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir, err := paths.GetBaseDir()
		if err != nil {
			return err
		}

		planPath := filepath.Join(baseDir, "cache", "plan.yaml")
		data, err := os.ReadFile(planPath)
		if err != nil {
			return fmt.Errorf("failed to read execution plan (run 'setupper plan' first): %w", err)
		}

		var plan planner.ExecutionPlan
		if err := yaml.Unmarshal(data, &plan); err != nil {
			return fmt.Errorf("failed to parse plan: %w", err)
		}

		if len(plan.Steps) == 0 {
			fmt.Println("No steps to execute. Your system matches the desired state.")
			return nil
		}

		r := runner.NewSubprocessRunner()
		opts := installer.ApplyOptions{FailFast: failFast}
		
		results, err := installer.RunApplyTUI(&plan, r, opts)
		if err != nil {
			return fmt.Errorf("apply failed: %w", err)
		}

		failures := 0
		for _, res := range results {
			if res.Error != nil {
				failures++
			}
		}

		configPath := filepath.Join(os.Getenv("HOME"), ".zshrc")
		if err := configurator.ConfigureShell(configPath); err != nil {
			fmt.Printf("Warning: failed to configure shell: %v\n", err)
		}

		if failures > 0 {
			return fmt.Errorf("apply completed with %d failures", failures)
		}

		return nil
	},
}

func init() {
	applyCmd.Flags().BoolVar(&failFast, "fail-fast", false, "Stop execution immediately on first failure")
	rootCmd.AddCommand(applyCmd)
}
