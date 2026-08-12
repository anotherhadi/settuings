package bluetooth

import (
	"strings"

	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/icons"
)

// renderTabBar draws a header matching the network page's tab bar, but with
// a single always-active entry: this page has no tabs of its own, just a
// visual label for consistency.
func renderTabBar(width int) string {
	style := lipgloss.NewStyle().Bold(true).Foreground(ilovetui.S.Primary).Padding(0, 2).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(ilovetui.S.Primary)

	label := style.Render(icons.I.Bluetooth + "Bluetooth")

	filler := width - lipgloss.Width(label)
	if filler > 0 {
		label = lipgloss.JoinHorizontal(lipgloss.Bottom, label,
			lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(ilovetui.S.Primary).
				Render(strings.Repeat(" ", filler)))
	}
	return label
}
