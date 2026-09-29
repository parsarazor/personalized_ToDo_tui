package domain

import (
	"errors"
	"testing"
	"time"
)

var now = time.Date(2026, 9, 29, 15, 0, 0, 0, time.UTC)

func TestParseDeadline(t *testing.T) {
	cases := map[string]string{
		"":           "",
		"2026-10-05": "2026-10-05",
		"10-05":      "2026-10-05",
		"01-05":      "2027-01-05", // already passed this year → next year
		"+3":         "2026-10-02",
		"today":      "2026-09-29",
		"tomorrow":   "2026-09-30",
	}
	for in, want := range cases {
		got, err := ParseDeadline(in, now)
		if err != nil || got != want {
			t.Errorf("ParseDeadline(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"soon", "13-45", "+x"} {
		if _, err := ParseDeadline(bad, now); !errors.Is(err, ErrBadDeadline) {
			t.Errorf("ParseDeadline(%q) err = %v; want ErrBadDeadline", bad, err)
		}
	}
}

func TestDaysUntil(t *testing.T) {
	if d, ok := DaysUntil("2026-09-27", now); !ok || d != -2 {
		t.Fatalf("got %d,%v", d, ok)
	}
	if _, ok := DaysUntil("", now); ok {
		t.Fatal("empty deadline must not parse")
	}
}

func TestBoardValidation(t *testing.T) {
	b := NewBoard(1, nil)
	bad := []Input{
		{Title: "  "},
		{Title: "x", Priority: 10},
		{Title: "x", Zone: NumZones},
		{Title: "x", Deadline: "nope"},
	}
	for _, in := range bad {
		if _, err := b.Add(in); err == nil {
			t.Errorf("Add(%+v) should fail", in)
		}
	}
	if _, err := b.Update(42, Input{Title: "x"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("Update unknown id: %v", err)
	}
}

func TestOrderingWithinZone(t *testing.T) {
	b := NewBoard(1, nil)
	add := func(title, dl string, p int) {
		if _, err := b.Add(Input{Title: title, Deadline: dl, Priority: p, Zone: ZoneDo}); err != nil {
			t.Fatal(err)
		}
	}
	add("none", "", 0)
	add("p2-late", "2026-12-01", 2)
	add("p2-soon", "2026-10-01", 2)
	add("p1", "", 1)
	var got []string
	for _, tk := range b.InZone(ZoneDo) {
		got = append(got, tk.Title)
	}
	want := []string{"p1", "p2-soon", "p2-late", "none"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v; want %v", got, want)
		}
	}
}

func TestMoveDeleteReprioritize(t *testing.T) {
	b := NewBoard(1, nil)
	tk, _ := b.Add(Input{Title: "a"})
	if err := b.Move(tk.ID, ZoneDone); err != nil || tk.Zone != ZoneDone {
		t.Fatal("move failed")
	}
	_ = b.Reprioritize(tk.ID, -1) // unset → 9
	_ = b.Reprioritize(tk.ID, -1) // → 8
	_ = b.Reprioritize(tk.ID, 1)  // → 9
	if tk.Priority != 9 {
		t.Fatalf("priority = %d", tk.Priority)
	}
	if err := b.Delete(tk.ID); err != nil || b.Get(tk.ID) != nil {
		t.Fatal("delete failed")
	}
	if err := b.Delete(tk.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal("second delete should fail")
	}
}

func TestNewBoardNeverReusesIDs(t *testing.T) {
	b := NewBoard(1, []*Task{{ID: 7, Title: "old"}})
	tk, _ := b.Add(Input{Title: "new"})
	if tk.ID != 8 {
		t.Fatalf("id = %d; want 8", tk.ID)
	}
}

func TestAdvance(t *testing.T) {
	if ZoneDo.Advance() != ZoneDoing || ZoneDoing.Advance() != ZoneDone || ZoneDone.Advance() != ZoneDone {
		t.Fatal("advance chain broken")
	}
}
