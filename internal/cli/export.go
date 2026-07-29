package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"setupper/internal/exporter"
	"setupper/internal/manifest"
	"setupper/internal/paths"
	"github.com/spf13/cobra"
)

var (
	exportFormat string
	exportOutput string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the desired manifest to a bootstrap shell script or Brewfile",
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

		exp, err := exporter.Get(exportFormat)
		if err != nil {
			return err
		}

		var out *os.File
		if exportOutput != "" {
			out, err = os.Create(exportOutput)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer out.Close()
		} else {
			out = os.Stdout
		}

		if err := exp.Export(des, out); err != nil {
			return fmt.Errorf("failed to export: %w", err)
		}

		if exportOutput != "" {
			fmt.Printf("Successfully exported to %s using %s format.\n", exportOutput, exportFormat)
		}

		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "shell", "Export format (shell, brewfile)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file path (default stdout)")
	rootCmd.AddCommand(exportCmd)
}
