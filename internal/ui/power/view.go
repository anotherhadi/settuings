package power

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
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

	box := lipgloss.NewStyle().Width(innerW).Height(innerH).Render(m.renderBody(innerW))
	content := windowStyle().Render(lipgloss.JoinVertical(lipgloss.Left, tabBar, box))

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, content, statusBar))
}

func (m Model) renderStatusBar() string {
	km := powerKeyMap{m: m, width: m.width}
	return lipgloss.NewStyle().Padding(0, 1).Render(m.help.View(km))
}

// contentBudget returns the exact space available inside the bordered page
// box, below the tab bar and above the status bar.
func (m Model) contentBudget() (w, h int) {
	frameW := windowStyle().GetHorizontalFrameSize()
	frameH := windowStyle().GetVerticalFrameSize()
	w = m.width - frameW

	statusH := strings.Count(m.renderStatusBar(), "\n") + 1
	tabBarH := lipgloss.Height(renderTabBar(w))

	h = m.height - tabBarH - frameH - statusH
	if h < 1 {
		h = 1
	}
	return w, h
}

// renderBody lays out the two sections side by side once there's enough
// width for both to breathe, and stacks them otherwise.
func (m Model) renderBody(width int) string {
	const splitThreshold = 60
	const gap = 3

	if width < splitThreshold {
		return lipgloss.JoinVertical(lipgloss.Left, m.renderProfiles(), "", m.renderBattery(width))
	}

	leftW := width / 3
	if leftW < 20 {
		leftW = 20
	}
	rightW := width - leftW - gap
	left := lipgloss.NewStyle().Width(leftW).Render(m.renderProfiles())
	right := lipgloss.NewStyle().Width(rightW).Render(m.renderBattery(rightW))
	return lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", gap), right)
}
