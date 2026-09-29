package ui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"kanban/internal/domain"
)

const (
	minWidth  = 64
	minHeight = 18
)

func (m *Model) View() string {
	if m.w == 0 {
		return ""
	}
	if m.w < minWidth || m.h < minHeight {
		return fmt.Sprintf("terminal too small — need at least %d×%d", minWidth, minHeight)
	}
	lay := computeLayout(m.w, m.h)
	cv := newCanvas(m.w, m.h)

	m.drawHeader(cv)
	cv.box(lay.backlog, borderRounded, colEdge, "BACKLOG", "", -1)
	for z := domain.Zone(0); z < domain.NumZones; z++ {
		m.drawZone(cv, lay, z)
	}
	m.drawStatus(cv)
	m.drawDragGhost(cv)
	if m.mode == modeForm {
		m.drawForm(cv)
	}
	return cv.String()
}

func (m *Model) drawHeader(cv *canvas) {
	backlog := 0
	for z := domain.ZoneDo; z <= domain.ZoneDrop; z++ {
		backlog += len(m.list(z))
	}
	cv.put(1, 0, "KANBAN", colBone, colBg, true, m.w)
	cv.put(9, 0, fmt.Sprintf("%d backlog · %d doing · %d done",
		backlog, len(m.list(domain.ZoneDoing)), len(m.list(domain.ZoneDone))), colDim, colBg, false, m.w)
	date := m.now().Format("Mon 02 Jan 2006")
	cv.put(m.w-1-len(date), 0, date, colDim, colBg, false, m.w)
}

func (m *Model) drawZone(cv *canvas, lay layout, z domain.Zone) {
	r := lay.zone[z]
	zs := zoneStyles[z]
	focused := z == m.zone && m.mode == modeNormal

	col, kind := zs.border, borderRounded
	switch {
	case m.drag.active && m.drag.over == z:
		col, kind = colBone, borderDouble
	case focused:
		col, kind = zs.accent, borderHeavy
	}
	list := m.list(z)
	cv.box(r, kind, col, zs.name, zs.sub, len(list))

	rows := r.h - 2
	if len(list) == 0 && rows > 0 {
		cv.put(r.x+2, r.y+1, "empty", colDim, colBg, false, r.x+r.w-2)
	}
	textL, textR := r.x+2, r.x+r.w-2
	now := m.now()
	for row := 0; row < rows; row++ {
		i := row + m.scroll[z]
		if i >= len(list) {
			break
		}
		t, y := list[i], r.y+1+row
		selected := focused && i == m.idx
		bg := colBg
		if selected {
			bg = colSel
			cv.paintBg(y, r.x+1, r.x+r.w-1, bg)
		}
		fg := colBone
		if z == domain.ZoneDone {
			fg = colDim
		}

		prio := "·"
		if t.Priority > 0 {
			prio = strconv.Itoa(t.Priority)
		}
		cv.put(textL, y, prio, zs.accent, bg, true, textR)

		right := textR
		if t.Deadline != "" {
			ds := domain.FormatShort(t.Deadline, now)
			right = textR - runewidth.StringWidth(ds)
			cv.put(right, y, ds, deadlineColor(t, now), bg, false, textR)
			right--
		}
		title := runewidth.Truncate(t.Title, max(0, right-(textL+2)), "…")
		cv.put(textL+2, y, title, fg, bg, selected, right)
	}
	if m.scroll[z] > 0 {
		cv.put(r.x+r.w-2, r.y+1, "▲", colDim, colBg, false, r.x+r.w)
	}
	if m.scroll[z]+rows < len(list) {
		cv.put(r.x+r.w-2, r.y+r.h-2, "▼", colDim, colBg, false, r.x+r.w)
	}
}

// deadlineColor: red when overdue, amber within two days, dim otherwise.
// Finished tasks are never flagged.
func deadlineColor(t *domain.Task, now time.Time) lipgloss.Color {
	if t.Zone == domain.ZoneDone {
		return colDim
	}
	switch d, ok := domain.DaysUntil(t.Deadline, now); {
	case ok && d < 0:
		return colRed
	case ok && d <= 2:
		return colAmb
	}
	return colDim
}

func (m *Model) drawStatus(cv *canvas) {
	y := m.h - 1
	switch {
	case m.mode == modeConfirm:
		if t := m.cur(); t != nil {
			cv.put(1, y, "Delete “"+runewidth.Truncate(t.Title, 40, "…")+"”?  y = yes · any other key = cancel",
				colAmb, colBg, true, m.w)
		}
	case m.msg != "":
		cv.put(1, y, m.msg, colRed, colBg, true, m.w)
	case m.drag.active && m.drag.over != noZone:
		if t := m.board.Get(m.drag.pressID); t != nil {
			cv.put(1, y, "drop “"+runewidth.Truncate(t.Title, 30, "…")+"” into  "+zoneStyles[m.drag.over].name,
				colBone, colBg, false, m.w)
		}
	default:
		cv.put(1, y, "←↑↓→ move · a add · e edit · d delete · enter advance · 1-4 quadrant · 5 doing · 6 done · [ ] priority · drag with mouse · q quit",
			colDim, colBg, false, m.w)
	}
}

// drawDragGhost draws a tag with the task title next to the pointer.
func (m *Model) drawDragGhost(cv *canvas) {
	if !m.drag.active {
		return
	}
	if t := m.board.Get(m.drag.pressID); t != nil {
		ghost := " " + runewidth.Truncate(t.Title, 26, "…") + " "
		cv.put(m.drag.x+2, m.drag.y, ghost, colBg, colBone, true, m.w)
	}
}
