package account

import (
	"context"
	"fmt"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newAPIKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api-keys",
		Short: "Manage API keys",
	}
	cmd.AddCommand(
		newAPIKeysListCmd(),
		newAPIKeysCreateCmd(),
		newAPIKeysRevokeCmd(),
	)
	return cmd
}

func newAPIKeysListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List API keys",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE:    runAPIKeysList,
	}
}

func runAPIKeysList(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	acct := declaw.NewAccountClient(cmdutil.AccountOpts(cfg)...)
	defer acct.Close()

	keys, err := acct.ListAPIKeys(context.Background())
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(keys)
	}

	if len(keys) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No API keys found.")
		return nil
	}

	headers := []string{"ID", "Name", "Key", "Created", "Revoked"}
	rows := make([][]string, len(keys))
	for i, k := range keys {
		revoked := ""
		if k.Revoked {
			revoked = "yes"
		}
		rows[i] = []string{
			k.KeyID,
			k.Name,
			k.MaskedKey,
			k.CreatedAt.Format("2006-01-02"),
			revoked,
		}
	}
	p.PrintTable(headers, rows)
	return nil
}

func newAPIKeysCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new API key",
		Args:  cobra.ExactArgs(1),
		RunE:  runAPIKeysCreate,
	}
}

func runAPIKeysCreate(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	acct := declaw.NewAccountClient(cmdutil.AccountOpts(cfg)...)
	defer acct.Close()

	result, err := acct.CreateAPIKey(context.Background(), args[0])
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(result)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created API key: %s\n", result.Name)
	fmt.Fprintf(cmd.OutOrStdout(), "Key ID:          %s\n", result.KeyID)
	fmt.Fprintf(cmd.OutOrStdout(), "API Key:         %s\n", result.APIKey)
	fmt.Fprintln(cmd.OutOrStdout(), "\nSave this key now — it will not be shown again.")
	return nil
}

func newAPIKeysRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <key-id>",
		Short: "Revoke an API key",
		Args:  cobra.ExactArgs(1),
		RunE:  runAPIKeysRevoke,
	}
}

func runAPIKeysRevoke(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	acct := declaw.NewAccountClient(cmdutil.AccountOpts(cfg)...)
	defer acct.Close()

	if err := acct.RevokeAPIKey(context.Background(), args[0]); err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Revoked API key %s\n", args[0])
	return nil
}
