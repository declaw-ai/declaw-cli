package volume

import (
	"context"
	"fmt"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <volume-id>",
		Short: "Show volume details",
		Args:  cobra.ExactArgs(1),
		RunE:  runGet,
	}
}

func runGet(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	vol, err := declaw.GetVolume(context.Background(), args[0], cmdutil.SandboxOpts(cfg)...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(vol)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "ID:      %s\n", vol.VolumeID)
	fmt.Fprintf(cmd.OutOrStdout(), "Name:    %s\n", vol.Name)
	fmt.Fprintf(cmd.OutOrStdout(), "Size:    %s\n", formatBytes(vol.SizeBytes))
	fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", vol.CreatedAt)
	if len(vol.Metadata) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Metadata:")
		for k, v := range vol.Metadata {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s: %s\n", k, v)
		}
	}
	return nil
}
