package account

import (
	"context"
	"fmt"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newUsageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "usage",
		Short: "Show account usage summary",
		Args:  cobra.NoArgs,
		RunE:  runUsage,
	}
}

func runUsage(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	acct := declaw.NewAccountClient(cmdutil.AccountOpts(cfg)...)
	defer acct.Close()

	usage, err := acct.GetUsage(context.Background())
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(usage)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Usage since: %s\n", usage.Since.Format("2006-01-02"))
	fmt.Fprintf(cmd.OutOrStdout(), "Sandboxes:   %d\n", usage.SandboxCount)
	fmt.Fprintf(cmd.OutOrStdout(), "Total time:  %.0fs\n", usage.TotalSeconds)
	fmt.Fprintf(cmd.OutOrStdout(), "Total cost:  %s\n", usage.TotalCostUSD)
	fmt.Fprintf(cmd.OutOrStdout(), "\nBalances:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  Sandbox:     $%.2f remaining\n", float64(usage.SandboxBalanceRemainingMicros)/1e6)
	fmt.Fprintf(cmd.OutOrStdout(), "  Guardrails:  $%.2f remaining\n", float64(usage.GuardrailsBalanceRemainingMicros)/1e6)
	fmt.Fprintf(cmd.OutOrStdout(), "  Paid:        $%.2f remaining\n", float64(usage.BalanceRemainingMicros)/1e6)
	return nil
}
