package cli

import (
	"fmt"
	"path/filepath"

	"setupper/internal/diff"
	"setupper/internal/manifest"
	"setupper/internal/paths"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show differences between desired and observed state",
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir, err := paths.GetBaseDir()
		if err != nil {
			return err
		}

		obsPath := filepath.Join(baseDir, "cache", "observed.yaml")
		desPath := filepath.Join(baseDir, "config", "desired.yaml")

		obs, err := manifest.LoadObserved(obsPath)
		if err != nil {
			return fmt.Errorf("failed to load observed manifest (run 'setupper scan' first): %w", err)
		}

		des, err := manifest.LoadDesired(desPath)
		if err != nil {
			return fmt.Errorf("failed to load desired manifest (run 'setupper init' first): %w", err)
		}

		result := diff.Compare(obs, des)

		fmt.Printf("--- Diff Results ---\n\n")
		
		fmt.Printf("Missing (in desired, but not observed): %d\n", len(result.Missing))
		for _, res := range result.Missing {
			fmt.Printf("  + %s (%s)\n", res.Name, res.Type)
		}

		fmt.Printf("\nUnmanaged (in observed, but not desired): %d\n", len(result.Unmanaged))
		for _, res := range result.Unmanaged {
			fmt.Printf("  - %s (%s)\n", res.Name, res.Type)
		}

		fmt.Printf("\nMatching (in both): %d\n", len(result.Matching))
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
