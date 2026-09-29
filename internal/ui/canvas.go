package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// The UI is drawn on an explicit cell grid so every cell gets the black
// background and mouse coordinates map 1:1 onto what is on screen.

type cell struct {
	ch     string
	fg, bg lipgloss.Color
	bold   bool
}

type canvas struct {
	w, h  int
	cells [][]cell
}

func newCanvas(w, h int) *canvas {
	cv := &canvas{w: w, h: h, cells: make([][]cell, h)}
	for y := range cv.cells {
		cv.cells[y] = make([]cell, w)
		for x := range cv.cells[y] {
			cv.cells[y][x] = cell{" ", colBone, colBg, false}
		}
	}
	return cv
}

// put writes s at (x,y); nothing is drawn at or beyond column maxX.
func (cv *canvas) put(x, y int, s string, fg, bg lipgloss.Color, bold bool, maxX int) {
	if y < 0 || y >= cv.h {
		return
	}
	maxX = min(maxX, cv.w)
	prev := -1
	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if rw == 0 { // combining mark: attach to the previous cell
			if prev >= 0 {
				cv.cells[y][prev].ch += string(r)
			}
			continue
		}
		if x < 0 {
			x += rw
			continue
		}
		if x+rw > maxX {
			return
		}
		cv.cells[y][x] = cell{string(r), fg, bg, bold}
		prev = x
		if rw == 2 {
			cv.cells[y][x+1] = cell{"", fg, bg, bold}
		}
		x += rw
	}
}

func (cv *canvas) fill(r rect, bg lipgloss.Color) {
	for y := r.y; y < r.y+r.h && y < cv.h; y++ {
		for x := r.x; x < r.x+r.w && x < cv.w; x++ {
			cv.cells[y][x] = cell{" ", colBone, bg, false}
		}
	}
}

func (cv *canvas) paintBg(y, x0, x1 int, bg lipgloss.Color) {
	if y < 0 || y >= cv.h {
		return
	}
	for x := max(0, x0); x < x1 && x < cv.w; x++ {
		cv.cells[y][x].bg = bg
	}
}

type borderKind int

const (
	borderRounded borderKind = iota
	borderHeavy              // focused zone
	borderDouble             // drop target while dragging
)

var borderSets = [...][6]string{
	borderRounded: {"╭", "╮", "╰", "╯", "─", "│"},
	borderHeavy:   {"┏", "┓", "┗", "┛", "━", "┃"},
	borderDouble:  {"╔", "╗", "╚", "╝", "═", "║"},
}

// box draws a titled border. count < 0 hides the counter.
func (cv *canvas) box(r rect, kind borderKind, col lipgloss.Color, name, sub string, count int) {
	s := borderSets[kind]
	bold := kind != borderRounded
	right, bottom := r.x+r.w-1, r.y+r.h-1
	cv.put(r.x, r.y, s[0], col, colBg, bold, cv.w)
	cv.put(right, r.y, s[1], col, colBg, bold, cv.w)
	cv.put(r.x, bottom, s[2], col, colBg, bold, cv.w)
	cv.put(right, bottom, s[3], col, colBg, bold, cv.w)
	for x := r.x + 1; x < right; x++ {
		cv.put(x, r.y, s[4], col, colBg, bold, cv.w)
		cv.put(x, bottom, s[4], col, colBg, bold, cv.w)
	}
	for y := r.y + 1; y < bottom; y++ {
		cv.put(r.x, y, s[5], col, colBg, bold, cv.w)
		cv.put(right, y, s[5], col, colBg, bold, cv.w)
	}

	end := right
	if count >= 0 {
		cs := fmt.Sprintf(" %d ", count)
		cx := end - 1 - runewidth.StringWidth(cs)
		cv.put(cx, r.y, cs, colDim, colBg, false, end)
		end = cx
	}
	cv.put(r.x+2, r.y, " "+name+" ", col, colBg, true, end)
	if sub != "" {
		cv.put(r.x+2+runewidth.StringWidth(name)+2, r.y, sub+" ", colDim, colBg, false, end)
	}
}

type styleKey struct {
	fg, bg lipgloss.Color
	bold   bool
}

// String renders the grid, merging neighbouring cells with the same style.
func (cv *canvas) String() string {
	var sb strings.Builder
	cache := map[styleKey]lipgloss.Style{}
	for y, row := range cv.cells {
		if y > 0 {
			sb.WriteByte('\n')
		}
		for x := 0; x < len(row); {
			k := styleKey{row[x].fg, row[x].bg, row[x].bold}
			var run strings.Builder
			for x < len(row) && (styleKey{row[x].fg, row[x].bg, row[x].bold}) == k {
				run.WriteString(row[x].ch)
				x++
			}
			st, ok := cache[k]
			if !ok {
				st = lipgloss.NewStyle().Foreground(k.fg).Background(k.bg).Bold(k.bold)
				cache[k] = st
			}
			sb.WriteString(st.Render(run.String()))
		}
	}
	return sb.String()
}
