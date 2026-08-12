package network

import (
	"strings"

	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
)

// cycleTab moves to the next visible tab. From the hidden Known tab it
// returns to whichever tab was active before Known was opened.
func (m *Model) cycleTab() {
	if m.active == tabKnown {
		m.active = m.previousTab
		return
	}
	for i, t := range visibleTabs {
		if t == m.active {
			m.active = visibleTabs[(i+1)%len(visibleTabs)]
			return
		}
	}
}

// toggleKnown opens the hidden Known tab, or leaves it back to the
// previously active tab if it's already open.
func (m *Model) toggleKnown() {
	if m.active == tabKnown {
		m.active = m.previousTab
		return
	}
	m.previousTab = m.active
	m.active = tabKnown
}

func (m Model) renderTabBar(width int) string {
	active := lipgloss.NewStyle().Bold(true).Foreground(ilovetui.S.Primary).Padding(0, 2).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(ilovetui.S.Primary)
	inactive := lipgloss.NewStyle().Foreground(ilovetui.S.Subtle).Padding(0, 2).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(ilovetui.S.Subtle)

	var tabs []string
	for _, t := range visibleTabs {
		style := inactive
		if t == m.active {
			style = active
		}
		tabs = append(tabs, style.Render(t.icon()+t.label()))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)

	filler := width - lipgloss.Width(row)
	if filler > 0 {
		row = lipgloss.JoinHorizontal(lipgloss.Bottom, row,
			lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(ilovetui.S.Subtle).
				Render(strings.Repeat(" ", filler)))
	}
	return row
}
