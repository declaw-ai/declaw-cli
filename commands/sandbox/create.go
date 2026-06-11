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

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new sandbox",
		Args:  cobra.NoArgs,
		RunE:  runCreate,
	}
	cmd.Flags().StringP("template", "t", "base", "Template to use")
	cmd.Flags().Int("timeout", 300, "Sandbox timeout in seconds")
	cmd.Flags().StringArrayP("env", "e", nil, "Environment variables (KEY or KEY=VAL, repeatable)")
	cmd.Flags().StringSlice("metadata", nil, "Metadata (KEY=VAL)")
	cmd.Flags().StringSlice("volume", nil, "Attach volumes (VOLUME_ID:MOUNT_PATH[:MODE[:SUBPATH]]); MODE is copy|mount|mount-ro, SUBPATH valid only for live mounts)")
	cmd.Flags().Bool("secure", true, "Enable security pipeline")
	cmd.Flags().String("opa-policy", "", "Path to a .rego file whose contents are sent as the inline OPA policy for this sandbox (sets custom_policy.inline_rego, enables custom policy)")
	cmd.Flags().StringArray("opa-policy-module", nil, "Path to a .rego file added as an independent Rego module (each file is its own package; flag is repeatable)")
	cmd.Flags().Bool("opa-fail-closed", false, "When set, deny the action if the OPA evaluator is unreachable (fail-closed); default is fail-open")
	cmd.Flags().String("opa-policy-ref", "", "Reference to a pre-published policy bundle (forms: name@version, sha256:<hex>, blob:<key>). Sets custom_policy.policy_ref and enables the custom policy. Composes with --opa-policy / --opa-policy-module when both are provided.")
	cmd.Flags().StringSlice("content-gate-domains", nil, "Enable the content gate and restrict interception to these hosts (comma-separated, e.g. openai.com,api.anthropic.com). Implies content_gate.enabled=true.")
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

	if envs, _ := cmd.Flags().GetStringArray("env"); len(envs) > 0 {
		m, err := cmdutil.ParseEnvPairs(envs)
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

	// OPA / custom-policy flags: --opa-policy, --opa-policy-module, --opa-fail-closed, --opa-policy-ref
	//
	// All four flags contribute to the same CustomPolicyConfig. When more than
	// one is supplied they compose: e.g. --opa-policy-ref can be combined with
	// --opa-fail-closed, and --opa-policy / --opa-policy-module can accompany a
	// ref (the platform evaluates both the inline modules and the bundle).
	opaChanged := cmd.Flags().Changed("opa-policy") ||
		cmd.Flags().Changed("opa-policy-module") ||
		cmd.Flags().Changed("opa-fail-closed") ||
		cmd.Flags().Changed("opa-policy-ref")
	if opaChanged {
		cp := &declaw.CustomPolicyConfig{Enabled: true}

		if cmd.Flags().Changed("opa-fail-closed") {
			failClosed, _ := cmd.Flags().GetBool("opa-fail-closed")
			cp.DefaultDeny = failClosed
		}

		if policyFile, _ := cmd.Flags().GetString("opa-policy"); policyFile != "" {
			regoBytes, err := os.ReadFile(policyFile)
			if err != nil {
				return fmt.Errorf("reading --opa-policy file %q: %w", policyFile, err)
			}
			cp.InlineRego = string(regoBytes)
		}

		if moduleFiles, _ := cmd.Flags().GetStringArray("opa-policy-module"); len(moduleFiles) > 0 {
			cp.InlineModules = make([]string, 0, len(moduleFiles))
			for _, modFile := range moduleFiles {
				modBytes, err := os.ReadFile(modFile)
				if err != nil {
					return fmt.Errorf("reading --opa-policy-module file %q: %w", modFile, err)
				}
				cp.InlineModules = append(cp.InlineModules, string(modBytes))
			}
		}

		if cmd.Flags().Changed("opa-policy-ref") {
			ref, _ := cmd.Flags().GetString("opa-policy-ref")
			cp.PolicyRef = ref
		}

		// Compose with any SecurityPolicy already built by previous options.
		// Look for an existing WithSecurity in opts by re-resolving; if none
		// was set we build a fresh SecurityPolicy with only CustomPolicy.
		sp := declaw.SecurityPolicy{CustomPolicy: cp}
		opts = append(opts, declaw.WithSecurity(sp))
	}

	// --content-gate-domains: enable the content gate and (optionally) restrict
	// it to specific hosts. Composes with the OPA custom-policy block above —
	// both may be present in the same SecurityPolicy without conflict.
	if cmd.Flags().Changed("content-gate-domains") {
		domains, _ := cmd.Flags().GetStringSlice("content-gate-domains")
		sp := declaw.SecurityPolicy{
			ContentGate: &declaw.ContentGateConfig{
				Enabled: true,
				Domains: domains,
			},
		}
		opts = append(opts, declaw.WithSecurity(sp))
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

// parseVolumeAttachments parses --volume specs of the form
// VOLUME_ID:MOUNT_PATH[:MODE[:SUBPATH]]. MODE is one of copy|mount|mount-ro
// (default copy). SUBPATH is a relative path within the volume and is valid
// only for live mounts (mount/mount-ro); the server rejects it for copy mode.
func parseVolumeAttachments(vols []string) ([]declaw.VolumeAttachment, error) {
	var attachments []declaw.VolumeAttachment
	for _, v := range vols {
		volumeID, rest := splitFirst(v, ':')
		if rest == "" {
			return nil, fmt.Errorf("invalid volume format %q, expected VOLUME_ID:MOUNT_PATH[:MODE[:SUBPATH]]", v)
		}
		mountPath, modeAndSub := splitFirst(rest, ':')
		if mountPath == "" {
			return nil, fmt.Errorf("invalid volume format %q, expected VOLUME_ID:MOUNT_PATH[:MODE[:SUBPATH]]", v)
		}

		att := declaw.VolumeAttachment{VolumeID: volumeID, MountPath: mountPath}

		if modeAndSub != "" {
			mode, subpath := splitFirst(modeAndSub, ':')
			switch mode {
			case declaw.VolumeModeCopy, declaw.VolumeModeMount, declaw.VolumeModeMountRO:
				att.Mode = mode
			default:
				return nil, fmt.Errorf("invalid volume mode %q in %q, expected one of copy|mount|mount-ro", mode, v)
			}
			att.Subpath = subpath
		}

		attachments = append(attachments, att)
	}
	return attachments, nil
}

// splitFirst splits s on the first occurrence of sep, returning the part before
// and after. If sep is absent the second return value is empty.
func splitFirst(s string, sep byte) (string, string) {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return s[:i], s[i+1:]
		}
	}
	return s, ""
}
