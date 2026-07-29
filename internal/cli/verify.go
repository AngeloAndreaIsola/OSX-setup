package cli

import (
	"fmt"
	"path/filepath"

	"setupper/internal/manifest"
	"setupper/internal/paths"
	"setupper/internal/runner"
	"setupper/internal/scanner"
	"setupper/internal/verifier"
	"github.com/spf13/cobra"
)

var deepCheck bool

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Run health checks against desired manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir, err := paths.GetBaseDir()
		if err != nil {
			return err
		}

		desPath := filepath.Join(baseDir, "config", "desired.yaml")
		des, err := manifest.LoadDesired(desPath)
		if err != nil {
			return fmt.Errorf("failed to load desired manifest (run 'setupper init' first): %w", err)
		}

		r := runner.NewSubprocessRunner()
		s := scanner.New(r)
		v := verifier.New(s)

		results, obs, err := verifier.RunVerifyTUI(v, des, deepCheck)
		if err != nil {
			return fmt.Errorf("verification failed: %w", err)
		}

		if obs != nil {
			obsPath := filepath.Join(baseDir, "cache", "observed.yaml")
			if err := manifest.SaveObserved(obsPath, obs); err != nil {
				return fmt.Errorf("failed to update observed manifest drift surface: %w", err)
			}
		}

		_, warnings, failures := verifier.Summarize(results)

		if failures > 0 {
			return fmt.Errorf("verification completed with %d failures and %d warnings", failures, warnings)
		}
		if warnings > 0 {
			fmt.Printf("Verification passed with %d warnings.\n", warnings)
		}

		return nil
	},
}

func init() {
	verifyCmd.Flags().BoolVar(&deepCheck, "deep", false, "Exercise credentials and services deeply")
	rootCmd.AddCommand(verifyCmd)
}
