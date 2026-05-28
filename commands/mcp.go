package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func newMcpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp [flags] -- <command> [args...]",
		Short: "Run an MCP server inside a sandboxed environment",
		Long: `Wrap any stdio-based MCP server in a Declaw sandbox.

Add "declaw mcp --" before your existing MCP server command in your
client config (Claude Desktop, Cursor, Windsurf, Claude Code, etc.) to sandbox it.

Example config:
  {
    "mcpServers": {
      "github": {
        "command": "declaw",
        "args": ["mcp", "--env", "GITHUB_PERSONAL_ACCESS_TOKEN", "--network-allow", "registry.npmjs.org,api.github.com,github.com,codeload.github.com", "--", "npx", "-y", "@modelcontextprotocol/server-github"],
        "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_..." }
      }
    }
  }`,
		Args:              cobra.MinimumNArgs(1),
		RunE:              runMcp,
		DisableFlagParsing: false,
	}

	cmd.SetOut(os.Stderr)
	cmd.SetErr(os.Stderr)

	cmd.Flags().StringP("template", "t", "mcp-server", "Sandbox template (default: mcp-server, includes Node.js + Python)")
	cmd.Flags().Int("timeout", 86400, "Sandbox timeout in seconds (default 24h)")
	cmd.Flags().StringArrayP("env", "e", nil, "Environment variables to forward (KEY or KEY=VAL, repeatable)")
	cmd.Flags().StringSlice("network-allow", nil, "Allowed outbound hosts (comma-separated)")
	cmd.Flags().BoolP("verbose", "v", false, "Diagnostic logging to stderr")

	return cmd
}

func runMcp(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	verbose, _ := cmd.Flags().GetBool("verbose")
	logf := func(format string, a ...interface{}) {
		if verbose {
			fmt.Fprintf(os.Stderr, "[declaw] "+format+"\n", a...)
		}
	}

	tmpl, _ := cmd.Flags().GetString("template")
	timeout, _ := cmd.Flags().GetInt("timeout")

	opts := cmdutil.SandboxOpts(cfg)
	opts = append(opts, declaw.WithTemplate(tmpl))
	opts = append(opts, declaw.WithTimeout(timeout))

	var envs map[string]string
	if envPairs, _ := cmd.Flags().GetStringArray("env"); len(envPairs) > 0 {
		var err error
		envs, err = cmdutil.ParseEnvPairs(envPairs)
		if err != nil {
			return err
		}
		opts = append(opts, declaw.WithEnvs(envs))
	}

	allowHosts, _ := cmd.Flags().GetStringSlice("network-allow")
	if len(allowHosts) > 0 {
		opts = append(opts, declaw.WithNetwork(declaw.SandboxNetworkOpts{
			AllowOut: allowHosts,
			DenyOut:  []string{"*"},
		}))
	} else {
		fmt.Fprintln(os.Stderr, "[declaw] network: deny-all (use --network-allow to permit outbound hosts)")
		opts = append(opts, declaw.WithNetwork(declaw.SandboxNetworkOpts{
			DenyOut: []string{"*"},
		}))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	go func() {
		sig := <-sigCh
		fmt.Fprintf(os.Stderr, "[declaw] received %s, shutting down\n", sig)
		cancel()
	}()

	logf("creating sandbox (template=%s, timeout=%ds)", tmpl, timeout)
	sbx, err := declaw.Create(ctx, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[declaw] sandbox creation failed: %v\n", err)
		os.Exit(125)
	}
	killSandbox := func() {
		logf("killing sandbox %s", sbx.ID)
		sbx.Kill(context.Background())
	}
	defer killSandbox()
	logf("sandbox ready: %s", sbx.ID)

	fullCmd := buildShellCommand(args)
	logf("starting: %s", fullCmd)

	var stdiOpts *declaw.StdioStartOpts
	if envs != nil {
		stdiOpts = &declaw.StdioStartOpts{Envs: envs}
	}

	handle, err := sbx.Stdio.Start(ctx, fullCmd, stdiOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[declaw] process start failed: %v\n", err)
		killSandbox()
		os.Exit(125)
	}

	exitCh := make(chan int, 1)
	errCh := make(chan error, 1)

	go func() {
		result, err := handle.Stream(ctx, &declaw.StdioStreamOpts{
			OnStdout: func(data []byte) {
				os.Stdout.Write(data)
			},
			OnStderr: func(data []byte) {
				os.Stderr.Write(data)
			},
		})
		if err != nil {
			errCh <- err
			return
		}
		exitCh <- result.ExitCode
	}()

	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, readErr := os.Stdin.Read(buf)
			if n > 0 {
				if sendErr := handle.SendStdin(ctx, buf[:n]); sendErr != nil {
					logf("stdin send error: %v", sendErr)
					break
				}
			}
			if readErr != nil {
				logf("stdin closed, terminating sandbox")
				handle.CloseStdin(context.Background())
				cancel()
				break
			}
		}
	}()

	select {
	case code := <-exitCh:
		logf("process exited with code %d", code)
		killSandbox()
		os.Exit(code)
	case err := <-errCh:
		fmt.Fprintf(os.Stderr, "[declaw] stream error: %v\n", err)
		killSandbox()
		os.Exit(125)
	case <-ctx.Done():
		handle.Kill(context.Background())
		killSandbox()
		os.Exit(125)
	}

	return nil
}

func buildShellCommand(args []string) string {
	if len(args) == 1 {
		return args[0]
	}
	quoted := make([]string, len(args))
	for i, a := range args {
		if a == "" || strings.ContainsAny(a, " \t\n\r\"'\\$`!&|;(){}[]<>?*#~") {
			quoted[i] = "'" + strings.ReplaceAll(a, "'", "'\"'\"'") + "'"
		} else {
			quoted[i] = a
		}
	}
	return strings.Join(quoted, " ")
}
