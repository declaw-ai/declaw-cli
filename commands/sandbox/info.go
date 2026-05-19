package sandbox

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
		Use:   "info <sandbox-id>",
		Short: "Show sandbox details",
		Args:  cobra.ExactArgs(1),
		RunE:  runInfo,
	}
}

func runInfo(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	sbx, err := declaw.Connect(context.Background(), args[0], cmdutil.SandboxOpts(cfg)...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	info, err := sbx.GetInfo(context.Background())
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(info)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "ID:        %s\n", info.SandboxID)
	fmt.Fprintf(cmd.OutOrStdout(), "Template:  %s\n", info.TemplateID)
	fmt.Fprintf(cmd.OutOrStdout(), "State:     %s\n", info.State)
	fmt.Fprintf(cmd.OutOrStdout(), "Name:      %s\n", info.Name)
	fmt.Fprintf(cmd.OutOrStdout(), "Started:   %s\n", output.RelativeTime(info.StartedAt))
	if info.EndAt != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "Ends at:   %s\n", info.EndAt.Format("2006-01-02T15:04:05Z"))
	}
	if len(info.Metadata) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Metadata:")
		for k, v := range info.Metadata {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s: %s\n", k, v)
		}
	}
	return nil
}
