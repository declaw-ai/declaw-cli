package auth

import (
	"fmt"

	"github.com/declaw-ai/declaw-cli/internal/config"
	"github.com/spf13/cobra"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if cfg.APIKey == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Not logged in.")
				return nil
			}
			cfg.APIKey = ""
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Logged out. API key removed from config.")
			return nil
		},
	}
}
