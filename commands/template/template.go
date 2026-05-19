package template

import "github.com/spf13/cobra"

func NewTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "template",
		Short:   "Manage templates",
		Aliases: []string{"tpl"},
	}
	cmd.AddCommand(
		newListCmd(),
		newBuildCmd(),
		newInfoCmd(),
		newDeleteCmd(),
	)
	return cmd
}
