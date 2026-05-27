package sandbox

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	"github.com/declaw-ai/declaw-cli/internal/output"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newExecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec <sandbox-id> -- <command...>",
		Short: "Run a command in a sandbox",
		Long:  "Execute a command inside a running sandbox. Use -- to separate CLI flags from the command.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runExec,
	}
	cmd.Flags().String("cwd", "", "Working directory inside the sandbox")
	cmd.Flags().String("user", "", "User to run the command as")
	cmd.Flags().Int("timeout", 60, "Command timeout in seconds")
	cmd.Flags().StringArrayP("env", "e", nil, "Environment variables (KEY or KEY=VAL, repeatable)")
	return cmd
}

func runExec(cmd *cobra.Command, args []string) error {
	sandboxID := args[0]
	cmdArgs := args[1:]

	if len(cmdArgs) == 0 {
		return fmt.Errorf("no command specified; use -- to separate: declaw sandbox exec %s -- <command>", sandboxID)
	}

	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	sbx, err := declaw.Connect(context.Background(), sandboxID, cmdutil.SandboxOpts(cfg)...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	command := shellJoin(cmdArgs)
	var runOpts []declaw.RunOption

	if cwd, _ := cmd.Flags().GetString("cwd"); cwd != "" {
		runOpts = append(runOpts, declaw.WithCwd(cwd))
	}
	if user, _ := cmd.Flags().GetString("user"); user != "" {
		runOpts = append(runOpts, declaw.WithUser(user))
	}
	if timeout, _ := cmd.Flags().GetInt("timeout"); timeout > 0 {
		runOpts = append(runOpts, declaw.WithRunTimeout(time.Duration(timeout)*time.Second))
	}
	if envs, _ := cmd.Flags().GetStringArray("env"); len(envs) > 0 {
		m, err := cmdutil.ParseEnvPairs(envs)
		if err != nil {
			return err
		}
		runOpts = append(runOpts, declaw.WithRunEnvs(m))
	}

	jsonMode := cmdutil.JSONOutput(cmd)
	if !jsonMode {
		runOpts = append(runOpts,
			declaw.WithOnStdout(func(line string) {
				fmt.Fprintln(os.Stdout, line)
			}),
			declaw.WithOnStderr(func(line string) {
				fmt.Fprintln(os.Stderr, line)
			}),
		)
	}

	result, err := sbx.Commands.Run(context.Background(), command, runOpts...)
	if err != nil {
		var cmdExit *declaw.CommandExitError
		if errors.As(err, &cmdExit) {
			if jsonMode {
				p := output.New(true)
				p.PrintJSON(map[string]interface{}{
					"exit_code": cmdExit.ExitCode,
					"stdout":    cmdExit.Stdout,
					"stderr":    cmdExit.Stderr,
				})
			}
			os.Exit(cmdExit.ExitCode)
		}
		cmdutil.HandleError(err)
		return nil
	}

	if jsonMode {
		p := output.New(true)
		return p.PrintJSON(map[string]interface{}{
			"exit_code": result.ExitCode,
			"stdout":    result.Stdout,
			"stderr":    result.Stderr,
		})
	}

	return nil
}

func shellJoin(args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		if a == "" || needsQuoting(a) {
			quoted[i] = "'" + strings.ReplaceAll(a, "'", "'\\''") + "'"
		} else {
			quoted[i] = a
		}
	}
	return strings.Join(quoted, " ")
}

func needsQuoting(s string) bool {
	for _, r := range s {
		if unicode.IsSpace(r) {
			return true
		}
		switch r {
		case '\'', '"', '\\', '$', '`', '!', '(', ')', '{', '}', '[', ']',
			'|', '&', ';', '<', '>', '*', '?', '#', '~', '^':
			return true
		}
	}
	return false
}
