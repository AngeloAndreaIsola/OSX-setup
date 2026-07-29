package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"setupper/internal/manifest"
	"setupper/internal/paths"
	"setupper/internal/runner"
	"setupper/internal/scanner"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a starter desired manifest from current system state",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Scanning system to generate starter manifest...")
		
		r := runner.NewSubprocessRunner()
		s := scanner.New(r)
		obs, err := s.Scan(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to scan: %w", err)
		}

		desired := &manifest.DesiredManifest{
			SchemaVersion: "1.0",
			Resources:     make(map[string]manifest.Resource),
		}

		for key, res := range obs.Resources {
			desired.Resources[key] = res
		}

		baseDir, err := paths.GetBaseDir()
		if err != nil {
			return err
		}

		desiredPath := filepath.Join(baseDir, "config", "desired.yaml")
		
		if _, err := os.Stat(desiredPath); err == nil {
			return fmt.Errorf("desired manifest already exists at %s", desiredPath)
		}

		if err := manifest.SaveDesired(desiredPath, desired); err != nil {
			return fmt.Errorf("failed to save desired manifest: %w", err)
		}

		fmt.Printf("Starter desired manifest created at %s\n", desiredPath)
		fmt.Println("Please review and edit it before running apply.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
