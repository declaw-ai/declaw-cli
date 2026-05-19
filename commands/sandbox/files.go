package sandbox

import (
	"context"
	"fmt"
	"os"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newFilesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "files",
		Short: "File operations on a sandbox",
	}
	cmd.AddCommand(
		newFilesLsCmd(),
		newFilesReadCmd(),
		newFilesWriteCmd(),
	)
	return cmd
}

func newFilesLsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls <sandbox-id> <path>",
		Short: "List files in a sandbox directory",
		Args:  cobra.ExactArgs(2),
		RunE:  runFilesLs,
	}
}

func runFilesLs(cmd *cobra.Command, args []string) error {
	sandboxID, path := args[0], args[1]

	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	sbx, err := declaw.Connect(context.Background(), sandboxID, cmdutil.SandboxOpts(cfg)...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	entries, err := sbx.Files.List(context.Background(), path)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(entries)
	}

	if len(entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Empty directory.")
		return nil
	}

	headers := []string{"Path", "Type", "Size"}
	rows := make([][]string, len(entries))
	for i, e := range entries {
		rows[i] = []string{
			e.Path,
			string(e.Type),
			fmt.Sprintf("%d", e.Size),
		}
	}
	p.PrintTable(headers, rows)
	return nil
}

func newFilesReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read <sandbox-id> <path>",
		Short: "Read a file from a sandbox",
		Args:  cobra.ExactArgs(2),
		RunE:  runFilesRead,
	}
	cmd.Flags().StringP("output", "o", "", "Write to local file instead of stdout")
	return cmd
}

func runFilesRead(cmd *cobra.Command, args []string) error {
	sandboxID, path := args[0], args[1]

	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	sbx, err := declaw.Connect(context.Background(), sandboxID, cmdutil.SandboxOpts(cfg)...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	data, err := sbx.Files.Read(context.Background(), path)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	if outFile, _ := cmd.Flags().GetString("output"); outFile != "" {
		if err := os.WriteFile(outFile, []byte(data), 0600); err != nil {
			return fmt.Errorf("writing to %s: %w", outFile, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Written to %s\n", outFile)
		return nil
	}

	fmt.Fprint(cmd.OutOrStdout(), data)
	return nil
}

func newFilesWriteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "write <sandbox-id> <remote-path> <local-path>",
		Short: "Upload a local file to a sandbox",
		Args:  cobra.ExactArgs(3),
		RunE:  runFilesWrite,
	}
}

func runFilesWrite(cmd *cobra.Command, args []string) error {
	sandboxID, remotePath, localPath := args[0], args[1], args[2]

	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", localPath, err)
	}

	sbx, err := declaw.Connect(context.Background(), sandboxID, cmdutil.SandboxOpts(cfg)...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	_, err = sbx.Files.WriteBytes(context.Background(), remotePath, data)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Uploaded %s → %s (%d bytes)\n", localPath, remotePath, len(data))
	return nil
}
