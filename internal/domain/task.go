// Package domain holds the pure business model of the board.
// It knows nothing about terminals, JSON files or Bubble Tea.
package domain

// Zone is a place a task can live: one of the four Eisenhower-matrix
// quadrants of the backlog, or one of the two workflow columns.
type Zone int

const (
	ZoneDo       Zone = iota // urgent + important
	ZonePlan                 // important, not urgent
	ZoneDelegate             // urgent, not important
	ZoneDrop                 // neither
	ZoneDoing
	ZoneDone
	NumZones
)

// IsBacklog reports whether the zone is one of the four matrix quadrants.
func (z Zone) IsBacklog() bool { return z >= ZoneDo && z <= ZoneDrop }

// Valid reports whether z is a real zone.
func (z Zone) Valid() bool { return z >= 0 && z < NumZones }

// Advance returns the next workflow stage: backlog → doing → done.
func (z Zone) Advance() Zone {
	switch {
	case z.IsBacklog():
		return ZoneDoing
	case z == ZoneDoing:
		return ZoneDone
	}
	return z
}

// Task is a single unit of work.
type Task struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Deadline string `json:"deadline,omitempty"` // YYYY-MM-DD, empty = none
	Priority int    `json:"priority"`           // 1 (top) … 9, 0 = unset
	Zone     Zone   `json:"zone"`
}

// Repository persists a whole board.
type Repository interface {
	Load() (*Board, error)
	Save(*Board) error
}
