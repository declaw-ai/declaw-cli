package template

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
		Use:   "info <template-id>",
		Short: "Show template details",
		Args:  cobra.ExactArgs(1),
		RunE:  runInfo,
	}
}

func runInfo(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	info, err := declaw.GetTemplate(context.Background(), args[0], cmdutil.SandboxOpts(cfg)...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(info)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "ID:      %s\n", info.TemplateID)
	fmt.Fprintf(cmd.OutOrStdout(), "Alias:   %s\n", info.Alias)
	fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", info.CreatedAt)
	return nil
}
