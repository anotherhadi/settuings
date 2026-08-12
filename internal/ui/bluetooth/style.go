package bluetooth

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
