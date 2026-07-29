package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"setupper/internal/paths"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Upgrade the desired manifest to the current schema version",
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir, err := paths.GetBaseDir()
		if err != nil {
			return err
		}

		desiredPath := filepath.Join(baseDir, "config", "desired.yaml")
		data, err := os.ReadFile(desiredPath)
		if err != nil {
			return fmt.Errorf("could not read desired manifest: %w", err)
		}

		var raw struct {
			SchemaVersion string `yaml:"schema_version"`
		}
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("failed to parse manifest: %w", err)
		}

		currentVersion := "1.0"
		if raw.SchemaVersion == currentVersion {
			fmt.Println("Manifest is already up-to-date with schema version", currentVersion)
			return nil
		}

		fmt.Printf("Migrating from %s to %s (stub - no changes needed yet)\n", raw.SchemaVersion, currentVersion)
		
		fmt.Println("Migration complete!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
