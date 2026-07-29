package cli

import (
	"fmt"
	"path/filepath"
	"sort"

	"setupper/internal/manifest"
	"setupper/internal/paths"
	"github.com/spf13/cobra"
)

var inventoryCmd = &cobra.Command{
	Use:   "inventory",
	Short: "Display a read-only view of the observed state",
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir, err := paths.GetBaseDir()
		if err != nil {
			return err
		}

		obsPath := filepath.Join(baseDir, "cache", "observed.yaml")
		obsPtr, err := manifest.LoadObserved(obsPath)
		if err != nil {
			return fmt.Errorf("failed to load observed manifest (run 'setupper scan' first): %w", err)
		}
		obs := *obsPtr

		grouped := make(map[string][]manifest.Resource)
		for _, res := range obs.Resources {
			grouped[res.Type] = append(grouped[res.Type], res)
		}

		var types []string
		for t := range grouped {
			types = append(types, t)
		}
		sort.Strings(types)

		fmt.Println("--- System Inventory ---")
		for _, t := range types {
			fmt.Printf("\n[%s]\n", t)
			resources := grouped[t]
			
			sort.Slice(resources, func(i, j int) bool {
				return resources[i].Name < resources[j].Name
			})

			for _, res := range resources {
				if t == "mas" {
					fmt.Printf("  - %s (id: %s)\n", res.Name, res.Options["id"])
				} else {
					fmt.Printf("  - %s\n", res.Name)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(inventoryCmd)
}
