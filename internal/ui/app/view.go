package app

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
)

func (m Model) newView(content string) tea.View {
	v := tea.NewView(content)
	v.AltScreen = true
	v.WindowTitle = "Settuings"
	return v
}

func (m Model) View() tea.View {
	if m.width == 0 {
		return m.newView("")
	}

	rendered := m.renderNormal()
	if m.notifications.HasNotifications() {
		rendered = m.notifications.View(rendered)
	}

	v := m.newView(rendered)
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (m Model) renderNormal() string {
	sidebar := m.renderSidebar()
	content := m.renderActivePage()
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)
}

func (m *Model) renderActivePage() string {
	for _, e := range pageRegistry {
		if e.id == m.page && e.render != nil {
			return e.render(m)
		}
	}
	return ilovetui.S.Faint.Render("Work in progress")
}
