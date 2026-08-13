package input

import (
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
)

func errorStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ilovetui.S.Error)
}

func successStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ilovetui.S.Success)
}

func subtleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ilovetui.S.Subtle)
}

func mutedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ilovetui.S.Muted)
}

// sectionTitleStyle highlights a section's title when it holds keyboard
// focus, so Tab-cycling between the layout list and the mouse slider is
// visible even before the user starts navigating within a section.
func sectionTitleStyle(focused bool) lipgloss.Style {
	style := lipgloss.NewStyle().Bold(true)
	if focused {
		style = style.Foreground(ilovetui.S.Primary)
	}
	return style
}
