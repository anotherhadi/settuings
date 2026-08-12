package network

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anotherhadi/settuings/internal/network"
)

const tickInterval = 5 * time.Second

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

func checkCmd() tea.Cmd {
	return func() tea.Msg { return checkMsg{err: network.CheckAvailable()} }
}

func fetchWifiCmd() tea.Cmd {
	return func() tea.Msg {
		radioOn, _ := network.RadioEnabled()
		networks, err := network.ListWifi()
		if err != nil {
			return wifiListMsg{radioOn: radioOn, err: err}
		}
		var ip, device, connectionName string
		if dev, err := network.WifiDevice(); err == nil {
			device = dev.Name
			if dev.Connected() {
				ip, _ = network.DeviceIP4(dev.Name)
				connectionName = dev.Connection
			}
		}
		return wifiListMsg{networks: networks, ip: ip, device: device, connectionName: connectionName, radioOn: radioOn}
	}
}

func toggleWifiRadioCmd(on bool) tea.Cmd {
	return func() tea.Msg { return radioActionMsg{err: network.SetWifiRadio(on)} }
}

func rescanWifiCmd() tea.Cmd {
	return func() tea.Msg { return wifiActionMsg{err: network.Rescan()} }
}

func connectWifiCmd(ssid, password string) tea.Cmd {
	return func() tea.Msg { return wifiActionMsg{err: network.ConnectWifi(ssid, password)} }
}

func disconnectWifiCmd() tea.Cmd {
	return func() tea.Msg {
		dev, err := network.WifiDevice()
		if err != nil {
			return wifiActionMsg{err: err}
		}
		return wifiActionMsg{err: network.DisconnectWifi(dev.Name)}
	}
}

func fetchEthernetCmd() tea.Cmd {
	return func() tea.Msg {
		devices, err := network.ListEthernet()
		return ethernetListMsg{devices: devices, err: err}
	}
}

func ethernetConnectCmd(device string) tea.Cmd {
	return func() tea.Msg { return ethernetActionMsg{err: network.DeviceConnect(device)} }
}

func ethernetDownCmd(device string) tea.Cmd {
	return func() tea.Msg { return ethernetActionMsg{err: network.EthernetDown(device)} }
}

func fetchVPNCmd() tea.Cmd {
	return func() tea.Msg {
		conns, err := network.ListVPN()
		return vpnListMsg{conns: conns, err: err}
	}
}

func vpnUpCmd(name string) tea.Cmd {
	return func() tea.Msg { return vpnActionMsg{err: network.VPNUp(name)} }
}

func vpnDownCmd(name string) tea.Cmd {
	return func() tea.Msg { return vpnActionMsg{err: network.VPNDown(name)} }
}

func fetchKnownCmd() tea.Cmd {
	return func() tea.Msg {
		networks, err := network.ListKnownWifi()
		return knownListMsg{networks: networks, err: err}
	}
}

func fetchSecretCmd(name string) tea.Cmd {
	return func() tea.Msg {
		secret, err := network.Secret(name)
		return knownSecretMsg{name: name, secret: secret, err: err}
	}
}

func toggleAutoConnectCmd(name string, enabled bool) tea.Cmd {
	return func() tea.Msg { return knownActionMsg{err: network.SetAutoConnect(name, enabled)} }
}

func forgetCmd(name string) tea.Cmd {
	return func() tea.Msg { return knownActionMsg{err: network.Forget(name)} }
}

func fetchDeviceDetailCmd(device string) tea.Cmd {
	return func() tea.Msg {
		detail, err := network.DeviceDetails(device)
		return deviceDetailMsg{detail: detail, err: err}
	}
}
