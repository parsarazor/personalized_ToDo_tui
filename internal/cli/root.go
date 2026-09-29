// Package cli provides the Cobra command surface for the kanban TUI.
//
// This package is purely additive: the legacy flag-based entrypoint in
// cmd/kanban/main.go is left untouched. The global binaries defined under
// cmd/todo and cmd/taskboard are thin wrappers around Execute, so typing
// `todo` or `taskboard` from any directory launches the same board.
package cli

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"kanban/internal/storage"
	"kanban/internal/ui"
)

// envKeys lists, in priority order, the environment variables that can
// override the default tasks-file location. The generic KANBAN_FILE wins
// so one export covers every binary; the per-binary names are conveniences.
var envKeys = []string{"KANBAN_FILE", "TODO_FILE", "TASKBOARD_FILE"}

// filePathFromEnv returns the first non-empty override from envKeys,
// or "" when none is set.
func filePathFromEnv() string {
	for _, k := range envKeys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

// DefaultFilePath resolves the tasks file in the same way every binary
// does: explicit --file/-f flag wins (handled by cobra), otherwise the
// environment override wins, otherwise storage.DefaultPath().
func DefaultFilePath() string {
	if v := filePathFromEnv(); v != "" {
		return v
	}
	return storage.DefaultPath()
}

// RunBoard loads the board from filePath and runs the Bubble Tea program.
// It mirrors the behavior of the legacy cmd/kanban entrypoint.
func RunBoard(filePath string) error {
	repo := storage.NewJSONFile(filePath)
	board, err := repo.Load()
	if err != nil {
		return fmt.Errorf("load tasks file: %w", err)
	}
	p := tea.NewProgram(ui.New(board, repo), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run TUI: %w", err)
	}
	return nil
}

// NewRootCmd builds the root command for one global binary.
// use is the binary name ("todo", "taskboard", ...), short is a one-line
// description, version is printed by --version and the version subcommand
// (inject with: -ldflags "-X main.version=$(git describe --tags --always --dirty)").
func NewRootCmd(use, short, version string) *cobra.Command {
	if version == "" {
		version = "dev"
	}
	var file string

	root := &cobra.Command{
		Use:     use,
		Short:   short,
		Version: version,
		Args:    cobra.NoArgs,
		// Runtime failures (bad tasks file, TUI error) must not dump
		// usage text; main() already prints the error with a prefix.
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: fmt.Sprintf(`%s

Launches the terminal Kanban board (Eisenhower-matrix backlog).
Tasks are stored in %s unless overridden.

Environment:
  KANBAN_FILE        tasks file used by every binary (highest priority)
  TODO_FILE          alias override read by the todo binary
  TASKBOARD_FILE     alias override read by the taskboard binary`, short, storage.DefaultPath()),
		Example: fmt.Sprintf(`  %s
  %s -f /tmp/tasks.json
  %s --file ~/tasks/work.json
  %s run --file ./tasks.json
  %s version`, use, use, use, use, use),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return RunBoard(file)
		},
	}

	// Same semantics as the legacy `-f` flag: shorthand -f, same help text,
	// same default. Persistent so `todo run -f X` works too.
	root.PersistentFlags().StringVarP(&file, "file", "f", DefaultFilePath(), "path of the tasks file")

	root.AddCommand(newRunCmd(use, &file))
	root.AddCommand(newVersionCmd(use, version))

	return root
}

// newRunCmd is the explicit form of launching the board:
// `todo run -f FILE`. It reuses the root's --file value via closure,
// so the flag is defined exactly once (on the root).
func newRunCmd(use string, file *string) *cobra.Command {
	return &cobra.Command{
		Use:     "run",
		Aliases: []string{"open", "board"},
		Short:   "Open the Kanban board",
		Long:    fmt.Sprintf("Open the Kanban board.\n\nSame as running `%s` with no arguments.", use),
		Example: fmt.Sprintf(`  %s run
  %s run -f /tmp/tasks.json`, use, use),
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return RunBoard(*file)
		},
	}
}

// newVersionCmd prints the binary name plus its build version.
func newVersionCmd(use, version string) *cobra.Command {
	return &cobra.Command{
		Use:           "version",
		Short:         "Print the version",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", use, version)
			return err
		},
	}
}

// Execute builds the root command for use and runs it.
// It returns the error so each cmd/*/main.go can print it prefixed
// with its own binary name (matching the legacy "kanban: ..." style).
func Execute(use, short, version string) error {
	return NewRootCmd(use, short, version).Execute()
}
