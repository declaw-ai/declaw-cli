package sandbox

import "github.com/spf13/cobra"

func NewSandboxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sandbox",
		Short:   "Manage sandboxes",
		Aliases: []string{"sb"},
	}
	cmd.AddCommand(
		newCreateCmd(),
		newListCmd(),
		newInfoCmd(),
		newKillCmd(),
		newPauseCmd(),
		newResumeCmd(),
		newExecCmd(),
		newConnectCmd(),
		newFilesCmd(),
	)
	return cmd
}
