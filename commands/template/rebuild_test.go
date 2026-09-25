package template

import (
	"io"
	"strings"
	"testing"
)

func TestRebuildCmd_RequiresTemplateID(t *testing.T) {
	cmd := newRebuildCmd()
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "1 arg") {
		t.Fatalf("err = %v, want an exact-args error", err)
	}
}

func TestRebuildCmd_HasNoWaitFlag(t *testing.T) {
	cmd := newRebuildCmd()
	if err := cmd.ParseFlags([]string{"--no-wait"}); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	noWait, err := cmd.Flags().GetBool("no-wait")
	if err != nil || !noWait {
		t.Fatalf("no-wait = %v, %v; want true", noWait, err)
	}
}

// The subcommand is registered under "template" with the same name the
// sandbox-manager's 409 hint tells users to run.
func TestTemplateCmd_RegistersRebuild(t *testing.T) {
	root := NewTemplateCmd()
	for _, c := range root.Commands() {
		if c.Name() == "rebuild" {
			return
		}
	}
	t.Fatal("template command has no rebuild subcommand")
}
