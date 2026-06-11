package governance

import "github.com/spf13/cobra"

// NewGovernanceCmd returns the top-level 'governance' command group.
func NewGovernanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "governance",
		Short: "Manage governance and compliance packs",
	}
	cmd.AddCommand(newPacksCmd())
	return cmd
}
