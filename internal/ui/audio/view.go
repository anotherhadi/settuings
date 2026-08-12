package audio

import (
	"errors"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/audio"
)

func windowStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ilovetui.S.Subtle).
		Padding(1, 2)
}

func (m Model) View() tea.View {
	statusBar := m.renderStatusBar()
	statusH := strings.Count(statusBar, "\n") + 1

	frameW := windowStyle().GetHorizontalFrameSize()
	frameH := windowStyle().GetVerticalFrameSize()
	innerW := m.width - frameW

	tabBar := m.renderTabBar(innerW)
	tabBarH := lipgloss.Height(tabBar)

	innerH := m.height - tabBarH - frameH - statusH

	var body string
	switch {
	case !m.checked:
		body = ilovetui.S.Faint.Render("Checking wpctl…")
	case m.checkErr != nil:
		body = m.renderUnavailable()
	default:
		body = m.renderActiveTab(innerW, innerH)
	}

	box := lipgloss.NewStyle().Width(innerW).Height(innerH).Render(body)
	content := windowStyle().Render(lipgloss.JoinVertical(lipgloss.Left, tabBar, box))

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, content, statusBar))
}

func (m Model) renderStatusBar() string {
	km := audioKeyMap{m: m, width: m.width}
	return lipgloss.NewStyle().Padding(0, 1).Render(m.help.View(km))
}

func (m Model) renderActiveTab(width, height int) string {
	switch m.active {
	case tabOutput:
		return m.renderOutput(width, height)
	case tabInput:
		return m.renderInput(width, height)
	case tabApps:
		return m.renderApps(width, height)
	default:
		return ""
	}
}

func (m Model) renderUnavailable() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(ilovetui.S.Error).Render("Audio page unavailable")

	var hint string
	switch {
	case errors.Is(m.checkErr, audio.ErrNotFound):
		hint = "wpctl was not found in PATH. Install PipeWire/WirePlumber to use this page."
	default:
		hint = m.checkErr.Error()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		subtleStyle().Render(hint),
		"",
		mutedStyle().Render("Press r to try again once fixed."),
	)
}
