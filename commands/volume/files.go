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

func newFilesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "files",
		Short: "Manage files inside a file-granular volume",
	}
	cmd.AddCommand(
		newFilesReadCmd(),
		newFilesWriteCmd(),
		newFilesListCmd(),
		newFilesInfoCmd(),
		newFilesExistsCmd(),
		newFilesRemoveCmd(),
		newFilesRenameCmd(),
		newFilesMkdirCmd(),
	)
	return cmd
}

// ---------------------------------------------------------------------------
// read
// ---------------------------------------------------------------------------

func newFilesReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read <volume-id>",
		Short: "Read a file from a volume",
		Args:  cobra.ExactArgs(1),
		RunE:  runFilesRead,
	}
	cmd.Flags().String("path", "", "Path inside the volume (required)")
	cmd.Flags().StringP("output", "o", "", "Write file contents to this local path instead of stdout")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func runFilesRead(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}
	path, _ := cmd.Flags().GetString("path")
	out, _ := cmd.Flags().GetString("output")

	data, err := declaw.VolumeFilesFor(args[0], cmdutil.SandboxOpts(cfg)...).Read(context.Background(), path)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	if out != "" {
		if err := os.WriteFile(out, data, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", out, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Wrote %d bytes to %s\n", len(data), out)
		return nil
	}

	_, err = cmd.OutOrStdout().Write(data)
	return err
}

// ---------------------------------------------------------------------------
// write
// ---------------------------------------------------------------------------

func newFilesWriteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "write <volume-id>",
		Short: "Write a file to a volume",
		Args:  cobra.ExactArgs(1),
		RunE:  runFilesWrite,
	}
	cmd.Flags().String("path", "", "Path inside the volume (required)")
	cmd.Flags().String("file", "", "Local file to upload (mutually exclusive with --content)")
	cmd.Flags().String("content", "", "Inline content to write (mutually exclusive with --file)")
	cmd.Flags().String("if-version", "", "Conditional write (CAS): only write if the file's current version matches")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func runFilesWrite(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}
	path, _ := cmd.Flags().GetString("path")
	file, _ := cmd.Flags().GetString("file")
	content, _ := cmd.Flags().GetString("content")

	if file != "" && content != "" {
		return fmt.Errorf("--file and --content are mutually exclusive")
	}

	var data []byte
	if file != "" {
		data, err = os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading %s: %w", file, err)
		}
	} else {
		data = []byte(content)
	}

	var wopts []declaw.WriteFileOption
	if cmd.Flags().Changed("if-version") {
		ver, _ := cmd.Flags().GetString("if-version")
		wopts = append(wopts, declaw.WithIfVersion(ver))
	}

	if err := declaw.VolumeFilesFor(args[0], cmdutil.SandboxOpts(cfg)...).Write(context.Background(), path, data, wopts...); err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Wrote %d bytes to %s\n", len(data), path)
	return nil
}

// ---------------------------------------------------------------------------
// list
// ---------------------------------------------------------------------------

func newFilesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <volume-id>",
		Short:   "List entries of a directory inside a volume",
		Aliases: []string{"ls"},
		Args:    cobra.ExactArgs(1),
		RunE:    runFilesList,
	}
	cmd.Flags().String("path", "/", "Directory path inside the volume")
	return cmd
}

func runFilesList(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}
	path, _ := cmd.Flags().GetString("path")

	entries, err := declaw.VolumeFilesFor(args[0], cmdutil.SandboxOpts(cfg)...).List(context.Background(), path)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(entries)
	}

	if len(entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No entries found.")
		return nil
	}

	headers := []string{"Name", "Type", "Size", "Modified"}
	rows := make([][]string, len(entries))
	for i, e := range entries {
		typ := "file"
		if e.IsDir {
			typ = "dir"
		}
		rows[i] = []string{e.Name, typ, formatBytes(e.Size), e.ModTime}
	}
	p.PrintTable(headers, rows)
	return nil
}

// ---------------------------------------------------------------------------
// info
// ---------------------------------------------------------------------------

func newFilesInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info <volume-id>",
		Short: "Stat a file inside a volume (includes the CAS version)",
		Args:  cobra.ExactArgs(1),
		RunE:  runFilesInfo,
	}
	cmd.Flags().String("path", "", "Path inside the volume (required)")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func runFilesInfo(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}
	path, _ := cmd.Flags().GetString("path")

	info, err := declaw.VolumeFilesFor(args[0], cmdutil.SandboxOpts(cfg)...).Info(context.Background(), path)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(info)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Name:     %s\n", info.Name)
	fmt.Fprintf(cmd.OutOrStdout(), "Path:     %s\n", info.Path)
	fmt.Fprintf(cmd.OutOrStdout(), "IsDir:    %t\n", info.IsDir)
	fmt.Fprintf(cmd.OutOrStdout(), "Size:     %s\n", formatBytes(info.Size))
	fmt.Fprintf(cmd.OutOrStdout(), "Modified: %s\n", info.ModTime)
	fmt.Fprintf(cmd.OutOrStdout(), "Mode:     %o\n", info.Mode)
	fmt.Fprintf(cmd.OutOrStdout(), "Version:  %s\n", info.Version)
	return nil
}

// ---------------------------------------------------------------------------
// exists
// ---------------------------------------------------------------------------

func newFilesExistsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exists <volume-id>",
		Short: "Check whether a path exists inside a volume",
		Args:  cobra.ExactArgs(1),
		RunE:  runFilesExists,
	}
	cmd.Flags().String("path", "", "Path inside the volume (required)")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func runFilesExists(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}
	path, _ := cmd.Flags().GetString("path")

	exists, err := declaw.VolumeFilesFor(args[0], cmdutil.SandboxOpts(cfg)...).Exists(context.Background(), path)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(map[string]bool{"exists": exists})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%t\n", exists)
	return nil
}

// ---------------------------------------------------------------------------
// rm
// ---------------------------------------------------------------------------

func newFilesRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm <volume-id>",
		Short: "Remove a path inside a volume",
		Args:  cobra.ExactArgs(1),
		RunE:  runFilesRemove,
	}
	cmd.Flags().String("path", "", "Path inside the volume (required)")
	cmd.Flags().BoolP("recursive", "r", false, "Remove directories and their contents recursively")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func runFilesRemove(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}
	path, _ := cmd.Flags().GetString("path")
	recursive, _ := cmd.Flags().GetBool("recursive")

	if err := declaw.VolumeFilesFor(args[0], cmdutil.SandboxOpts(cfg)...).Remove(context.Background(), path, recursive); err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Removed %s\n", path)
	return nil
}

// ---------------------------------------------------------------------------
// mv
// ---------------------------------------------------------------------------

func newFilesRenameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mv <volume-id>",
		Short: "Rename a path inside a volume",
		Args:  cobra.ExactArgs(1),
		RunE:  runFilesRename,
	}
	cmd.Flags().String("old-path", "", "Existing path (required)")
	cmd.Flags().String("new-path", "", "New path (required)")
	_ = cmd.MarkFlagRequired("old-path")
	_ = cmd.MarkFlagRequired("new-path")
	return cmd
}

func runFilesRename(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}
	oldPath, _ := cmd.Flags().GetString("old-path")
	newPath, _ := cmd.Flags().GetString("new-path")

	if err := declaw.VolumeFilesFor(args[0], cmdutil.SandboxOpts(cfg)...).Rename(context.Background(), oldPath, newPath); err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Renamed %s -> %s\n", oldPath, newPath)
	return nil
}

// ---------------------------------------------------------------------------
// mkdir
// ---------------------------------------------------------------------------

func newFilesMkdirCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mkdir <volume-id>",
		Short: "Create a directory inside a volume",
		Args:  cobra.ExactArgs(1),
		RunE:  runFilesMkdir,
	}
	cmd.Flags().String("path", "", "Directory path inside the volume (required)")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func runFilesMkdir(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}
	path, _ := cmd.Flags().GetString("path")

	if err := declaw.VolumeFilesFor(args[0], cmdutil.SandboxOpts(cfg)...).Mkdir(context.Background(), path); err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created directory %s\n", path)
	return nil
}
