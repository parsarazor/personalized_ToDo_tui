package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DateLayout is the storage format of deadlines.
const DateLayout = "2006-01-02"

var ErrBadDeadline = errors.New("deadline: use YYYY-MM-DD, MM-DD, +N, today or tomorrow")

func dayOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// ParseDeadline turns user input into a normalised YYYY-MM-DD string.
// Accepted: "", YYYY-MM-DD, MM-DD (next occurrence), +N (days), today, tomorrow.
func ParseDeadline(s string, now time.Time) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	today := dayOf(now)
	switch {
	case s == "":
		return "", nil
	case s == "today" || s == "t":
		return today.Format(DateLayout), nil
	case s == "tomorrow":
		return today.AddDate(0, 0, 1).Format(DateLayout), nil
	case strings.HasPrefix(s, "+"):
		if n, err := strconv.Atoi(s[1:]); err == nil {
			return today.AddDate(0, 0, n).Format(DateLayout), nil
		}
	}
	if d, err := time.Parse(DateLayout, s); err == nil {
		return d.Format(DateLayout), nil
	}
	if d, err := time.Parse(DateLayout, fmt.Sprintf("%d-%s", today.Year(), s)); err == nil {
		if d.Before(today) {
			d = d.AddDate(1, 0, 0)
		}
		return d.Format(DateLayout), nil
	}
	return "", ErrBadDeadline
}

// DaysUntil returns whole days from now to the deadline (negative = overdue).
func DaysUntil(deadline string, now time.Time) (int, bool) {
	d, err := time.Parse(DateLayout, deadline)
	if err != nil {
		return 0, false
	}
	return int(d.Sub(dayOf(now)).Hours() / 24), true
}

// FormatShort renders a deadline for display, e.g. "05 Oct" or "05 Oct 27".
func FormatShort(deadline string, now time.Time) string {
	d, err := time.Parse(DateLayout, deadline)
	if err != nil {
		return deadline
	}
	if d.Year() == now.Year() {
		return d.Format("02 Jan")
	}
	return d.Format("02 Jan 06")
}
