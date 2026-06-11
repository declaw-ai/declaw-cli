package volume

import (
	"context"
	"fmt"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newEmptyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "empty",
		Short: "Create a new empty file-granular volume",
		Args:  cobra.NoArgs,
		RunE:  runEmpty,
	}
	cmd.Flags().String("name", "", "Name for the new volume (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func runEmpty(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	name, _ := cmd.Flags().GetString("name")

	vol, err := declaw.CreateEmptyVolume(context.Background(), name, cmdutil.SandboxOpts(cfg)...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(vol)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created empty volume %s (ID: %s)\n", vol.Name, vol.VolumeID)
	return nil
}
