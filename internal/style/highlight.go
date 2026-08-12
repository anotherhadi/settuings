package style

import (
	"charm.land/lipgloss/v2"
	"image/color"
)

func Paint(c color.Color, s string) string {
	return lipgloss.NewStyle().Foreground(c).Render(s)
}
