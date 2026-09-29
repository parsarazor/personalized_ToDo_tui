package ui

import "kanban/internal/domain"

const noZone domain.Zone = -1

type rect struct{ x, y, w, h int }

func (r rect) has(x, y int) bool { return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h }

type layout struct {
	backlog rect
	zone    [domain.NumZones]rect
}

// computeLayout: row 0 is the header, the last row the status line,
// everything in between is the board (backlog | in progress | done).
// The backlog box holds the 2×2 priority matrix.
func computeLayout(w, h int) layout {
	by, bh := 1, h-2
	bw := w / 2
	dw := (w - bw) / 2
	nw := w - bw - dw

	l := layout{backlog: rect{0, by, bw, bh}}
	iw, ih := bw-2, bh-2
	lw, th := iw/2, ih/2
	x0, y0 := 1, by+1
	l.zone[domain.ZoneDo] = rect{x0, y0, lw, th}
	l.zone[domain.ZonePlan] = rect{x0 + lw, y0, iw - lw, th}
	l.zone[domain.ZoneDelegate] = rect{x0, y0 + th, lw, ih - th}
	l.zone[domain.ZoneDrop] = rect{x0 + lw, y0 + th, iw - lw, ih - th}
	l.zone[domain.ZoneDoing] = rect{bw, by, dw, bh}
	l.zone[domain.ZoneDone] = rect{bw + dw, by, nw, bh}
	return l
}

func (l layout) zoneAt(x, y int) domain.Zone {
	for z := domain.Zone(0); z < domain.NumZones; z++ {
		if l.zone[z].has(x, y) {
			return z
		}
	}
	return noZone
}
