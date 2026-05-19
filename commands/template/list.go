package template

import (
	"context"
	"fmt"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List templates",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE:    runList,
	}
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	templates, err := declaw.ListTemplates(context.Background(), cmdutil.SandboxOpts(cfg)...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	p := output.New(cmdutil.JSONOutput(cmd))
	if p.JSON {
		return p.PrintJSON(templates)
	}

	if len(templates) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No templates found.")
		return nil
	}

	headers := []string{"ID", "Alias", "Created"}
	rows := make([][]string, len(templates))
	for i, t := range templates {
		rows[i] = []string{t.TemplateID, t.Alias, t.CreatedAt}
	}
	p.PrintTable(headers, rows)
	return nil
}
