package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/config"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Declaw",
		Long:  "Save an API key for future CLI commands. Pass --api-key or enter interactively.",
		Args:  cobra.NoArgs,
		RunE:  runLogin,
	}
}

func runLogin(cmd *cobra.Command, args []string) error {
	key, _ := cmd.Flags().GetString("api-key")

	if key == "" {
		fmt.Fprintln(os.Stderr, "Don't have an API key? Sign up at https://console.declaw.ai")
		fmt.Fprintln(os.Stderr)
		fmt.Fprint(os.Stderr, "Enter your Declaw API key: ")
		raw, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return fmt.Errorf("reading API key: %w", err)
		}
		key = string(raw)
	}

	if key == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	acct := declaw.NewAccountClient(declaw.WithAPIKey(key))
	defer acct.Close()

	info, err := acct.GetAccount(context.Background())
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cfg.APIKey = key
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Logged in as %s (tier: %s)\n", info.Email, info.Tier)
	fmt.Fprintf(cmd.OutOrStdout(), "Credentials saved to %s\n", config.Path())
	return nil
}
