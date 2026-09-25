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

func newRebuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rebuild <template-id>",
		Short: "Retry a failed template build",
		Long: `Re-run the build of a template whose last build failed, reusing its stored
spec. Templates are immutable once built, so only a failed template can be
rebuilt: a ready template, or one whose build is still running, is refused.

A failed template keeps its alias, so running "template build" again with the
same alias is refused. To retry the same spec, rebuild it; to build a changed
spec, delete it first. The template ID is in that error, and in
"declaw template list".

The command waits for the build to finish and streams its output to stderr;
pass --no-wait to return as soon as the build has started. Ctrl-C stops the
wait, not the build.`,
		Example: `  declaw template rebuild tpl-1234567890abcdef
  declaw template rebuild tpl-1234567890abcdef --no-wait`,
		Args: cobra.ExactArgs(1),
		RunE: runRebuild,
	}
	cmd.Flags().Bool("no-wait", false, "Return immediately without waiting for the build to complete")
	return cmd
}

func runRebuild(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	templateID := args[0]
	opts := cmdutil.SandboxOpts(cfg)
	noWait, _ := cmd.Flags().GetBool("no-wait")

	// Ctrl-C stops the wait, not the build (same contract as "template build").
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	info, err := declaw.RebuildTemplateBackground(ctx, templateID, opts...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	// The alias is not in the build response. Look it up once for the hints
	// below, and just drop the hints rather than fail the command if that
	// lookup fails.
	alias := ""
	if tpl, err := declaw.GetTemplate(ctx, templateID, opts...); err == nil {
		alias = tpl.Alias
	}

	if !noWait {
		stderr := cmd.ErrOrStderr()
		fmt.Fprintf(stderr, "Rebuilding template %s (build %s); this usually takes several minutes...\n", templateID, info.BuildID)
		final, err := declaw.WaitForBuild(ctx, info.BuildID, func(line string) { fmt.Fprintln(stderr, line) }, opts...)
		if err != nil {
			if ctx.Err() != nil {
				fmt.Fprintf(stderr, "\nStopped waiting; build %s keeps running.", info.BuildID)
				if alias != "" {
					fmt.Fprintf(stderr, " Once it completes, create sandboxes with:\n  declaw sandbox create --template %s", alias)
				}
				fmt.Fprintln(stderr)
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
	if alias != "" {
		fmt.Fprintf(out, "Alias:       %s\n", alias)
	}
	fmt.Fprintf(out, "Template ID: %s\n", templateID)
	fmt.Fprintf(out, "Build ID:    %s\n", info.BuildID)
	fmt.Fprintf(out, "Status:      %s\n", info.Status)
	if info.Status == declaw.BuildStatusCompleted && alias != "" {
		fmt.Fprintf(out, "\nCreate a sandbox from it: declaw sandbox create --template %s\n", alias)
	}
	return nil
}
