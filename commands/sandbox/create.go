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
	cmd.Flags().StringArray("vault-ref", nil, "Vault-backed env var (ENV_VAR=secret-name, repeatable). The VM receives a placeholder; the real secret is injected at the egress proxy and never enters the sandbox.")
	cmd.Flags().StringSlice("metadata", nil, "Metadata (KEY=VAL)")
	cmd.Flags().StringSlice("volume", nil, "Attach volumes (VOLUME_ID:MOUNT_PATH[:MODE[:SUBPATH]]); MODE is copy|mount|mount-ro, SUBPATH valid only for live mounts)")
	cmd.Flags().Bool("secure", true, "Enable security pipeline")
	cmd.Flags().String("opa-policy", "", "Path to a .rego file whose contents are sent as the inline OPA policy for this sandbox (sets custom_policy.inline_rego, enables custom policy)")
	cmd.Flags().StringArray("opa-policy-module", nil, "Path to a .rego file added as an independent Rego module (each file is its own package; flag is repeatable)")
	cmd.Flags().Bool("opa-fail-closed", false, "When set, deny the action if the OPA evaluator is unreachable (fail-closed); default is fail-open")
	cmd.Flags().String("opa-policy-ref", "", "Reference to a pre-published policy bundle (forms: name@version, sha256:<hex>, blob:<key>). Sets custom_policy.policy_ref and enables the custom policy. Composes with --opa-policy / --opa-policy-module when both are provided.")
	cmd.Flags().StringSlice("content-gate-domains", nil, "Enable the content gate and restrict interception to these hosts (comma-separated, e.g. openai.com,api.anthropic.com). Implies content_gate.enabled=true.")
	cmd.Flags().Bool("injection-full", false, "Enable the FULL prompt-injection cascade in one flag: Tier-1 ML classifier + Layer-A static signatures + posture + Tier-2 Gemma LLM judge (multi-turn/indirect) + the OPA prompt-injection pack")
	cmd.Flags().String("injection-mode", "balanced", "Posture for --injection-full: strict|balanced|permissive|agentic-tool|data-egress-sensitive")
	cmd.Flags().String("injection-policy", "", "Natural-language description of what the agent may do (the Tier-2 judge uses it to tell task-aligned egress from injection); used with --injection-full")
	cmd.Flags().Bool("injection-always-judge", false, "Run the Tier-2 judge on every egress (high-assurance, costlier); used with --injection-full")
	cmd.Flags().StringArray("injection-domain", nil, "Restrict injection scanning to these destination hosts (repeatable, e.g. api.openai.com). Required for injection scanning to run: with no hosts listed, no injection scanning happens. Hosts support exact matches, \"*.suffix.com\" wildcards, and \"~regex\" patterns.")
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

	// --vault-ref ENV_VAR=vault://team/env/secret: the VM env carries only a
	// placeholder; the real value is resolved + injected at the egress proxy.
	if refs, _ := cmd.Flags().GetStringArray("vault-ref"); len(refs) > 0 {
		m, err := cmdutil.ParseKeyValues(refs)
		if err != nil {
			return fmt.Errorf("invalid --vault-ref (want ENV_VAR=secret-name): %w", err)
		}
		opts = append(opts, declaw.WithVaultRefs(m))
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

	// Policy-shaping flags accumulate into ONE SecurityPolicy, applied with a
	// single WithSecurity at the end — so --injection-full, the OPA custom-policy
	// flags, and --content-gate-domains compose instead of overwriting each other.
	var sp declaw.SecurityPolicy
	secTouched := false

	// --injection-full: the entire prompt-injection cascade in one flag —
	// Tier-1 ML classifier + Layer-A static signatures + posture + the Tier-2
	// Gemma judge (multi-turn/indirect) + the OPA prompt-injection governance pack.
	if full, _ := cmd.Flags().GetBool("injection-full"); full {
		mode, _ := cmd.Flags().GetString("injection-mode")
		agentPolicy, _ := cmd.Flags().GetString("injection-policy")
		always, _ := cmd.Flags().GetBool("injection-always-judge")
		fp := declaw.FullInjectionDefensePolicy(declaw.FullInjectionDefenseOptions{
			Mode:        mode,
			AgentPolicy: agentPolicy,
			AlwaysJudge: always,
		})
		sp.InjectionDefense = fp.InjectionDefense
		sp.CustomPolicy = fp.CustomPolicy
		secTouched = true
	}

	// --injection-domain: scope injection scanning to specific destination hosts.
	// Injection is opt-in per host — with no hosts listed no scanning runs — so a
	// bare --injection-domain (without --injection-full) still enables injection
	// defense, just confined to the named hosts. When --injection-full is also
	// present these hosts attach to the cascade it built above.
	if cmd.Flags().Changed("injection-domain") {
		domains, _ := cmd.Flags().GetStringArray("injection-domain")
		if sp.InjectionDefense == nil {
			sp.InjectionDefense = &declaw.InjectionDefenseConfig{Enabled: true}
		}
		sp.InjectionDefense.Domains = domains
		secTouched = true
	}

	// OPA / custom-policy flags: --opa-policy, --opa-policy-module,
	// --opa-fail-closed, --opa-policy-ref. All contribute to the same
	// CustomPolicyConfig and compose with each other (and with --injection-full,
	// whose pack ref an explicit --opa-policy-ref overrides).
	opaChanged := cmd.Flags().Changed("opa-policy") ||
		cmd.Flags().Changed("opa-policy-module") ||
		cmd.Flags().Changed("opa-fail-closed") ||
		cmd.Flags().Changed("opa-policy-ref")
	if opaChanged {
		cp := sp.CustomPolicy
		if cp == nil {
			cp = &declaw.CustomPolicyConfig{}
		}
		cp.Enabled = true

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

		sp.CustomPolicy = cp
		secTouched = true
	}

	// --content-gate-domains: enable the content gate and (optionally) restrict
	// it to specific hosts.
	if cmd.Flags().Changed("content-gate-domains") {
		domains, _ := cmd.Flags().GetStringSlice("content-gate-domains")
		sp.ContentGate = &declaw.ContentGateConfig{
			Enabled: true,
			Domains: domains,
		}
		secTouched = true
	}

	if secTouched {
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
