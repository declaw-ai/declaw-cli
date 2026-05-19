package volume

import (
	"context"
	"fmt"
	"os"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new volume",
		Args:  cobra.ExactArgs(1),
		RunE:  runCreate,
	}
	cmd.Flags().String("from-tar", "", "Path to a gzipped tar archive for initial data")
	return cmd
}

func runCreate(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	var data []byte
	if tarPath, _ := cmd.Flags().GetString("from-tar"); tarPath != "" {
		data, err = os.ReadFile(tarPath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", tarPath, err)
		}
	}

	vol, err := declaw.CreateVolume(context.Background(), args[0], data, cmdutil.SandboxOpts(cfg)...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(vol)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created volume %s (ID: %s, %d bytes)\n", vol.Name, vol.VolumeID, vol.SizeBytes)
	return nil
}
