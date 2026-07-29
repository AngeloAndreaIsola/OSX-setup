package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"setupper/internal/manifest"
	"setupper/internal/paths"
	"setupper/internal/profiles"
	"setupper/internal/recommender"
	"github.com/spf13/cobra"
)

var recommendCmd = &cobra.Command{
	Use:   "recommend",
	Short: "Suggest profiles based on your installed tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir, err := paths.GetBaseDir()
		if err != nil {
			return err
		}

		obsPath := filepath.Join(baseDir, "cache", "observed.yaml")
		obs, err := manifest.LoadObserved(obsPath)
		if err != nil {
			return fmt.Errorf("failed to load observed manifest (run 'setupper scan' first): %w", err)
		}

		profs, err := profiles.LoadDefaults()
		if err != nil {
			return fmt.Errorf("failed to load default profiles: %w", err)
		}

		recEngine := recommender.New(profs)
		recommendations := recEngine.Recommend(obs)

		if len(recommendations) == 0 {
			fmt.Println("No recommendations found based on your current inventory.")
			return nil
		}

		fmt.Printf("--- Recommendations ---\n\n")
		for _, rec := range recommendations {
			fmt.Printf("📦 Profile: %s (%s)\n", rec.Profile.Name, rec.Profile.ID)
			fmt.Printf("   Description: %s\n", rec.Profile.Description)
			fmt.Printf("   Triggered by: %s\n", strings.Join(rec.TriggeredBy, ", "))
			fmt.Println("   Suggested Tools:")
			
			for _, res := range rec.Profile.Resources {
				key := manifest.FormatKey(res.Type, res.Name)
				if _, ok := obs.Resources[key]; ok {
					fmt.Printf("    ✓ %s (%s) - already installed\n", res.Name, res.Type)
				} else {
					fmt.Printf("    + %s (%s)\n", res.Name, res.Type)
				}
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(recommendCmd)
}
