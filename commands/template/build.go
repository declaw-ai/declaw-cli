package template

import (
	"context"
	"fmt"
	"os"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build a new template",
		Args:  cobra.NoArgs,
		RunE:  runBuild,
	}
	cmd.Flags().String("base-image", "", "Base image (e.g., ubuntu:22.04)")
	cmd.Flags().StringSlice("run-cmd", nil, "Commands to run during build (repeatable)")
	cmd.Flags().StringSlice("apt-package", nil, "Apt packages to install (repeatable)")
	cmd.Flags().String("start-cmd", "", "Command to run on sandbox start")
	cmd.Flags().String("dockerfile", "", "Path to a Dockerfile to use instead of structured fields")
	cmd.Flags().Int("disk-mb", 0, "Disk size in MB")
	cmd.Flags().Bool("no-wait", false, "Return immediately without waiting for build to complete")
	return cmd
}

func runBuild(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	spec := declaw.TemplateSpec{}

	spec.BaseImage, _ = cmd.Flags().GetString("base-image")
	spec.RunCmds, _ = cmd.Flags().GetStringSlice("run-cmd")
	spec.AptPackages, _ = cmd.Flags().GetStringSlice("apt-package")
	spec.StartCmd, _ = cmd.Flags().GetString("start-cmd")
	spec.DiskMB, _ = cmd.Flags().GetInt("disk-mb")

	if df, _ := cmd.Flags().GetString("dockerfile"); df != "" {
		data, err := os.ReadFile(df)
		if err != nil {
			return fmt.Errorf("reading dockerfile %s: %w", df, err)
		}
		spec.Dockerfile = string(data)
	}

	opts := cmdutil.SandboxOpts(cfg)
	noWait, _ := cmd.Flags().GetBool("no-wait")

	var info *declaw.BuildInfo
	if noWait {
		info, err = declaw.BuildTemplateBackground(context.Background(), spec, opts...)
	} else {
		info, err = declaw.BuildTemplate(context.Background(), spec, opts...)
	}
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(info)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Build ID:    %s\n", info.BuildID)
	fmt.Fprintf(cmd.OutOrStdout(), "Status:      %s\n", info.Status)
	if info.TemplateID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Template ID: %s\n", info.TemplateID)
	}
	return nil
}
