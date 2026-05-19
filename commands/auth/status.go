package auth

import (
	"context"
	"fmt"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Args:  cobra.NoArgs,
		RunE:  runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}
	if cfg.APIKey == "" {
		fmt.Fprintln(cmd.OutOrStdout(), "Not authenticated. Run `declaw auth login` to set your API key.")
		return nil
	}

	acct := declaw.NewAccountClient(cmdutil.AccountOpts(cfg)...)
	defer acct.Close()

	info, err := acct.GetAccount(context.Background())
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(map[string]interface{}{
			"authenticated": true,
			"email":         info.Email,
			"tier":          info.Tier,
			"owner_id":      info.OwnerID,
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Authenticated as %s\n", info.Email)
	fmt.Fprintf(cmd.OutOrStdout(), "Tier: %s\n", info.Tier)
	fmt.Fprintf(cmd.OutOrStdout(), "Owner ID: %s\n", info.OwnerID)
	return nil
}
