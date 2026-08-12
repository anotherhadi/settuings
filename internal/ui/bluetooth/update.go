package bluetooth

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
			return m, tea.Batch(fetchControllerCmd(), fetchDevicesCmd(), tickCmd())
		}
		return m, nil

	case tickMsg:
		if m.checkErr != nil {
			return m, nil
		}
		return m, tea.Batch(fetchControllerCmd(), fetchDevicesCmd(), tickCmd())

	case controllerMsg:
		m.controller = msg.info
		return m, nil

	case deviceListMsg:
		m.handleDeviceList(msg)
		return m, nil
	case deviceActionMsg:
		return m, m.handleDeviceAction(msg)
	case powerActionMsg:
		return m, m.handlePowerAction(msg)
	case scanDoneMsg:
		return m, m.handleScanDone(msg)
	case detailMsg:
		m.handleDetail(msg)
		return m, nil

	case tea.MouseWheelMsg:
		if m.detail.open {
			return m, m.updateDetail(msg)
		}
		return m, nil

	case tea.KeyPressMsg:
		if m.checkErr != nil {
			if key.Matches(msg, keys.Keys.Bluetooth.Refresh) {
				return m, checkCmd()
			}
			return m, nil
		}

		if m.detail.open {
			return m, m.updateDetail(msg)
		}

		if !m.IsEditing() && key.Matches(msg, keys.Keys.Global.Help) {
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}

		return m, m.updateList(msg)
	}
	return m, nil
}
