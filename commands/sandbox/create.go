package sandbox

import (
	"context"
	"fmt"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new sandbox",
		Args:  cobra.NoArgs,
		RunE:  runCreate,
	}
	cmd.Flags().StringP("template", "t", "base", "Template to use")
	cmd.Flags().Int("timeout", 300, "Sandbox timeout in seconds")
	cmd.Flags().StringSliceP("env", "e", nil, "Environment variables (KEY=VAL)")
	cmd.Flags().StringSlice("metadata", nil, "Metadata (KEY=VAL)")
	cmd.Flags().StringSlice("volume", nil, "Attach volumes (VOLUME_ID:MOUNT_PATH)")
	cmd.Flags().Bool("secure", true, "Enable security pipeline")
	return cmd
}

func runCreate(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	opts := cmdutil.SandboxOpts(cfg)

	tmpl, _ := cmd.Flags().GetString("template")
	opts = append(opts, declaw.WithTemplate(tmpl))

	timeout, _ := cmd.Flags().GetInt("timeout")
	opts = append(opts, declaw.WithTimeout(timeout))

	if envs, _ := cmd.Flags().GetStringSlice("env"); len(envs) > 0 {
		m, err := cmdutil.ParseKeyValues(envs)
		if err != nil {
			return err
		}
		opts = append(opts, declaw.WithEnvs(m))
	}

	if meta, _ := cmd.Flags().GetStringSlice("metadata"); len(meta) > 0 {
		m, err := cmdutil.ParseKeyValues(meta)
		if err != nil {
			return err
		}
		opts = append(opts, declaw.WithMetadata(m))
	}

	if vols, _ := cmd.Flags().GetStringSlice("volume"); len(vols) > 0 {
		attachments, err := parseVolumeAttachments(vols)
		if err != nil {
			return err
		}
		opts = append(opts, declaw.WithVolumes(attachments))
	}

	if cmd.Flags().Changed("secure") {
		secure, _ := cmd.Flags().GetBool("secure")
		opts = append(opts, declaw.WithSecure(secure))
	}

	sbx, err := declaw.Create(context.Background(), opts...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(map[string]interface{}{
			"sandbox_id": sbx.ID,
			"template":   tmpl,
			"timeout":    timeout,
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created sandbox %s (template: %s, timeout: %ds)\n", sbx.ID, tmpl, timeout)
	return nil
}

func parseVolumeAttachments(vols []string) ([]declaw.VolumeAttachment, error) {
	var attachments []declaw.VolumeAttachment
	for _, v := range vols {
		parts := splitFirst(v, ':')
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid volume format %q, expected VOLUME_ID:MOUNT_PATH", v)
		}
		attachments = append(attachments, declaw.VolumeAttachment{
			VolumeID:  parts[0],
			MountPath: parts[1],
		})
	}
	return attachments, nil
}

func splitFirst(s string, sep byte) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}
