// Package power provides the "Power" TUI page: power-profile selection (via
// powerprofilesctl) and battery/AC status, in the same spirit as the
// bluetooth page's controller power + device state.
package power

import (
	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/power"
)

type Model struct {
	width  int
	height int
	help   help.Model

	checkErr error // reason powerprofilesctl is unusable; nil once confirmed OK
	checked  bool

	profilesLoading bool
	profilesErr     error
	profiles        []power.Profile
	cursor          int
	pending         bool
	feedback        string

	batteryLoading bool
	batteryErr     error
	batteries      []power.Battery
	ac             power.ACStatus
}

func New() Model {
	return Model{
		help:            ilovetui.NewHelp(),
		profilesLoading: true,
		batteryLoading:  true,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// Activate is called every time the user navigates onto this page. It
// re-checks powerprofilesctl availability, fetches profiles + battery
// state, and starts the periodic auto-refresh loop.
func (m *Model) Activate() tea.Cmd {
	return tea.Batch(checkCmd(), fetchProfilesCmd(), fetchBatteryCmd(), tickCmd())
}

func (m Model) IsEditing() bool {
	return false
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.help.SetWidth(w - 2)
}
