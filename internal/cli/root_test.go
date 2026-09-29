package cli

import (
	"testing"

	"kanban/internal/storage"
)

func TestNewRootCmdDefaults(t *testing.T) {
	cmd := NewRootCmd("todo", "todo — terminal Kanban board", "dev")
	if got := cmd.Use; got != "todo" {
		t.Fatalf("Use = %q, want %q", got, "todo")
	}
	f, err := cmd.PersistentFlags().GetString("file")
	if err != nil {
		t.Fatalf("read --file flag: %v", err)
	}
	if f == "" {
		t.Fatal("--file default must not be empty")
	}
	if want := DefaultFilePath(); f != want {
		t.Fatalf("--file default = %q, want %q", f, want)
	}
	if cmd.Version != "dev" {
		t.Fatalf("Version = %q, want dev", cmd.Version)
	}
	if c := cmd.Commands(); len(c) != 2 {
		t.Fatalf("expected 2 subcommands (run, version), got %d", len(c))
	}
}

func TestDefaultFilePathEnvOverride(t *testing.T) {
	t.Setenv("KANBAN_FILE", "/tmp/kanban-override.json")
	t.Setenv("TODO_FILE", "")
	t.Setenv("TASKBOARD_FILE", "")
	if got := DefaultFilePath(); got != "/tmp/kanban-override.json" {
		t.Fatalf("DefaultFilePath = %q, want override", got)
	}
}

func TestDefaultFilePathFallsBackToStorageDefault(t *testing.T) {
	t.Setenv("KANBAN_FILE", "")
	t.Setenv("TODO_FILE", "")
	t.Setenv("TASKBOARD_FILE", "")
	if got, want := DefaultFilePath(), storage.DefaultPath(); got != want {
		t.Fatalf("DefaultFilePath = %q, want %q", got, want)
	}
}

func TestVersionSubcommandPrintsNameAndVersion(t *testing.T) {
	cmd := NewRootCmd("taskboard", "taskboard — terminal Kanban board", "v1.2.3")
	cmd.SetArgs([]string{"version"})
	out, err := captureOutput(cmd)
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if out != "taskboard v1.2.3\n" {
		t.Fatalf("version output = %q, want %q", out, "taskboard v1.2.3\n")
	}
}

func TestHelpMentionsRunAndEnv(t *testing.T) {
	cmd := NewRootCmd("todo", "todo — terminal Kanban board", "dev")
	cmd.SetArgs([]string{"--help"})
	out, err := captureOutput(cmd)
	if err != nil {
		t.Fatalf("--help: %v", err)
	}
	for _, want := range []string{"run", "KANBAN_FILE", "--file"} {
		if !contains(out, want) {
			t.Fatalf("--help output missing %q:\n%s", want, out)
		}
	}
}
