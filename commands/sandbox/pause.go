package sandbox

import (
	"context"
	"fmt"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newPauseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pause <sandbox-id>",
		Short: "Pause a running sandbox",
		Args:  cobra.ExactArgs(1),
		RunE:  runPause,
	}
}

func runPause(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	sbx, err := declaw.Connect(context.Background(), args[0], cmdutil.SandboxOpts(cfg)...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	if err := sbx.Pause(context.Background()); err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Paused sandbox %s\n", args[0])
	return nil
}
