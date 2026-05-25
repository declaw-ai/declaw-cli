package commands

import (
	"github.com/declaw-ai/declaw-cli/commands/auth"
	"github.com/declaw-ai/declaw-cli/commands/account"
	"github.com/declaw-ai/declaw-cli/commands/sandbox"
	"github.com/declaw-ai/declaw-cli/commands/template"
	"github.com/declaw-ai/declaw-cli/commands/volume"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "declaw",
		Short: "CLI for the Declaw sandbox platform",
		Long:  "Create, manage, and interact with secure sandboxes for AI agents.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().String("api-key", "", "Declaw API key (overrides DECLAW_API_KEY and config file)")
	cmd.PersistentFlags().String("domain", "", "API domain (overrides DECLAW_DOMAIN and config file)")
	cmd.PersistentFlags().Bool("json", false, "Output as JSON")

	cmd.AddCommand(
		auth.NewAuthCmd(),
		sandbox.NewSandboxCmd(),
		template.NewTemplateCmd(),
		volume.NewVolumeCmd(),
		account.NewAccountCmd(),
		newVersionCmd(),
		newMcpCmd(),
	)

	return cmd
}
