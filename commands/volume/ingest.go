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

func newIngestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Create a new file-granular volume from a gzipped tar archive",
		Args:  cobra.NoArgs,
		RunE:  runIngest,
	}
	cmd.Flags().String("name", "", "Name for the new volume (required)")
	cmd.Flags().String("file", "", "Path to a gzipped tar archive (.tar.gz) (required)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func runIngest(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	name, _ := cmd.Flags().GetString("name")
	file, _ := cmd.Flags().GetString("file")

	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading %s: %w", file, err)
	}

	vol, err := declaw.IngestVolume(context.Background(), name, data, cmdutil.SandboxOpts(cfg)...)
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
