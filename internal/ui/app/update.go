package app

import (
	"os/exec"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/anotherhadi/settuings/internal/keys"
	notificationsUI "github.com/anotherhadi/settuings/internal/ui/components/notifications"
	"github.com/anotherhadi/settuings/internal/util"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case notificationsUI.NotificationMsg:
		var cmd tea.Cmd
		m.notifications, cmd = m.notifications.Update(msg)
		return m, cmd
	case notificationsUI.DismissMsg:
		var cmd tea.Cmd
		m.notifications, cmd = m.notifications.Update(msg)
		return m, cmd
	}

	var activateCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeChildren()

	case tea.KeyPressMsg:
		if key.Matches(msg, keys.Keys.Global.Quit) && !m.activeIsEditing() {
			m.deactivatePage(m.page)
			return m, tea.Quit
		}

		if key.Matches(msg, keys.Keys.Global.OpenLogs) {
			return m, tea.ExecProcess(exec.Command(util.ResolveEditor(), m.logPath), nil) // #nosec G204 -- editor from trusted config/$EDITOR, logPath is a fixed path
		}

		if !m.activeIsEditing() {
			switch {
			case key.Matches(msg, keys.Keys.Global.ToggleSidebar):
				m.cycleSidebarState()
				m.resizeChildren()

			default:
				if p, ok := m.pageShortcuts[msg.String()]; ok {
					prev := m.page
					m.page = p
					if prev != p {
						m.deactivatePage(prev)
						activateCmd = m.activatePage(p)
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m, cmd = m.updateActivePage(msg)
	return m, tea.Batch(activateCmd, cmd)
}

func (m *Model) activatePage(p page) tea.Cmd {
	for _, e := range pageRegistry {
		if e.id == p && e.activate != nil {
			return e.activate(m)
		}
	}
	return nil
}

// deactivatePage synchronously tears down page p's background state, if it
// has any (see pageEntry.deactivate).
func (m *Model) deactivatePage(p page) {
	for _, e := range pageRegistry {
		if e.id == p && e.deactivate != nil {
			e.deactivate(m)
			return
		}
	}
}

func (m Model) activeIsEditing() bool {
	for _, e := range pageRegistry {
		if e.id == m.page && e.isEditing != nil {
			return e.isEditing(&m)
		}
	}
	return false
}

func (m Model) updateActivePage(msg tea.Msg) (Model, tea.Cmd) {
	for _, e := range pageRegistry {
		if e.id == m.page && e.update != nil {
			cmd := e.update(&m, msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m *Model) resizeChildren() {
	sidebarW := m.getSidebarWidth()
	h := m.height
	for _, e := range pageRegistry {
		if e.resize == nil {
			continue
		}
		e.resize(m, m.width-sidebarW, h)
	}
	m.notifications.SetSize(m.width, m.height)
}
