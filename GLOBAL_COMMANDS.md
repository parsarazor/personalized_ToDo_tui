# Global commands: `todo` and `taskboard`

> New capability, zero changes to existing code. The original
> `cmd/kanban` entrypoint, `Makefile`, `README.md`, and everything under
> `internal/` are untouched — this feature lives entirely in the new files
> `internal/cli/`, `cmd/todo/`, `cmd/taskboard/`, and `scripts/install.sh`
> (plus the unavoidable `go.mod`/`go.sum` entries for the Cobra dependency).

## One-time install

```sh
# from the repo root — installs both binaries to ~/go/bin (already on your PATH)
sh scripts/install.sh

# with shell completions (bash/zsh/fish)
sh scripts/install.sh --with-completions

# stamp a version string (shows in `todo version` / `todo --version`)
VERSION=v1.0.0 sh scripts/install.sh
```

Manual equivalent (same thing the script does):

```sh
go install ./cmd/todo ./cmd/taskboard
```

## Run from anywhere

```sh
todo                       # open the board (no args needed)
taskboard                  # same board, alternate name
todo -f /tmp/tasks.json    # use a different tasks file
taskboard --file ~/work.json
todo run                   # explicit form (aliases: open, board)
todo run -f ./tasks.json
todo version / todo --version
todo completion zsh        # free shell completions from Cobra
```

## Tasks-file resolution (all binaries agree)

1. `--file` / `-f` flag (same shorthand + help text as legacy `cmd/kanban -f`)
2. `$KANBAN_FILE` env var (highest-priority env; one export covers every binary)
3. `$TODO_FILE` / `$TASKBOARD_FILE` (per-binary convenience overrides)
4. `~/.config/kanban-tui/tasks.json` (same `storage.DefaultPath()` default)

```sh
export KANBAN_FILE=~/tasks.json   # todo & taskboard both use it
export TODO_FILE=/tmp/t.json      # only the todo binary
```

## Design notes (expert-level choices)

- **Single shared root** (`internal/cli.NewRootCmd(use, short, version)`) parameterized
  by binary name — `cmd/todo/main.go` and `cmd/taskboard/main.go` are ~10-line
  wrappers. Adding a third alias (e.g. `tasks`) is one new folder + one call.
- **Persistent `--file` flag** defined once on the root, inherited by `run`,
  so `todo -f X` and `todo run -f X` behave identically.
- **`SilenceUsage: true`** on every command: a corrupt tasks file or TUI error
  prints `todo: <err>` (matching the legacy `kanban: ...` style) without
  dumping usage text — the standard Cobra CLI convention.
- **Version injection** via `-X main.version=...` keeps reproducible builds
  without code changes; defaults to `dev`.
- **Tests** (`internal/cli/*_test.go`) cover flag defaults, env precedence,
  `version` output, and help text — run with `go test ./internal/cli/`.
