package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/declaw-ai/declaw-cli/internal/cmdutil"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newConnectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "connect <sandbox-id>",
		Short: "Open an interactive terminal in a sandbox",
		Args:  cobra.ExactArgs(1),
		RunE:  runConnect,
	}
}

func runConnect(cmd *cobra.Command, args []string) error {
	cfg, err := cmdutil.ResolveConfig(cmd)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Connecting to %s...\n", args[0])

	sbx, err := declaw.Connect(context.Background(), args[0], cmdutil.SandboxOpts(cfg)...)
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		width, height = 80, 24
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pty, err := sbx.PTY.Create(ctx, declaw.PtySize{Cols: width, Rows: height})
	if err != nil {
		cmdutil.HandleError(err)
		return nil
	}

	stream, err := pty.Stream(ctx)
	if err != nil {
		return fmt.Errorf("starting PTY stream: %w", err)
	}

	if err := pty.SendInput(ctx, []byte("\n")); err != nil {
		return fmt.Errorf("sending initial input: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Connected. Press Ctrl-D or type 'exit' to disconnect.\n")

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("setting terminal to raw mode: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	go func() {
		for {
			select {
			case <-sigCh:
				w, h, err := term.GetSize(int(os.Stdin.Fd()))
				if err == nil {
					pty.SetSize(ctx, w, h)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		for data := range stream {
			os.Stdout.Write(data)
		}
		cancel()
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, readErr := os.Stdin.Read(buf)
			if n > 0 {
				if sendErr := pty.SendInput(ctx, buf[:n]); sendErr != nil {
					cancel()
					return
				}
			}
			if readErr != nil {
				cancel()
				return
			}
		}
	}()

	<-ctx.Done()
	signal.Stop(sigCh)
	return nil
}
