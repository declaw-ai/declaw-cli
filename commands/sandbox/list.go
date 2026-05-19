package sandbox

import (
	"context"
	"fmt"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List sandboxes",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE:    runList,
	}
	cmd.Flags().String("state", "", "Filter by state (live, paused, killed)")
	cmd.Flags().Int("limit", 0, "Maximum number of results")
	cmd.Flags().Int("offset", 0, "Offset for pagination")
	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	opts := cmdutil.ListOpts(cfg)

	if state, _ := cmd.Flags().GetString("state"); state != "" {
		opts = append(opts, declaw.WithState(declaw.SandboxState(state)))
	}
	if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
		opts = append(opts, declaw.WithLimit(limit))
	}
	if offset, _ := cmd.Flags().GetInt("offset"); offset > 0 {
		opts = append(opts, declaw.WithOffset(offset))
	}

	page, err := declaw.ListSandboxes(context.Background(), opts...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(page)
	}

	if len(page.Sandboxes) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No sandboxes found.")
		return nil
	}

	headers := []string{"ID", "Template", "State", "Started"}
	rows := make([][]string, len(page.Sandboxes))
	for i, s := range page.Sandboxes {
		rows[i] = []string{
			s.SandboxID,
			s.TemplateID,
			string(s.State),
			output.RelativeTime(s.StartedAt),
		}
	}
	p.PrintTable(headers, rows)
	fmt.Fprintf(cmd.OutOrStdout(), "\nTotal: %d\n", page.Total)
	return nil
}
