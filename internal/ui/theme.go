package ui

import (
	"github.com/charmbracelet/lipgloss"

	"kanban/internal/domain"
)

// Palette: black canvas, bone-white text, coloured outlines per quadrant.
const (
	colBg   = lipgloss.Color("#000000")
	colBone = lipgloss.Color("#E8E4D8")
	colDim  = lipgloss.Color("#77736A")
	colSel  = lipgloss.Color("#22201C")
	colEdge = lipgloss.Color("#4A473F")
	colRed  = lipgloss.Color("#E5645A")
	colAmb  = lipgloss.Color("#E0B050")
)

type zoneStyle struct {
	name, sub string
	accent    lipgloss.Color // priority digit, focused border
	border    lipgloss.Color // border when not focused
}

var zoneStyles = [domain.NumZones]zoneStyle{
	domain.ZoneDo:       {"1 DO NOW", "urgent · important", "#E5645A", "#E5645A"},
	domain.ZonePlan:     {"2 SCHEDULE", "important · not urgent", "#6C9BD1", "#6C9BD1"},
	domain.ZoneDelegate: {"3 DELEGATE", "urgent · not important", "#E0B050", "#E0B050"},
	domain.ZoneDrop:     {"4 ELIMINATE", "neither", "#8F8FA3", "#8F8FA3"},
	domain.ZoneDoing:    {"IN PROGRESS", "", colBone, colDim},
	domain.ZoneDone:     {"DONE", "", "#7FA88A", "#4F6656"},
}
