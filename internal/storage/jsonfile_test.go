package storage

import (
	"os"
	"path/filepath"
	"testing"

	"kanban/internal/domain"
)

func TestRoundTrip(t *testing.T) {
	repo := NewJSONFile(filepath.Join(t.TempDir(), "sub", "tasks.json"))

	b, err := repo.Load() // missing file → empty board
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.Add(domain.Input{Title: "سلام", Deadline: "2026-10-05", Priority: 2, Zone: domain.ZonePlan}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(b); err != nil {
		t.Fatal(err)
	}

	b2, err := repo.Load()
	if err != nil {
		t.Fatal(err)
	}
	got := b2.InZone(domain.ZonePlan)
	if len(got) != 1 || got[0].Title != "سلام" || got[0].Priority != 2 || got[0].Deadline != "2026-10-05" {
		t.Fatalf("unexpected: %+v", got)
	}
	if tk, _ := b2.Add(domain.Input{Title: "next"}); tk.ID != 2 {
		t.Fatalf("id counter lost: %d", tk.ID)
	}
}

func TestCorruptFileIsAnErrorNotOverwritten(t *testing.T) {
	p := filepath.Join(t.TempDir(), "tasks.json")
	_ = os.WriteFile(p, []byte("{not json"), 0o644)
	if _, err := NewJSONFile(p).Load(); err == nil {
		t.Fatal("expected error for corrupt file")
	}
	if raw, _ := os.ReadFile(p); string(raw) != "{not json" {
		t.Fatal("file was modified")
	}
}
