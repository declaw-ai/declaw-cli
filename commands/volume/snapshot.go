package volume

import (
	"context"
	"fmt"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot <sandbox-id>",
		Short: "Capture an in-sandbox path into a new volume",
		Args:  cobra.ExactArgs(1),
		RunE:  runSnapshot,
	}
	cmd.Flags().String("path", "", "Absolute in-sandbox path to capture (required)")
	cmd.Flags().String("name", "", "Name for the new volume (default: snapshot)")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func runSnapshot(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	path, _ := cmd.Flags().GetString("path")
	name, _ := cmd.Flags().GetString("name")

	vol, err := declaw.SnapshotVolume(context.Background(), args[0], path, name, cmdutil.SandboxOpts(cfg)...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(vol)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created volume %s (ID: %s, %d bytes) from %s\n", vol.Name, vol.VolumeID, vol.SizeBytes, path)
	return nil
}
