// Package ui is the Bubble Tea front-end. It talks to the domain.Board and
// persists through a domain.Repository; it contains no business rules.
package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"kanban/internal/domain"
)

type mode int

const (
	modeNormal mode = iota
	modeForm
	modeConfirm
)

type dragState struct {
	pressID int         // task under the mouse button, 0 = none
	active  bool        // pointer moved since the press
	over    domain.Zone // zone under the pointer
	x, y    int
}

// Model implements tea.Model.
type Model struct {
	board *domain.Board
	repo  domain.Repository
	now   func() time.Time

	w, h   int
	zone   domain.Zone // focused zone
	idx    int         // focused row inside it
	scroll [domain.NumZones]int
	mode   mode
	msg    string // transient error shown in the status line

	drag dragState
	form formState
}

func New(board *domain.Board, repo domain.Repository) *Model {
	return &Model{board: board, repo: repo, now: time.Now, drag: dragState{over: noZone}}
}

func (m *Model) Init() tea.Cmd { return nil }

// ── helpers ──────────────────────────────────────────────────────────────

func (m *Model) list(z domain.Zone) []*domain.Task { return m.board.InZone(z) }

func (m *Model) cur() *domain.Task {
	if l := m.list(m.zone); m.idx >= 0 && m.idx < len(l) {
		return l[m.idx]
	}
	return nil
}

// persist saves the board and surfaces failures in the status line.
func (m *Model) persist() {
	if err := m.repo.Save(m.board); err != nil {
		m.msg = "save failed: " + err.Error()
	}
}

// fail records a domain error for display.
func (m *Model) fail(err error) bool {
	if err != nil {
		m.msg = err.Error()
	}
	return err != nil
}

// clamp keeps the cursor and every zone's scroll offset inside their lists.
func (m *Model) clamp() {
	m.idx = max(0, min(m.idx, len(m.list(m.zone))-1))
	if m.w < 1 {
		return
	}
	lay := computeLayout(m.w, m.h)
	for z := domain.Zone(0); z < domain.NumZones; z++ {
		rows := max(1, lay.zone[z].h-2)
		m.scroll[z] = max(0, min(m.scroll[z], len(m.list(z))-rows))
		if z == m.zone {
			m.scroll[z] = max(m.scroll[z], m.idx-rows+1)
			m.scroll[z] = min(m.scroll[z], m.idx)
		}
	}
}

// focus moves the cursor to the task with the given id, wherever it is now.
func (m *Model) focus(id int) {
	for z := domain.Zone(0); z < domain.NumZones; z++ {
		for i, t := range m.list(z) {
			if t.ID == id {
				m.zone, m.idx = z, i
				m.clamp()
				return
			}
		}
	}
}

func (m *Model) moveTask(id int, z domain.Zone) {
	if m.fail(m.board.Move(id, z)) {
		return
	}
	m.persist()
	m.focus(id)
}

// ── update ───────────────────────────────────────────────────────────────

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.clamp()
	case tea.MouseMsg:
		m.onMouse(msg)
	case tea.KeyMsg:
		m.msg = ""
		switch m.mode {
		case modeForm:
			return m, m.onFormKey(msg)
		case modeConfirm:
			m.onConfirmKey(msg)
		default:
			return m, m.onKey(msg)
		}
	}
	return m, nil
}
