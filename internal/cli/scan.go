package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"setupper/internal/paths"
	"setupper/internal/runner"
	"setupper/internal/scanner"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan the system to produce an observed manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Scanning system...")
		r := runner.NewSubprocessRunner()
		s := scanner.New(r)
		
		obs, err := s.Scan(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to scan: %w", err)
		}

		data, err := yaml.Marshal(obs)
		if err != nil {
			return fmt.Errorf("failed to marshal manifest: %w", err)
		}

		baseDir, err := paths.GetBaseDir()
		if err != nil {
			return err
		}

		obsPath := filepath.Join(baseDir, "cache", "observed.yaml")
		if err := os.WriteFile(obsPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write observed manifest: %w", err)
		}

		fmt.Printf("Scan complete! Found %d resources.\n", len(obs.Resources))
		fmt.Printf("Observed manifest written to %s\n", obsPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
