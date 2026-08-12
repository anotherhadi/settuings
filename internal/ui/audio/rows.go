package audio

import (
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/audio"
	"github.com/anotherhadi/settuings/internal/icons"
)

// deviceRowHeight and streamRowHeight are the number of lines each list row
// occupies (name line + bar line + one blank spacer line), used to size how
// many rows fit in the available height.
const (
	deviceRowHeight = 3
	streamRowHeight = 3
)

// barWidth derives a volume-bar width from the row's available width,
// leaving room for the name line's cursor/default marker on one side and the
// percentage/MUTED label on the other.
func barWidth(rowWidth int) int {
	w := rowWidth - 24
	switch {
	case w > 30:
		w = 30
	case w < 10:
		w = 10
	}
	return w
}

func renderDeviceRow(d audio.Device, selected bool, width int) string {
	textStyle := lipgloss.NewStyle().Foreground(ilovetui.S.Text)
	if selected {
		textStyle = textStyle.Bold(true).Foreground(ilovetui.S.Primary)
	}
	cursor := "  "
	if selected {
		cursor = "> "
	}
	name := textStyle.Render(d.Name)
	if d.Default {
		name += " " + successStyle().Render(icons.I.Check+"default")
	}
	nameLine := cursor + name
	barLine := "    " + renderVolumeBar(d.Volume, barWidth(width), d.Muted)
	return nameLine + "\n" + barLine
}

func renderStreamRow(s audio.Stream, selected bool, width int) string {
	textStyle := lipgloss.NewStyle().Foreground(ilovetui.S.Text)
	if selected {
		textStyle = textStyle.Bold(true).Foreground(ilovetui.S.Primary)
	}
	cursor := "  "
	if selected {
		cursor = "> "
	}
	nameLine := cursor + textStyle.Render(s.App)
	barLine := "    " + renderVolumeBar(s.Volume, barWidth(width), s.Muted)
	return nameLine + "\n" + barLine
}
