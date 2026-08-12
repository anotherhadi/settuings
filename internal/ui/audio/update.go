package audio

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
			// recovered: resume fetching + the auto-refresh loop
			return m, tea.Batch(fetchOutputCmd(), fetchInputCmd(), fetchAppsCmd(), tickCmd())
		}
		return m, nil

	case tickMsg:
		if m.checkErr != nil {
			return m, nil
		}
		return m, tea.Batch(fetchOutputCmd(), fetchInputCmd(), fetchAppsCmd(), tickCmd())

	case outputListMsg:
		m.handleOutputList(msg)
		return m, nil
	case outputActionMsg:
		return m, m.handleOutputAction(msg)

	case inputListMsg:
		m.handleInputList(msg)
		return m, nil
	case inputActionMsg:
		return m, m.handleInputAction(msg)
	case levelStartedMsg:
		return m, m.handleLevelStarted(msg)
	case levelMsg:
		return m, m.handleLevel(msg)

	case appsListMsg:
		m.handleAppsList(msg)
		return m, nil
	case appsActionMsg:
		return m, m.handleAppsAction(msg)

	case tea.KeyPressMsg:
		if m.checkErr != nil {
			if key.Matches(msg, keys.Keys.Audio.Refresh) {
				return m, checkCmd()
			}
			return m, nil
		}

		g := keys.Keys.Global
		switch {
		case key.Matches(msg, g.CycleFocus):
			m.cycleTab()
			return m, nil
		case key.Matches(msg, g.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}

		var cmd tea.Cmd
		switch m.active {
		case tabOutput:
			cmd = m.updateOutput(msg)
		case tabInput:
			cmd = m.updateInput(msg)
		case tabApps:
			cmd = m.updateApps(msg)
		}
		return m, cmd
	}
	return m, nil
}
