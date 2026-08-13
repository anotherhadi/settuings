package power

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/anotherhadi/settuings/internal/keys"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case checkMsg:
		wasUnavailable := m.checkErr != nil
		m.checked = true
		m.checkErr = msg.err
		if msg.err == nil && wasUnavailable {
			// recovered: resume fetching profiles
			return m, fetchProfilesCmd()
		}
		return m, nil

	case tickMsg:
		cmds := []tea.Cmd{tickCmd(), fetchBatteryCmd()}
		if m.checkErr == nil {
			cmds = append(cmds, fetchProfilesCmd())
		}
		return m, tea.Batch(cmds...)

	case profilesMsg:
		m.handleProfiles(msg)
		return m, nil
	case setProfileMsg:
		return m, m.handleSetProfile(msg)
	case batteryMsg:
		m.handleBattery(msg)
		return m, nil

	case tea.KeyPressMsg:
		if key.Matches(msg, keys.Keys.Global.Help) {
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}
		return m, m.updateProfiles(msg)
	}
	return m, nil
}
