package ui

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"

	"kanban/internal/domain"
)

const (
	fieldTitle = iota
	fieldDeadline
	fieldPriority
	fieldZone
	numFields
)

type formState struct {
	editID int // 0 = creating a new task
	text   [3][]rune
	zone   domain.Zone
	field  int
	err    string
}

func (m *Model) openForm(t *domain.Task) {
	f := formState{zone: m.zone}
	if t != nil {
		f.editID = t.ID
		f.text[fieldTitle] = []rune(t.Title)
		f.text[fieldDeadline] = []rune(t.Deadline)
		if t.Priority > 0 {
			f.text[fieldPriority] = []rune(strconv.Itoa(t.Priority))
		}
		f.zone = t.Zone
	}
	m.form, m.mode = f, modeForm
}

func (m *Model) onFormKey(k tea.KeyMsg) tea.Cmd {
	f := &m.form
	cycleZone := func(d int) { f.zone = (f.zone + domain.NumZones + domain.Zone(d)) % domain.NumZones }
	f.err = ""

	switch k.String() {
	case "ctrl+c":
		return tea.Quit
	case "esc":
		m.mode = modeNormal
	case "tab", "down":
		f.field = (f.field + 1) % numFields
	case "shift+tab", "up":
		f.field = (f.field + numFields - 1) % numFields
	case "enter":
		if f.field < fieldZone {
			f.field++
		} else {
			m.submit()
		}
	case "ctrl+s":
		m.submit()
	case "backspace":
		if f.field < fieldZone && len(f.text[f.field]) > 0 {
			f.text[f.field] = f.text[f.field][:len(f.text[f.field])-1]
		}
	case "ctrl+u":
		if f.field < fieldZone {
			f.text[f.field] = nil
		}
	case "left":
		if f.field == fieldZone {
			cycleZone(-1)
		}
	case "right":
		if f.field == fieldZone {
			cycleZone(1)
		}
	default:
		if k.Type != tea.KeyRunes && k.Type != tea.KeySpace {
			return nil
		}
		s := string(k.Runes)
		if k.Type == tea.KeySpace {
			s = " "
		}
		s = strings.NewReplacer("\n", " ", "\r", "").Replace(s)
		switch f.field {
		case fieldZone:
			if s == " " {
				cycleZone(1)
			}
		case fieldPriority:
			for _, r := range normDigits(s) {
				if r >= '0' && r <= '9' {
					f.text[fieldPriority] = []rune{r}
				}
			}
		default:
			f.text[f.field] = append(f.text[f.field], []rune(s)...)
		}
	}
	return nil
}

func (m *Model) submit() {
	f := &m.form
	deadline, err := domain.ParseDeadline(string(f.text[fieldDeadline]), m.now())
	if err != nil {
		f.err, f.field = err.Error(), fieldDeadline
		return
	}
	prio := 0
	if p := f.text[fieldPriority]; len(p) > 0 {
		prio = int(p[0] - '0')
	}
	in := domain.Input{Title: string(f.text[fieldTitle]), Deadline: deadline, Priority: prio, Zone: f.zone}

	var t *domain.Task
	if f.editID != 0 {
		t, err = m.board.Update(f.editID, in)
	} else {
		t, err = m.board.Add(in)
	}
	if err != nil {
		f.err = err.Error()
		if err == domain.ErrEmptyTitle {
			f.field = fieldTitle
		}
		return
	}
	m.persist()
	m.mode = modeNormal
	m.focus(t.ID)
}

// tailFit keeps the end of s so the caret stays visible in narrow fields.
func tailFit(s string, w int) string {
	rs := []rune(s)
	acc, i := 0, len(rs)
	for i > 0 && acc+runewidth.RuneWidth(rs[i-1]) <= w {
		i--
		acc += runewidth.RuneWidth(rs[i])
	}
	return string(rs[i:])
}

func (m *Model) drawForm(cv *canvas) {
	f := &m.form
	mw, mh := min(62, m.w-4), 13
	r := rect{(m.w - mw) / 2, (m.h - mh) / 2, mw, mh}
	cv.fill(r, colBg)

	title := "NEW TASK"
	if f.editID != 0 {
		title = "EDIT TASK"
	}
	cv.box(r, borderDouble, colBone, title, "", -1)

	labels := [numFields]string{"Title", "Deadline", "Priority", "Box"}
	hints := [numFields]string{
		"what needs doing",
		"YYYY-MM-DD · MM-DD · +3 (days from now) · today · tomorrow · empty = none",
		"1 = most important … 9 · empty = none",
		"← → or space to change the box",
	}
	right := r.x + r.w - 1
	vx := r.x + 13
	vw := right - 1 - vx
	for i := 0; i < numFields; i++ {
		y := r.y + 2 + i*2
		on := f.field == i
		lfg := colDim
		if on {
			lfg = colBone
		}
		cv.put(r.x+2, y, labels[i], lfg, colBg, on, vx)
		bg := colBg
		if on {
			bg = colSel
			cv.paintBg(y, vx-1, right, bg)
		}
		if i < fieldZone {
			s := tailFit(string(f.text[i]), vw-1)
			cv.put(vx, y, s, colBone, bg, false, right)
			if on {
				cv.put(vx+runewidth.StringWidth(s), y, "▏", colBone, bg, true, right)
			}
		} else {
			zs := zoneStyles[f.zone]
			cv.put(vx, y, "‹ "+zs.name+" ›", zs.accent, bg, true, right)
		}
	}
	cv.put(r.x+2, r.y+10, runewidth.Truncate(hints[f.field], mw-4, "…"), colDim, colBg, false, right)
	if f.err != "" {
		cv.put(r.x+2, r.y+11, f.err, colRed, colBg, true, right)
	} else {
		cv.put(r.x+2, r.y+11, "tab/↑↓ field · enter next/save · ctrl+s save · esc cancel", colDim, colBg, false, right)
	}
}
