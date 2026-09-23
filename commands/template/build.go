package template

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build a new template",
		Long: `Build a custom template. The command waits for the build to finish,
which usually takes several minutes, and streams its output to stderr;
pass --no-wait to return as soon as the build has started. Ctrl-C stops
the wait, not the build.

Sandboxes are created from the template by its alias:
  declaw sandbox create --template <alias>`,
		Example: `  declaw template build --alias my-tools --apt-package ffmpeg --apt-package jq
  declaw template build --alias my-mcp --dockerfile Dockerfile`,
		Args: cobra.NoArgs,
		RunE: runBuild,
	}
	cmd.Flags().String("alias", "", "Template name, used to create sandboxes from it (required; lowercase letters, digits and hyphens)")
	cmd.Flags().String("base-image", "", "Base image (e.g., ubuntu:22.04)")
	// A StringArray, not a StringSlice: each value is one shell command, taken
	// verbatim. A StringSlice splits on commas, which turned "echo a,b" into
	// two commands and rejected pip pins like "numpy>=1.24,<2" outright.
	cmd.Flags().StringArray("run-cmd", nil, "Command to run during the build (repeatable, one command per flag)")
	cmd.Flags().StringSlice("apt-package", nil, "Apt packages to install (repeatable)")
	cmd.Flags().String("start-cmd", "", "Command to run on sandbox start")
	cmd.Flags().String("dockerfile", "", "Path to a Dockerfile to use instead of structured fields")
	cmd.Flags().Int("disk-mb", 0, "Disk size in MB")
	cmd.Flags().Bool("no-wait", false, "Return immediately without waiting for build to complete")
	_ = cmd.MarkFlagRequired("alias")
	return cmd
}

// specFromFlags builds the template spec from the command's flags.
func specFromFlags(cmd *cobra.Command) (declaw.TemplateSpec, error) {
	spec := declaw.TemplateSpec{}

	spec.Alias, _ = cmd.Flags().GetString("alias")
	spec.BaseImage, _ = cmd.Flags().GetString("base-image")
	spec.RunCmds, _ = cmd.Flags().GetStringArray("run-cmd")
	spec.AptPackages, _ = cmd.Flags().GetStringSlice("apt-package")
	spec.StartCmd, _ = cmd.Flags().GetString("start-cmd")
	spec.DiskMB, _ = cmd.Flags().GetInt("disk-mb")

	if df, _ := cmd.Flags().GetString("dockerfile"); df != "" {
		data, err := os.ReadFile(df)
		if err != nil {
			return spec, fmt.Errorf("reading dockerfile %s: %w", df, err)
		}
		spec.Dockerfile = string(data)
	}
	return spec, nil
}

func runBuild(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	spec, err := specFromFlags(cmd)
	if err != nil {
		return err
	}

	opts := cmdutil.SandboxOpts(cfg)
	noWait, _ := cmd.Flags().GetBool("no-wait")

	// Ctrl-C stops the wait, not the build, so say so rather than dying
	// mid-wait: rerunning the same command would only hit "alias exists".
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	info, err := declaw.BuildTemplateBackground(ctx, spec, opts...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	if !noWait {
		stderr := cmd.ErrOrStderr()
		fmt.Fprintf(stderr, "Building template %q (build %s); this usually takes several minutes...\n", spec.Alias, info.BuildID)
		final, err := declaw.WaitForBuild(ctx, info.BuildID, func(line string) { fmt.Fprintln(stderr, line) }, opts...)
		if err != nil {
			if ctx.Err() != nil {
				fmt.Fprintf(stderr, "\nStopped waiting; build %s keeps running. Once it completes, create sandboxes with:\n  declaw sandbox create --template %s\n", info.BuildID, spec.Alias)
				os.Exit(130)
			}
			var buildErr *declaw.BuildError
			if errors.As(err, &buildErr) {
				// The build's output is already on stderr above; don't repeat its tail.
				fmt.Fprintf(stderr, "Error: template build %s failed\n", info.BuildID)
				os.Exit(1)
			}
			cmdutil.HandleError(err)
			return nil
		}
		info = final
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(info)
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Alias:       %s\n", spec.Alias)
	fmt.Fprintf(out, "Build ID:    %s\n", info.BuildID)
	fmt.Fprintf(out, "Status:      %s\n", info.Status)
	if info.TemplateID != "" {
		fmt.Fprintf(out, "Template ID: %s\n", info.TemplateID)
	}
	if info.Status == declaw.BuildStatusCompleted {
		fmt.Fprintf(out, "\nCreate a sandbox from it: declaw sandbox create --template %s\n", spec.Alias)
	}
	return nil
}
