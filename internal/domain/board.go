package domain

import (
	"errors"
	"sort"
	"strings"
	"time"
)

var (
	ErrEmptyTitle  = errors.New("title can't be empty")
	ErrBadPriority = errors.New("priority must be between 0 and 9")
	ErrBadZone     = errors.New("unknown zone")
	ErrNotFound    = errors.New("task not found")
)

// Input is the editable part of a task.
type Input struct {
	Title    string
	Deadline string // already normalised (see ParseDeadline)
	Priority int
	Zone     Zone
}

func (in *Input) validate() error {
	in.Title = strings.TrimSpace(in.Title)
	switch {
	case in.Title == "":
		return ErrEmptyTitle
	case in.Priority < 0 || in.Priority > 9:
		return ErrBadPriority
	case !in.Zone.Valid():
		return ErrBadZone
	}
	if in.Deadline != "" {
		if _, err := time.Parse(DateLayout, in.Deadline); err != nil {
			return ErrBadDeadline
		}
	}
	return nil
}

// Board is the aggregate that owns every task.
type Board struct {
	next  int
	tasks []*Task
}

// NewBoard rebuilds a board from persisted state.
func NewBoard(next int, tasks []*Task) *Board {
	b := &Board{next: max(next, 1), tasks: tasks}
	for _, t := range tasks {
		b.next = max(b.next, t.ID+1)
	}
	return b
}

// Export returns the state needed to persist the board.
func (b *Board) Export() (next int, tasks []*Task) { return b.next, b.tasks }

func (b *Board) Get(id int) *Task {
	for _, t := range b.tasks {
		if t.ID == id {
			return t
		}
	}
	return nil
}

func (b *Board) Add(in Input) (*Task, error) {
	if err := in.validate(); err != nil {
		return nil, err
	}
	t := &Task{ID: b.next, Title: in.Title, Deadline: in.Deadline, Priority: in.Priority, Zone: in.Zone}
	b.next++
	b.tasks = append(b.tasks, t)
	return t, nil
}

func (b *Board) Update(id int, in Input) (*Task, error) {
	t := b.Get(id)
	if t == nil {
		return nil, ErrNotFound
	}
	if err := in.validate(); err != nil {
		return nil, err
	}
	t.Title, t.Deadline, t.Priority, t.Zone = in.Title, in.Deadline, in.Priority, in.Zone
	return t, nil
}

func (b *Board) Delete(id int) error {
	for i, t := range b.tasks {
		if t.ID == id {
			b.tasks = append(b.tasks[:i], b.tasks[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (b *Board) Move(id int, z Zone) error {
	t := b.Get(id)
	switch {
	case t == nil:
		return ErrNotFound
	case !z.Valid():
		return ErrBadZone
	}
	t.Zone = z
	return nil
}

// Reprioritize nudges a task's priority. delta < 0 makes it more important.
// An unset priority becomes 9 when raised and stays unset when lowered.
func (b *Board) Reprioritize(id, delta int) error {
	t := b.Get(id)
	if t == nil {
		return ErrNotFound
	}
	switch {
	case delta < 0 && t.Priority == 0:
		t.Priority = 9
	case delta < 0:
		t.Priority = max(1, t.Priority-1)
	case t.Priority > 0:
		t.Priority = min(9, t.Priority+1)
	}
	return nil
}

func prioKey(p int) int {
	if p == 0 {
		return 99
	}
	return p
}

// InZone lists a zone's tasks: by priority, then nearest deadline, then age.
func (b *Board) InZone(z Zone) []*Task {
	var out []*Task
	for _, t := range b.tasks {
		if t.Zone == z {
			out = append(out, t)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, c := out[i], out[j]
		if pa, pc := prioKey(a.Priority), prioKey(c.Priority); pa != pc {
			return pa < pc
		}
		if a.Deadline != c.Deadline {
			if a.Deadline == "" {
				return false
			}
			if c.Deadline == "" {
				return true
			}
			return a.Deadline < c.Deadline
		}
		return a.ID < c.ID
	})
	return out
}
