package template

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestBuildCmd_AliasRequired(t *testing.T) {
	cmd := newBuildCmd()
	cmd.SetArgs([]string{"--base-image", "ubuntu:22.04"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), `"alias"`) {
		t.Fatalf("err = %v, want a missing --alias error", err)
	}
}

func TestSpecFromFlags(t *testing.T) {
	cmd := newBuildCmd()
	err := cmd.ParseFlags([]string{
		"--alias", "my-tools",
		"--base-image", "python:3.12",
		"--run-cmd", "pip install numpy",
		"--apt-package", "ffmpeg",
		"--apt-package", "jq",
		"--start-cmd", "python main.py",
		"--disk-mb", "2048",
	})
	if err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}

	spec, err := specFromFlags(cmd)
	if err != nil {
		t.Fatalf("specFromFlags: %v", err)
	}
	if spec.Alias != "my-tools" || spec.BaseImage != "python:3.12" || spec.StartCmd != "python main.py" || spec.DiskMB != 2048 {
		t.Errorf("spec = %+v", spec)
	}
	if !reflect.DeepEqual(spec.RunCmds, []string{"pip install numpy"}) {
		t.Errorf("RunCmds = %v", spec.RunCmds)
	}
	if !reflect.DeepEqual(spec.AptPackages, []string{"ffmpeg", "jq"}) {
		t.Errorf("AptPackages = %v", spec.AptPackages)
	}
}

// Each --run-cmd is one shell command, verbatim: commas and quotes are
// ordinary shell text, not list separators. Package names cannot contain a
// comma, so --apt-package keeps accepting comma-separated lists.
func TestSpecFromFlags_RunCmdsAreVerbatim(t *testing.T) {
	cmd := newBuildCmd()
	err := cmd.ParseFlags([]string{
		"--alias", "my-tools",
		"--run-cmd", "echo a,b",
		"--run-cmd", `pip install "numpy>=1.24,<2"`,
		"--apt-package", "curl,git",
		"--apt-package", "jq",
	})
	if err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}

	spec, err := specFromFlags(cmd)
	if err != nil {
		t.Fatalf("specFromFlags: %v", err)
	}
	if want := []string{"echo a,b", `pip install "numpy>=1.24,<2"`}; !reflect.DeepEqual(spec.RunCmds, want) {
		t.Errorf("RunCmds = %q, want %q", spec.RunCmds, want)
	}
	if want := []string{"curl", "git", "jq"}; !reflect.DeepEqual(spec.AptPackages, want) {
		t.Errorf("AptPackages = %q, want %q", spec.AptPackages, want)
	}
}

func TestSpecFromFlags_Dockerfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Dockerfile")
	const content = "FROM ubuntu:22.04\nRUN apt-get update\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := newBuildCmd()
	if err := cmd.ParseFlags([]string{"--alias", "my-mcp", "--dockerfile", path}); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	spec, err := specFromFlags(cmd)
	if err != nil {
		t.Fatalf("specFromFlags: %v", err)
	}
	if spec.Dockerfile != content || spec.Alias != "my-mcp" {
		t.Errorf("spec = %+v", spec)
	}

	cmd = newBuildCmd()
	if err := cmd.ParseFlags([]string{"--alias", "x", "--dockerfile", filepath.Join(t.TempDir(), "missing")}); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	if _, err := specFromFlags(cmd); err == nil {
		t.Error("expected an error for a missing Dockerfile")
	}
}
