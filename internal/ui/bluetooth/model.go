// Package bluetooth provides the "Bluetooth" TUI page: a thin wrapper
// around bluetoothctl covering controller power and device
// pairing/connection, in the same spirit as the network page's Wi-Fi tab.
package bluetooth

import (
	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/bluetooth"
)

type Model struct {
	width  int
	height int
	help   help.Model

	checkErr error // reason bluetoothctl is unusable; nil once confirmed OK
	checked  bool

	controller bluetooth.ControllerInfo

	list   deviceListState
	detail deviceDetailState
}

func New() Model {
	return Model{
		help: ilovetui.NewHelp(),
		list: newDeviceListState(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// Activate is called every time the user navigates onto this page. It
// re-checks bluetoothctl availability, fetches controller + device state,
// and starts the periodic auto-refresh loop.
func (m *Model) Activate() tea.Cmd {
	return tea.Batch(checkCmd(), fetchControllerCmd(), fetchDevicesCmd(), tickCmd())
}

func (m Model) IsEditing() bool {
	return m.list.confirmRemove
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.help.SetWidth(w - 2)
	m.resizeDetail()
}
