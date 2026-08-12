package bluetooth

import (
	"errors"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/bluetooth"
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
	tabBar := renderTabBar(innerW)

	var body string
	switch {
	case !m.checked:
		body = ilovetui.S.Faint.Render("Checking bluetoothctl…")
	case m.checkErr != nil:
		body = m.renderUnavailable()
	case m.detail.open:
		body = m.renderDetail()
	default:
		body = m.renderMain(innerW, innerH)
	}

	box := lipgloss.NewStyle().Width(innerW).Height(innerH).Render(body)
	content := windowStyle().Render(lipgloss.JoinVertical(lipgloss.Left, tabBar, box))

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, content, statusBar))
}

func (m Model) renderStatusBar() string {
	km := bluetoothKeyMap{m: m, width: m.width}
	return lipgloss.NewStyle().Padding(0, 1).Render(m.help.View(km))
}

func (m Model) renderMain(width, height int) string {
	header := m.renderPowerStatus()
	headerH := lipgloss.Height(header) + 1 // blank line after
	listH := height - headerH
	if listH < 1 {
		listH = 1
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, "", m.renderList(width, listH))
}

func (m Model) renderPowerStatus() string {
	label := "Bluetooth: "
	status := errorStyle().Bold(true).Render("Off")
	if m.controller.Powered {
		status = successStyle().Bold(true).Render("On")
	}
	line := label + status
	switch {
	case m.list.scanning:
		line += mutedStyle().Render("  scanning…")
	case m.list.pending:
		line += mutedStyle().Render("  working…")
	}
	return line
}

func (m Model) renderUnavailable() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(ilovetui.S.Error).Render("Bluetooth page unavailable")

	var hint string
	switch {
	case errors.Is(m.checkErr, bluetooth.ErrNotFound):
		hint = "bluetoothctl was not found in PATH. Install BlueZ to use this page."
	case errors.Is(m.checkErr, bluetooth.ErrNoController):
		hint = "No Bluetooth controller was found. Try: sudo systemctl start bluetooth"
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
