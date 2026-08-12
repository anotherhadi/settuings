package network

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/anotherhadi/settuings/internal/keys"
	"github.com/anotherhadi/settuings/internal/util"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case checkMsg:
		wasUnavailable := m.checkErr != nil
		m.checked = true
		m.checkErr = msg.err
		if msg.err == nil && wasUnavailable {
			// recovered: resume fetching + the auto-refresh loop
			return m, tea.Batch(fetchWifiCmd(), fetchEthernetCmd(), fetchVPNCmd(), fetchKnownCmd(), tickCmd())
		}
		return m, nil

	case tickMsg:
		if m.checkErr != nil {
			return m, nil
		}
		return m, tea.Batch(fetchWifiCmd(), fetchEthernetCmd(), fetchVPNCmd(), fetchKnownCmd(), tickCmd())

	case wifiListMsg:
		m.handleWifiList(msg)
		return m, nil
	case wifiActionMsg:
		return m, m.handleWifiAction(msg)
	case radioActionMsg:
		return m, m.handleRadioAction(msg)

	case ethernetListMsg:
		m.handleEthernetList(msg)
		return m, nil
	case ethernetActionMsg:
		return m, m.handleEthernetAction(msg)

	case vpnListMsg:
		m.handleVPNList(msg)
		return m, nil
	case vpnActionMsg:
		return m, m.handleVPNAction(msg)

	case knownListMsg:
		m.handleKnownList(msg)
		return m, nil
	case knownSecretMsg:
		m.handleKnownSecret(msg)
		m.handleDeviceDetailSecret(msg)
		return m, nil
	case knownActionMsg:
		return m, m.handleKnownAction(msg)

	case deviceDetailMsg:
		m.handleDeviceDetail(msg)
		return m, nil

	case tea.MouseWheelMsg:
		if m.deviceDetail.open {
			return m, m.updateDeviceDetail(msg)
		}
		if m.active == tabKnown && m.known.detail {
			util.HandleMouseWheel(msg, &m.known.viewport)
		}
		return m, nil

	case tea.KeyPressMsg:
		if m.checkErr != nil {
			if key.Matches(msg, keys.Keys.Network.Refresh) {
				return m, checkCmd()
			}
			return m, nil
		}

		if m.deviceDetail.open {
			return m, m.updateDeviceDetail(msg)
		}

		g := keys.Keys.Global
		n := keys.Keys.Network

		if !m.IsEditing() {
			switch {
			case key.Matches(msg, g.CycleFocus):
				m.cycleTab()
				return m, nil
			case key.Matches(msg, n.Known):
				m.toggleKnown()
				return m, nil
			case key.Matches(msg, g.Help):
				m.help.ShowAll = !m.help.ShowAll
				m.resizeDetails()
				return m, nil
			}
		}

		var cmd tea.Cmd
		switch m.active {
		case tabWifi:
			cmd = m.updateWifi(msg)
		case tabEthernet:
			cmd = m.updateEthernet(msg)
		case tabVPN:
			cmd = m.updateVPN(msg)
		case tabKnown:
			cmd = m.updateKnown(msg)
		}
		return m, cmd
	}
	return m, nil
}
