// Package input provides the "Inputs" TUI page: keyboard layout and mouse
// sensitivity, controlled via hyprctl, in the same spirit as the power
// page's controller-power + status split.
package input

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
)

// focusZone tracks which of the two sections (keyboard layout list, mouse
// sensitivity slider) currently receives Up/Down/Left/Right/Select.
type focusZone int

const (
	focusLayout focusZone = iota
	focusMouse
)

type Model struct {
	width  int
	height int
	help   help.Model
	focus  focusZone

	checkErr error // reason hyprctl is unusable; nil once confirmed OK
	checked  bool

	layoutErr      error
	currentLayout  string
	cursor         int
	cursorInit     bool // whether cursor has been snapped to the current layout yet
	layoutPending  bool
	layoutFeedback string
	filterInput    textinput.Model
	filtering      bool

	sensitivityErr   error
	sensitivity      float64
	sensitivityKnown bool
	mousePending     bool
	mouseFeedback    string
}

func New() Model {
	ti := textinput.New()
	ti.Prompt = "/"
	s := ti.Styles()
	s.Focused.Prompt = lipgloss.NewStyle().Foreground(ilovetui.S.Primary)
	ti.SetStyles(s)

	return Model{
		help:        ilovetui.NewHelp(),
		filterInput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// Activate is called every time the user navigates onto this page. It
// re-checks hyprctl availability and fetches the current layout + mouse
// sensitivity, then starts the periodic auto-refresh loop.
func (m *Model) Activate() tea.Cmd {
	return tea.Batch(checkCmd(), fetchLayoutCmd(), fetchSensitivityCmd(), tickCmd())
}

func (m Model) IsEditing() bool {
	return m.filtering
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.help.SetWidth(w - 2)
	m.filterInput.SetWidth(w - 4)
}
