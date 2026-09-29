package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"kanban/internal/domain"
)

// Neighbour zone when moving horizontally through the 2×2 matrix + columns.
var (
	hopRight = [domain.NumZones]domain.Zone{
		domain.ZoneDo: domain.ZonePlan, domain.ZonePlan: domain.ZoneDoing,
		domain.ZoneDelegate: domain.ZoneDrop, domain.ZoneDrop: domain.ZoneDoing,
		domain.ZoneDoing: domain.ZoneDone, domain.ZoneDone: domain.ZoneDone,
	}
	hopLeft = [domain.NumZones]domain.Zone{
		domain.ZoneDo: domain.ZoneDo, domain.ZonePlan: domain.ZoneDo,
		domain.ZoneDelegate: domain.ZoneDelegate, domain.ZoneDrop: domain.ZoneDelegate,
		domain.ZoneDoing: domain.ZonePlan, domain.ZoneDone: domain.ZoneDoing,
	}
)

func (m *Model) hop(dir int) {
	if dir > 0 {
		m.zone = hopRight[m.zone]
	} else {
		m.zone = hopLeft[m.zone]
	}
}

// step moves the cursor one row; at the edge of the top/bottom quadrants it
// crosses into the quadrant above/below.
func (m *Model) step(dir int) {
	n := len(m.list(m.zone))
	switch {
	case dir < 0 && m.idx > 0:
		m.idx--
	case dir < 0 && (m.zone == domain.ZoneDelegate || m.zone == domain.ZoneDrop):
		m.zone -= 2
		m.idx = max(0, len(m.list(m.zone))-1)
	case dir > 0 && m.idx < n-1:
		m.idx++
	case dir > 0 && (m.zone == domain.ZoneDo || m.zone == domain.ZonePlan):
		m.zone += 2
		m.idx = 0
	}
}

func (m *Model) onKey(k tea.KeyMsg) tea.Cmd {
	key := normKey(k.String())
	t := m.cur()
	switch key {
	case "q", "ctrl+c":
		return tea.Quit
	case "left", "h":
		m.hop(-1)
	case "right", "l":
		m.hop(1)
	case "up", "k":
		m.step(-1)
	case "down", "j":
		m.step(1)
	case "a", "n":
		m.openForm(nil)
	case "e":
		if t != nil {
			m.openForm(t)
		}
	case "d", "delete":
		if t != nil {
			m.mode = modeConfirm
		}
	case "1", "2", "3", "4", "5", "6":
		if t != nil {
			m.moveTask(t.ID, domain.Zone(key[0]-'1'))
		}
	case "enter", " ":
		if t != nil && t.Zone.Advance() != t.Zone {
			m.moveTask(t.ID, t.Zone.Advance())
		}
	case "[", "]":
		if t != nil {
			delta := map[string]int{"[": -1, "]": 1}[key]
			if !m.fail(m.board.Reprioritize(t.ID, delta)) {
				m.persist()
				m.focus(t.ID)
			}
		}
	}
	m.clamp()
	return nil
}

func (m *Model) onConfirmKey(k tea.KeyMsg) {
	if normKey(k.String()) == "y" {
		if t := m.cur(); t != nil && !m.fail(m.board.Delete(t.ID)) {
			m.persist()
		}
	}
	m.mode = modeNormal
	m.clamp()
}

func (m *Model) taskAt(l layout, z domain.Zone, y int) *domain.Task {
	r := l.zone[z]
	row := y - (r.y + 1)
	if row < 0 || row >= r.h-2 {
		return nil
	}
	if lst, i := m.list(z), row+m.scroll[z]; i < len(lst) {
		return lst[i]
	}
	return nil
}

func (m *Model) onMouse(e tea.MouseMsg) {
	if m.mode != modeNormal || m.w < 1 {
		return
	}
	lay := computeLayout(m.w, m.h)
	z := lay.zoneAt(e.X, e.Y)
	switch {
	case e.Button == tea.MouseButtonWheelUp || e.Button == tea.MouseButtonWheelDown:
		if z == noZone {
			return
		}
		m.zone = z
		if e.Button == tea.MouseButtonWheelUp {
			m.idx--
		} else {
			m.idx++
		}
		m.clamp()

	case e.Action == tea.MouseActionPress && e.Button == tea.MouseButtonLeft:
		m.drag = dragState{over: noZone}
		if z == noZone {
			return
		}
		m.zone = z
		if t := m.taskAt(lay, z, e.Y); t != nil {
			m.focus(t.ID)
			m.drag.pressID = t.ID
		}
		m.clamp()

	case e.Action == tea.MouseActionMotion && m.drag.pressID != 0:
		m.drag.active, m.drag.over, m.drag.x, m.drag.y = true, z, e.X, e.Y

	case e.Action == tea.MouseActionRelease && m.drag.pressID != 0:
		id := m.drag.pressID
		if t := m.board.Get(id); t != nil && m.drag.active && z != noZone && z != t.Zone {
			m.moveTask(id, z)
		}
		m.drag = dragState{over: noZone}
	}
}
