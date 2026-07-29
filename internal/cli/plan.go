package cli

import (
	"fmt"
	"path/filepath"

	"setupper/internal/diff"
	"setupper/internal/manifest"
	"setupper/internal/paths"
	"setupper/internal/planner"
	"github.com/spf13/cobra"
)

var yesFlag bool

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Generate and review an execution plan to reach desired state",
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir, err := paths.GetBaseDir()
		if err != nil {
			return err
		}

		obs, err := manifest.LoadObserved(filepath.Join(baseDir, "cache", "observed.yaml"))
		if err != nil {
			return fmt.Errorf("failed to load observed manifest: %w", err)
		}

		des, err := manifest.LoadDesired(filepath.Join(baseDir, "config", "desired.yaml"))
		if err != nil {
			return fmt.Errorf("failed to load desired manifest: %w", err)
		}

		d := diff.Compare(obs, des)
		
		var installs []manifest.Resource
		var removes []manifest.Resource

		if yesFlag {
			// Non-interactive mode: accept all missing (defaults)
			installs = d.Missing
			// By default, we don't automatically remove unmanaged things unless user interactive selects them.
			removes = nil
		} else {
			// Interactive mode
			var err error
			installs, removes, err = planner.RunInteractiveChecklist(d.Missing, d.Unmanaged)
			if err != nil {
				return err
			}
		}

		plan := planner.GeneratePlan(installs, removes)
		
		planPath := filepath.Join(baseDir, "cache", "plan.yaml")
		if err := planner.SavePlan(planPath, plan); err != nil {
			return fmt.Errorf("failed to save plan: %w", err)
		}

		fmt.Printf("Execution plan saved to %s\n", planPath)
		fmt.Printf("Steps: %d\n", len(plan.Steps))
		return nil
	},
}

func init() {
	planCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip interactive checklist and accept defaults")
	rootCmd.AddCommand(planCmd)
}
