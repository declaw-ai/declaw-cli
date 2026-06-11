package volume

import "github.com/spf13/cobra"

func NewVolumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "volume",
		Short:   "Manage volumes",
		Aliases: []string{"vol"},
	}
	cmd.AddCommand(
		newCreateCmd(),
		newListCmd(),
		newGetCmd(),
		newDeleteCmd(),
		newSnapshotCmd(),
		newEmptyCmd(),
		newIngestCmd(),
		newFilesCmd(),
		newLockCmd(),
	)
	return cmd
}
