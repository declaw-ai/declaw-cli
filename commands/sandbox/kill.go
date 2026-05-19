package sandbox

import (
	"context"
	"fmt"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newKillCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "kill <sandbox-id> [sandbox-id...]",
		Short: "Kill one or more sandboxes",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runKill,
	}
}

func runKill(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	opts := cmdutil.SandboxOpts(cfg)

	if len(args) == 1 {
		if err := declaw.KillSandbox(context.Background(), args[0], opts...); err != nil {
			cmdutil.HandleError(err)
			return nil
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Killed sandbox %s\n", args[0])
		return nil
	}

	results, err := declaw.KillManySandboxes(context.Background(), args, opts...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	for _, r := range results {
		if r.Error != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to kill %s: %s\n", r.SandboxID, r.Error)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Killed sandbox %s\n", r.SandboxID)
		}
	}
	return nil
}
