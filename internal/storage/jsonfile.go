// Package storage implements domain.Repository on top of a JSON file.
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"kanban/internal/domain"
)

type fileFormat struct {
	Next  int            `json:"next"`
	Tasks []*domain.Task `json:"tasks"`
}

// JSONFile stores the board as one JSON document.
type JSONFile struct{ path string }

func NewJSONFile(path string) *JSONFile { return &JSONFile{path: path} }

// DefaultPath is <user config dir>/kanban-tui/tasks.json.
func DefaultPath() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "kanban-tui", "tasks.json")
	}
	return "tasks.json"
}

// Load reads the board; a missing file yields an empty board.
// A corrupt file is an error so it never gets silently overwritten.
func (f *JSONFile) Load() (*domain.Board, error) {
	raw, err := os.ReadFile(f.path)
	if errors.Is(err, fs.ErrNotExist) {
		return domain.NewBoard(1, nil), nil
	}
	if err != nil {
		return nil, err
	}
	var ff fileFormat
	if err := json.Unmarshal(raw, &ff); err != nil {
		return nil, fmt.Errorf("parse %s: %w", f.path, err)
	}
	return domain.NewBoard(ff.Next, ff.Tasks), nil
}

// Save writes atomically (temp file + rename).
func (f *JSONFile) Save(b *domain.Board) error {
	if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
		return err
	}
	next, tasks := b.Export()
	raw, err := json.MarshalIndent(fileFormat{Next: next, Tasks: tasks}, "", "  ")
	if err != nil {
		return err
	}
	tmp := f.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, f.path)
}
