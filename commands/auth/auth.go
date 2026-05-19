package auth

import "github.com/spf13/cobra"

func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
	}
	cmd.AddCommand(
		newLoginCmd(),
		newLogoutCmd(),
		newStatusCmd(),
	)
	return cmd
}
