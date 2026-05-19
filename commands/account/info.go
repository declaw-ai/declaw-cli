package account

import (
	"context"
	"fmt"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show account information",
		Args:  cobra.NoArgs,
		RunE:  runInfo,
	}
}

func runInfo(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	acct := declaw.NewAccountClient(cmdutil.AccountOpts(cfg)...)
	defer acct.Close()

	overview, err := acct.GetOverview(context.Background())
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(overview)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Owner ID:          %s\n", overview.OwnerID)
	fmt.Fprintf(cmd.OutOrStdout(), "Tier:              %s\n", overview.Tier)
	fmt.Fprintf(cmd.OutOrStdout(), "Active Sandboxes:  %d\n", overview.ActiveSandboxes)
	fmt.Fprintf(cmd.OutOrStdout(), "\nLimits:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  Max Concurrent:  %d\n", overview.TierLimits.MaxConcurrent)
	fmt.Fprintf(cmd.OutOrStdout(), "  Max vCPUs:       %d\n", overview.TierLimits.MaxVCPUs)
	fmt.Fprintf(cmd.OutOrStdout(), "  Max Memory:      %d MB\n", overview.TierLimits.MaxMemoryMB)
	fmt.Fprintf(cmd.OutOrStdout(), "\nWallets:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  Sandbox Free:    $%.2f\n", float64(overview.Wallets.SandboxFreeMicros)/1e6)
	fmt.Fprintf(cmd.OutOrStdout(), "  Guardrails Free: $%.2f\n", float64(overview.Wallets.GuardrailsFreeMicros)/1e6)
	fmt.Fprintf(cmd.OutOrStdout(), "  Paid Balance:    $%.2f\n", float64(overview.Wallets.BalanceMicros)/1e6)
	fmt.Fprintf(cmd.OutOrStdout(), "\nToday's Spend:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  Compute:         $%.4f\n", float64(overview.Today.ComputeCostMicros)/1e6)
	fmt.Fprintf(cmd.OutOrStdout(), "  Guardrails:      $%.4f\n", float64(overview.Today.GuardrailsCostMicros)/1e6)
	fmt.Fprintf(cmd.OutOrStdout(), "  Total:           $%.4f\n", float64(overview.Today.TotalCostMicros)/1e6)
	return nil
}
