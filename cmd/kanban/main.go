// Command kanban is a terminal Kanban board with an Eisenhower matrix backlog.
package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"kanban/internal/storage"
	"kanban/internal/ui"
)

func main() {
	path := flag.String("f", storage.DefaultPath(), "path of the tasks file")
	flag.Parse()

	repo := storage.NewJSONFile(*path)
	board, err := repo.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "kanban:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.New(board, repo), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "kanban:", err)
		os.Exit(1)
	}
}
