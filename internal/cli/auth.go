package cli

import (
	"fmt"
	"path/filepath"

	"setupper/internal/auth"
	"setupper/internal/manifest"
	"setupper/internal/paths"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate configured tools interactively (e.g. gh, aws)",
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

		changed := false

		for key, res := range des.Resources {
			authenticator := auth.GetAuthenticator(res.Name)
			if authenticator == nil {
				continue
			}

			fmt.Printf("\nChecking auth status for %s...\n", res.Name)
			isAuthenticated, account, _ := authenticator.Check(cmd.Context())

			if isAuthenticated {
				fmt.Printf("✅ Already authenticated as: %s\n", account)
				if !res.Authenticated || res.Account != account {
					res.Authenticated = true
					res.Account = account
					des.Resources[key] = res
					changed = true
				}
				continue
			}

			fmt.Printf("🔑 Authentication required for %s. Starting login flow...\n", res.Name)
			if err := authenticator.Login(cmd.Context()); err != nil {
				fmt.Printf("❌ Login failed for %s: %v\n", res.Name, err)
				continue
			}

			isAuthenticated, account, _ = authenticator.Check(cmd.Context())
			if isAuthenticated {
				fmt.Printf("✅ Successfully authenticated as: %s\n", account)
				res.Authenticated = true
				res.Account = account
				des.Resources[key] = res
				changed = true
			}
		}

		if changed {
			if err := manifest.SaveDesired(desPath, des); err != nil {
				return fmt.Errorf("failed to save updated manifest: %w", err)
			}
			fmt.Println("\nUpdated desired manifest with authentication records (no secrets stored).")
		} else {
			fmt.Println("\nNo changes made.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}
