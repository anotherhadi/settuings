package bluetooth

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anotherhadi/settuings/internal/bluetooth"
)

const (
	tickInterval = 5 * time.Second
	scanSeconds  = 5
)

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

func checkCmd() tea.Cmd {
	return func() tea.Msg { return checkMsg{err: bluetooth.CheckAvailable()} }
}

func fetchControllerCmd() tea.Cmd {
	return func() tea.Msg {
		info, err := bluetooth.ShowController()
		return controllerMsg{info: info, err: err}
	}
}

func fetchDevicesCmd() tea.Cmd {
	return func() tea.Msg {
		devices, err := bluetooth.ListDevices()
		return deviceListMsg{devices: devices, err: err}
	}
}

func togglePowerCmd(on bool) tea.Cmd {
	return func() tea.Msg { return powerActionMsg{err: bluetooth.SetPower(on)} }
}

func scanCmd() tea.Cmd {
	return func() tea.Msg { return scanDoneMsg{err: bluetooth.Scan(scanSeconds)} }
}

func connectCmd(address string) tea.Cmd {
	return func() tea.Msg { return deviceActionMsg{err: bluetooth.Connect(address)} }
}

func disconnectCmd(address string) tea.Cmd {
	return func() tea.Msg { return deviceActionMsg{err: bluetooth.Disconnect(address)} }
}

// pairCmd pairs with a device and, on success, immediately connects it —
// most devices don't auto-connect right after pairing.
func pairCmd(address string) tea.Cmd {
	return func() tea.Msg {
		if err := bluetooth.Pair(address); err != nil {
			return deviceActionMsg{err: err}
		}
		return deviceActionMsg{err: bluetooth.Connect(address)}
	}
}

func trustCmd(address string, trust bool) tea.Cmd {
	return func() tea.Msg { return deviceActionMsg{err: bluetooth.SetTrusted(address, trust)} }
}

func removeCmd(address string) tea.Cmd {
	return func() tea.Msg { return deviceActionMsg{err: bluetooth.Remove(address)} }
}

func fetchDetailCmd(address string) tea.Cmd {
	return func() tea.Msg {
		detail, err := bluetooth.Info(address)
		return detailMsg{detail: detail, err: err}
	}
}
