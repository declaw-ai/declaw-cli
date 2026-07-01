package vault

import (
	"context"
	"fmt"
	"strings"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

// NewVaultCmd returns the top-level 'vault' command group.
func NewVaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vault",
		Short: "Manage vault secrets",
	}
	cmd.AddCommand(
		newVaultCreateCmd(),
		newVaultListCmd(),
		newVaultRotateCmd(),
		newVaultUpdateScopesCmd(),
		newVaultDeleteCmd(),
		newVaultPresetsCmd(),
	)
	return cmd
}

// parseScope parses a scope string in the format:
//
//	domain_regex,injection_type[,header_name]
//
// into a declaw.VaultScope. The third field (header_name) is optional.
func parseScope(s string) (declaw.VaultScope, error) {
	parts := strings.SplitN(s, ",", 3)
	if len(parts) < 2 {
		return declaw.VaultScope{}, fmt.Errorf("scope %q must be in format domain_regex,injection_type[,header_name]", s)
	}
	scope := declaw.VaultScope{
		DomainRegex:   parts[0],
		InjectionType: parts[1],
	}
	if len(parts) == 3 {
		scope.HeaderName = parts[2]
	}
	return scope, nil
}

// --- vault create ---

func newVaultCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a vault secret",
		Args:  cobra.NoArgs,
		RunE:  runVaultCreate,
	}
	cmd.Flags().String("name", "", "Secret name (defaults to provider key when --provider is set)")
	cmd.Flags().String("value", "", "Secret value (required)")
	cmd.Flags().String("provider", "", "Provider preset key (e.g. openai, anthropic); supplies scopes automatically")
	cmd.Flags().Int("rotation-days", 0, "Rotation interval in days (0 = no rotation policy)")
	cmd.Flags().StringArray("scope", nil, "Injection scope in format domain_regex,injection_type[,header_name] (repeatable; required unless --provider is set)")
	if err := cmd.MarkFlagRequired("value"); err != nil {
		panic(err)
	}
	return cmd
}

func runVaultCreate(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	name, _ := cmd.Flags().GetString("name")
	value, _ := cmd.Flags().GetString("value")
	provider, _ := cmd.Flags().GetString("provider")
	rotationDays, _ := cmd.Flags().GetInt("rotation-days")
	scopeStrs, _ := cmd.Flags().GetStringArray("scope")

	if provider == "" && len(scopeStrs) == 0 {
		return fmt.Errorf("either --provider or at least one --scope is required")
	}

	var scopes []declaw.VaultScope
	for _, s := range scopeStrs {
		scope, err := parseScope(s)
		if err != nil {
			return err
		}
		scopes = append(scopes, scope)
	}

	v := declaw.NewVaultClient(cmdutil.VaultOpts(cfg)...)
	defer v.Close()

	secret, err := v.CreateSecret(context.Background(), declaw.CreateSecretInput{
		Name:                 name,
		Value:                value,
		Provider:             provider,
		Scopes:               scopes,
		RotationIntervalDays: rotationDays,
	})
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(secret)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Secret ID: %s\n", secret.SecretID)
	fmt.Fprintf(cmd.OutOrStdout(), "Name:      %s\n", secret.Name)
	return nil
}

// --- vault list ---

func newVaultListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List vault secrets",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE:    runVaultList,
	}
	return cmd
}

func runVaultList(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	v := declaw.NewVaultClient(cmdutil.VaultOpts(cfg)...)
	defer v.Close()

	secrets, err := v.ListSecrets(context.Background())
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(secrets)
	}

	if len(secrets) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No secrets found.")
		return nil
	}

	headers := []string{"SECRET ID", "NAME", "INJECTION", "DOMAINS", "ROTATION DUE"}
	rows := make([][]string, len(secrets))
	for i, s := range secrets {
		injectionTypes := make([]string, 0, len(s.Scopes))
		domains := make([]string, 0, len(s.Scopes))
		for _, sc := range s.Scopes {
			injectionTypes = append(injectionTypes, sc.InjectionType)
			domains = append(domains, sc.DomainRegex)
		}
		rotationDue := "no"
		if s.RotationDue {
			rotationDue = "yes"
		}
		rows[i] = []string{
			s.SecretID,
			s.Name,
			strings.Join(injectionTypes, ","),
			strings.Join(domains, ","),
			rotationDue,
		}
	}
	p.PrintTable(headers, rows)
	return nil
}

// --- vault rotate ---

func newVaultRotateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate <name>",
		Short: "Rotate a secret value (by name)",
		Args:  cobra.ExactArgs(1),
		RunE:  runVaultRotate,
	}
	cmd.Flags().String("value", "", "New secret value (required)")
	if err := cmd.MarkFlagRequired("value"); err != nil {
		panic(err)
	}
	return cmd
}

func runVaultRotate(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	value, _ := cmd.Flags().GetString("value")

	v := declaw.NewVaultClient(cmdutil.VaultOpts(cfg)...)
	defer v.Close()

	if err := v.RotateSecret(context.Background(), args[0], value); err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Rotated secret %q\n", args[0])
	return nil
}

// --- vault update-scopes ---

func newVaultUpdateScopesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-scopes <name>",
		Short: "Replace a secret's injection scopes (by name); the value is unchanged",
		Args:  cobra.ExactArgs(1),
		RunE:  runVaultUpdateScopes,
	}
	cmd.Flags().StringArray("scope", nil, "Injection scope in format domain_regex,injection_type[,header_name] (repeatable; at least one required)")
	if err := cmd.MarkFlagRequired("scope"); err != nil {
		panic(err)
	}
	return cmd
}

func runVaultUpdateScopes(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	scopeStrs, _ := cmd.Flags().GetStringArray("scope")
	if len(scopeStrs) == 0 {
		return fmt.Errorf("at least one --scope is required")
	}
	var scopes []declaw.VaultScope
	for _, s := range scopeStrs {
		scope, err := parseScope(s)
		if err != nil {
			return err
		}
		scopes = append(scopes, scope)
	}

	v := declaw.NewVaultClient(cmdutil.VaultOpts(cfg)...)
	defer v.Close()

	if err := v.UpdateScopes(context.Background(), args[0], scopes); err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated scopes for secret %q\n", args[0])
	return nil
}

// --- vault delete ---

func newVaultDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a vault secret (by name)",
		Args:  cobra.ExactArgs(1),
		RunE:  runVaultDelete,
	}
}

func runVaultDelete(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	v := declaw.NewVaultClient(cmdutil.VaultOpts(cfg)...)
	defer v.Close()

	if err := v.DeleteSecret(context.Background(), args[0]); err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted secret %q\n", args[0])
	return nil
}

// --- vault presets ---

func newVaultPresetsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "presets",
		Short: "List built-in provider presets",
		Args:  cobra.NoArgs,
		RunE:  runVaultPresets,
	}
}

func runVaultPresets(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	v := declaw.NewVaultClient(cmdutil.VaultOpts(cfg)...)
	defer v.Close()

	presets, err := v.ListPresets(context.Background())
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(presets)
	}

	if len(presets) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No presets found.")
		return nil
	}

	headers := []string{"Key", "Name", "Category", "Key Hint"}
	rows := make([][]string, len(presets))
	for i, pr := range presets {
		rows[i] = []string{
			pr.Key,
			pr.Name,
			pr.Category,
			pr.KeyHint,
		}
	}
	p.PrintTable(headers, rows)
	return nil
}
