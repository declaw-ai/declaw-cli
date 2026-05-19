package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/declaw-ai/declaw-cli/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput, _ := cmd.Flags().GetBool("json"); jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(map[string]string{
					"version": version.Version,
					"commit":  version.Commit,
					"date":    version.Date,
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "declaw version %s (%s) built %s\n",
				version.Version, version.Commit, version.Date)
			return nil
		},
	}
}
