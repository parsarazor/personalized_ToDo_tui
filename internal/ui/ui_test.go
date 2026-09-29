package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"kanban/internal/domain"
)

// memRepo is an in-memory domain.Repository that counts saves.
type memRepo struct{ saves int }

func (r *memRepo) Load() (*domain.Board, error) { return domain.NewBoard(1, nil), nil }
func (r *memRepo) Save(*domain.Board) error     { r.saves++; return nil }

func newTestModel(w, h int) (*Model, *memRepo) {
	repo := &memRepo{}
	m := New(domain.NewBoard(1, nil), repo)
	m.now = func() time.Time { return time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC) }
	m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return m, repo
}

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func (m *Model) addViaForm(title, deadline, prio string, zoneRight int) {
	m.Update(key("a"))
	m.Update(key(title))
	m.Update(key("tab"))
	m.Update(key(deadline))
	m.Update(key("tab"))
	m.Update(key(prio))
	m.Update(key("tab"))
	for i := 0; i < zoneRight; i++ {
		m.Update(key("right"))
	}
	m.Update(key("enter"))
}

func TestAddViaFormAndRender(t *testing.T) {
	m, repo := newTestModel(140, 40)
	m.addViaForm("write report", "+1", "2", 0)
	m.addViaForm("سلام دنیا", "12-31", "۱", 1) // Persian digit priority

	if m.mode != modeNormal {
		t.Fatalf("form still open: %q", m.form.err)
	}
	if got := len(m.list(domain.ZoneDo)); got != 1 {
		t.Fatalf("Do zone has %d tasks", got)
	}
	if plan := m.list(domain.ZonePlan); len(plan) != 1 || plan[0].Priority != 1 {
		t.Fatalf("Plan zone wrong: %+v", plan)
	}
	if repo.saves != 2 {
		t.Fatalf("saves = %d", repo.saves)
	}
	v := m.View()
	if !strings.Contains(v, "write report") || !strings.Contains(v, "DO NOW") {
		t.Fatal("render is missing content")
	}
	if n := len(strings.Split(v, "\n")); n != 40 {
		t.Fatalf("rendered %d lines; want 40", n)
	}
}

func TestFormRejectsBadInput(t *testing.T) {
	m, _ := newTestModel(140, 40)
	m.addViaForm("x", "not-a-date", "", 0)
	if m.mode != modeForm || m.form.field != fieldDeadline || m.form.err == "" {
		t.Fatal("bad deadline should keep the form open on the deadline field")
	}
	m.Update(key("esc"))
	m.addViaForm("", "", "", 0)
	if m.mode != modeForm || m.form.field != fieldTitle {
		t.Fatal("empty title should be rejected")
	}
}

func TestMouseDragAcrossZones(t *testing.T) {
	m, _ := newTestModel(140, 40)
	m.addViaForm("drag me", "", "", 0)
	lay := computeLayout(140, 40)
	src, dst := lay.zone[domain.ZoneDo], lay.zone[domain.ZoneDone]
	id := m.list(domain.ZoneDo)[0].ID

	m.Update(tea.MouseMsg{X: src.x + 4, Y: src.y + 1, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	m.Update(tea.MouseMsg{X: dst.x + 3, Y: dst.y + 3, Action: tea.MouseActionMotion, Button: tea.MouseButtonLeft})
	if v := m.View(); !strings.Contains(v, "drop") {
		t.Fatal("status line should announce the drop target while dragging")
	}
	m.Update(tea.MouseMsg{X: dst.x + 3, Y: dst.y + 3, Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft})

	if m.board.Get(id).Zone != domain.ZoneDone {
		t.Fatalf("zone = %d after drag", m.board.Get(id).Zone)
	}
	// a plain click (no motion) must not move anything
	m.Update(tea.MouseMsg{X: dst.x + 3, Y: dst.y + 1, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	m.Update(tea.MouseMsg{X: src.x + 3, Y: src.y + 1, Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft})
	if m.board.Get(id).Zone != domain.ZoneDone {
		t.Fatal("click without motion moved the task")
	}
}

func TestKeyboardFlow(t *testing.T) {
	m, _ := newTestModel(140, 40)
	m.addViaForm("job", "", "", 0)

	m.Update(key("enter")) // backlog → doing
	if m.cur() == nil || m.cur().Zone != domain.ZoneDoing {
		t.Fatal("enter should advance to Doing")
	}
	m.Update(key("3")) // straight to quadrant 3
	if m.cur().Zone != domain.ZoneDelegate {
		t.Fatal("key 3 should move to DELEGATE")
	}
	m.Update(key("ش")) // Persian-layout 'a' opens the form
	if m.mode != modeForm {
		t.Fatal("Persian layout shortcut not recognised")
	}
	m.Update(key("esc"))
	m.Update(key("d"))
	m.Update(key("y"))
	if len(m.list(domain.ZoneDelegate)) != 0 {
		t.Fatal("task not deleted")
	}
}

func TestViewNeverPanicsOnAnySize(t *testing.T) {
	m, _ := newTestModel(140, 40)
	for i := 0; i < 30; i++ {
		m.addViaForm("task with a fairly long title to force truncation", "+2", "3", i%int(domain.NumZones))
	}
	for _, s := range [][2]int{{1, 1}, {10, 5}, {64, 18}, {80, 24}, {300, 90}} {
		m.Update(tea.WindowSizeMsg{Width: s[0], Height: s[1]})
		_ = m.View()
		m.Update(key("j"))
		m.Update(key("right"))
	}
}
