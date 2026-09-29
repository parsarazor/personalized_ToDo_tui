#!/bin/sh
# Installs the `todo` and `taskboard` commands globally (new file only —
# no existing project file is touched).
#
# Usage:
#   sh scripts/install.sh                 # installs both binaries
#   sh scripts/install.sh --with-completions
#   VERSION=v1.0.0 sh scripts/install.sh  # stamps --version / `version`
#
# After installing, `todo` / `taskboard` work from ANY directory because
# the binaries live on PATH (~/go/bin by default) and the tasks file
# defaults to ~/.config/kanban-tui/tasks.json (override with -f or $KANBAN_FILE).
set -eu

cd "$(dirname "$0")/.."

VERSION="${VERSION:-dev}"
GOBIN_DIR="$(go env GOBIN)"
if [ -z "$GOBIN_DIR" ] || [ "$GOBIN_DIR" = "" ]; then
  GOBIN_DIR="$(go env GOPATH)/bin"
fi

echo "→ building todo@$VERSION and taskboard@$VERSION into $GOBIN_DIR"
go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o "$GOBIN_DIR/todo" ./cmd/todo
go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o "$GOBIN_DIR/taskboard" ./cmd/taskboard

if [ "${1:-}" = "--with-completions" ]; then
  echo "→ installing shell completions"
  mkdir -p ~/.zsh/completions ~/.bash_completion.d ~/.config/fish/completions
  "$GOBIN_DIR/todo" completion zsh > ~/.zsh/completions/_todo
  "$GOBIN_DIR/todo" completion bash > ~/.bash_completion.d/todo
  "$GOBIN_DIR/todo" completion fish > ~/.config/fish/completions/todo.fish
  "$GOBIN_DIR/taskboard" completion zsh > ~/.zsh/completions/_taskboard
  "$GOBIN_DIR/taskboard" completion bash > ~/.bash_completion.d/taskboard
  "$GOBIN_DIR/taskboard" completion fish > ~/.config/fish/completions/taskboard.fish
  echo "  completions written (restart shell or source them)."
fi

case ":$PATH:" in
  *":$GOBIN_DIR:"*) ;;
  *)
    echo ""
    echo "NOTE: $GOBIN_DIR is not on your PATH."
    echo "Add this line to ~/.zshrc (or ~/.bashrc):"
    echo "  export PATH=\"\$PATH:$GOBIN_DIR\""
    ;;
esac

echo ""
echo "✓ done. Try from any directory:"
echo "    todo            # open the board"
echo "    taskboard -f /tmp/tasks.json"
echo "    todo version    # $VERSION"
