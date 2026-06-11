package volume

import (
	"context"
	"fmt"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newLockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "lock",
		Short:   "Manage advisory locks (leases) over a volume path",
		Aliases: []string{"locks"},
	}
	cmd.AddCommand(
		newLockAcquireCmd(),
		newLockReleaseCmd(),
		newLockRenewCmd(),
		newLockStatusCmd(),
	)
	return cmd
}

// ---------------------------------------------------------------------------
// acquire
// ---------------------------------------------------------------------------

func newLockAcquireCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "acquire <volume-id>",
		Short: "Acquire an advisory lock on a volume path",
		Args:  cobra.ExactArgs(1),
		RunE:  runLockAcquire,
	}
	cmd.Flags().String("path", "", "Path to lock (required)")
	cmd.Flags().Int("ttl", 0, "Lock TTL in seconds (0 = server default)")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func runLockAcquire(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}
	path, _ := cmd.Flags().GetString("path")
	ttl, _ := cmd.Flags().GetInt("ttl")

	lock, err := declaw.VolumeLocksFor(args[0], cmdutil.SandboxOpts(cfg)...).Acquire(context.Background(), path, ttl)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(lock)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Acquired lock on %s\n", path)
	fmt.Fprintf(cmd.OutOrStdout(), "Token:      %s\n", lock.Token)
	fmt.Fprintf(cmd.OutOrStdout(), "TTL:        %ds\n", lock.TTLSeconds)
	fmt.Fprintf(cmd.OutOrStdout(), "Expires at: %s\n", lock.ExpiresAt)
	return nil
}

// ---------------------------------------------------------------------------
// release
// ---------------------------------------------------------------------------

func newLockReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release <volume-id>",
		Short: "Release an advisory lock on a volume path",
		Args:  cobra.ExactArgs(1),
		RunE:  runLockRelease,
	}
	cmd.Flags().String("path", "", "Path to unlock (required)")
	cmd.Flags().String("token", "", "Lock token returned by acquire (required)")
	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagRequired("token")
	return cmd
}

func runLockRelease(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}
	path, _ := cmd.Flags().GetString("path")
	token, _ := cmd.Flags().GetString("token")

	released, err := declaw.VolumeLocksFor(args[0], cmdutil.SandboxOpts(cfg)...).Release(context.Background(), path, token)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(map[string]bool{"released": released})
	}

	if released {
		fmt.Fprintf(cmd.OutOrStdout(), "Released lock on %s\n", path)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Lock on %s was not held\n", path)
	}
	return nil
}

// ---------------------------------------------------------------------------
// renew
// ---------------------------------------------------------------------------

func newLockRenewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "renew <volume-id>",
		Short: "Renew an advisory lock on a volume path",
		Args:  cobra.ExactArgs(1),
		RunE:  runLockRenew,
	}
	cmd.Flags().String("path", "", "Locked path (required)")
	cmd.Flags().String("token", "", "Lock token returned by acquire (required)")
	cmd.Flags().Int("ttl", 0, "New TTL in seconds (0 = server default)")
	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagRequired("token")
	return cmd
}

func runLockRenew(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}
	path, _ := cmd.Flags().GetString("path")
	token, _ := cmd.Flags().GetString("token")
	ttl, _ := cmd.Flags().GetInt("ttl")

	if err := declaw.VolumeLocksFor(args[0], cmdutil.SandboxOpts(cfg)...).Renew(context.Background(), path, token, ttl); err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Renewed lock on %s\n", path)
	return nil
}

// ---------------------------------------------------------------------------
// status
// ---------------------------------------------------------------------------

func newLockStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <volume-id>",
		Short: "Show the lock status of a volume path",
		Args:  cobra.ExactArgs(1),
		RunE:  runLockStatus,
	}
	cmd.Flags().String("path", "", "Path to check (required)")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func runLockStatus(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}
	path, _ := cmd.Flags().GetString("path")

	status, err := declaw.VolumeLocksFor(args[0], cmdutil.SandboxOpts(cfg)...).Status(context.Background(), path)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(status)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Held:       %t\n", status.Held)
	fmt.Fprintf(cmd.OutOrStdout(), "Expires in: %dms\n", status.ExpiresInMs)
	return nil
}
