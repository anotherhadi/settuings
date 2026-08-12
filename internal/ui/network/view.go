package network

import (
	"errors"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/network"
)

func windowStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ilovetui.S.Subtle).
		Padding(1, 2)
}

func (m Model) View() tea.View {
	statusBar := m.renderStatusBar()
	innerW, innerH := m.contentBudget()
	tabBar := m.renderTabBar(innerW)

	var body string
	switch {
	case !m.checked:
		body = ilovetui.S.Faint.Render("Checking nmcli…")
	case m.checkErr != nil:
		body = m.renderUnavailable()
	case m.deviceDetail.open:
		body = m.renderDeviceDetail()
	default:
		body = m.renderActiveTab(innerW, innerH)
	}

	box := lipgloss.NewStyle().Width(innerW).Height(innerH).Render(body)
	content := windowStyle().Render(lipgloss.JoinVertical(lipgloss.Left, tabBar, box))

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, content, statusBar))
}

func (m Model) renderStatusBar() string {
	km := networkKeyMap{m: m, width: m.width}
	return lipgloss.NewStyle().Padding(0, 1).Render(m.help.View(km))
}

func (m Model) renderActiveTab(width, height int) string {
	switch m.active {
	case tabWifi:
		return m.renderWifi(width, height)
	case tabEthernet:
		return m.renderEthernet(height)
	case tabVPN:
		return m.renderVPN(height)
	case tabKnown:
		return m.renderKnown(height)
	default:
		return ""
	}
}

func (m Model) renderUnavailable() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(ilovetui.S.Error).Render("Network page unavailable")

	var hint string
	switch {
	case errors.Is(m.checkErr, network.ErrNotFound):
		hint = "nmcli was not found in PATH. Install NetworkManager to use this page."
	case errors.Is(m.checkErr, network.ErrNotRunning):
		hint = "NetworkManager doesn't seem to be running. Try: sudo systemctl start NetworkManager"
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
