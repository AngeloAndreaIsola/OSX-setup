package cli

import (
	"fmt"
	"os"

	"github.com/angeloandreaisola/OSX-setup/internal/paths"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "setupper",
	Short: "Setupper is a declarative macOS configuration manager",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return paths.EnsureDirs()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
