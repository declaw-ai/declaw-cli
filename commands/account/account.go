package account

import "github.com/spf13/cobra"

func NewAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage your Declaw account",
	}
	cmd.AddCommand(
		newInfoCmd(),
		newUsageCmd(),
		newAPIKeysCmd(),
	)
	return cmd
}
