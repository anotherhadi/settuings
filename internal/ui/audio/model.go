// Package audio provides the "Audio" TUI page: a thin, tabbed wrapper around
// wpctl (and the pw-play/pw-cat helper tools that ship with it) covering
// output devices, input devices, and per-application streams, in the same
// spirit as the network page's tabs.
package audio

import (
	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/icons"
)

type tab int

const (
	tabOutput tab = iota
	tabInput
	tabApps
)

var visibleTabs = []tab{tabOutput, tabInput, tabApps}

func (t tab) label() string {
	switch t {
	case tabOutput:
		return "Output"
	case tabInput:
		return "Input"
	case tabApps:
		return "Apps"
	default:
		return ""
	}
}

func (t tab) icon() string {
	switch t {
	case tabOutput:
		return icons.I.Output
	case tabInput:
		return icons.I.Input
	case tabApps:
		return icons.I.Apps
	default:
		return ""
	}
}

type Model struct {
	active tab
	width  int
	height int
	help   help.Model

	checkErr error // reason wpctl is unusable; nil once confirmed OK
	checked  bool

	output outputState
	input  inputState
	apps   appsState
}

func New() Model {
	return Model{
		active: tabOutput,
		help:   ilovetui.NewHelp(),
		output: newOutputState(),
		input:  newInputState(),
		apps:   newAppsState(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// Activate is called every time the user navigates onto this page. It
// re-checks wpctl availability, kicks off a fetch for every tab, and starts
// the periodic auto-refresh loop.
func (m *Model) Activate() tea.Cmd {
	return tea.Batch(
		checkCmd(),
		fetchOutputCmd(),
		fetchInputCmd(),
		fetchAppsCmd(),
		tickCmd(),
	)
}

// Deactivate is called synchronously when the user navigates away from this
// page (or quits) so the live input-level meter, if running, never leaks a
// background pw-cat process.
func (m *Model) Deactivate() {
	m.input.stopLevelMonitor()
}

func (m Model) IsEditing() bool {
	return false
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.help.SetWidth(w - 2)
}
