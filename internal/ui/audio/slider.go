package audio

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
)

// renderVolumeBar draws a static block-bar slider. bubbles has no dedicated
// slider widget, only an animated progress bar meant for async progress, not
// an immediately-adjustable value — a static bar reflects every keypress
// instantly without depending on progress.Model's frame-tick animation.
func renderVolumeBar(pct, barWidth int, muted bool) string {
	if barWidth < 4 {
		barWidth = 4
	}
	filled := pct * barWidth / 100
	switch {
	case filled > barWidth:
		filled = barWidth
	case filled < 0:
		filled = 0
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	pctText := fmt.Sprintf("%3d%%", pct)

	barColor := ilovetui.S.Primary
	if muted {
		barColor = ilovetui.S.Subtle
	}
	line := lipgloss.NewStyle().Foreground(barColor).Render("["+bar+"]") + " " + subtleStyle().Render(pctText)
	if muted {
		line += " " + errorStyle().Bold(true).Render("MUTED")
	}
	return line
}

// renderLevelMeter draws a live input-level meter from a peak value in
// [0, 1], used to test a microphone.
func renderLevelMeter(peak float64, barWidth int) string {
	if barWidth < 4 {
		barWidth = 4
	}
	filled := int(peak*float64(barWidth) + 0.5)
	switch {
	case filled > barWidth:
		filled = barWidth
	case filled < 0:
		filled = 0
	}
	color := ilovetui.S.Success
	if peak > 0.9 {
		color = ilovetui.S.Error
	} else if peak > 0.65 {
		color = ilovetui.S.Warning
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	return "Level: " + lipgloss.NewStyle().Foreground(color).Render("["+bar+"]")
}
