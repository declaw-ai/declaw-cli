package governance

import (
	"context"
	"fmt"
	"strconv"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

// newPacksCmd returns the 'governance packs' subcommand group.
func newPacksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packs",
		Short: "Manage governance packs",
	}
	cmd.AddCommand(newPacksListCmd())
	return cmd
}

func newPacksListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List available governance packs",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE:    runPacksList,
	}
}

func runPacksList(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	packs, err := declaw.ListGovernancePacks(context.Background(), cmdutil.SandboxOpts(cfg)...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(packs)
	}

	if len(packs) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No governance packs found.")
		return nil
	}

	headers := []string{"Name", "Version", "Framework", "Enforces", "Advisory", "Policy Ref"}
	rows := make([][]string, len(packs))
	for i, pack := range packs {
		rows[i] = []string{
			pack.Name,
			pack.Version,
			pack.Framework,
			strconv.Itoa(len(pack.Enforces)),
			strconv.Itoa(len(pack.Advisory)),
			pack.PolicyRef,
		}
	}
	p.PrintTable(headers, rows)
	return nil
}
