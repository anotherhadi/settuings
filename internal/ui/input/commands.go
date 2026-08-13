package input

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anotherhadi/settuings/internal/input"
)

const tickInterval = 5 * time.Second

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

func checkCmd() tea.Cmd {
	return func() tea.Msg { return checkMsg{err: input.CheckAvailable()} }
}

func fetchLayoutCmd() tea.Cmd {
	return func() tea.Msg {
		layout, err := input.CurrentLayout()
		return layoutMsg{layout: layout, err: err}
	}
}

func setLayoutCmd(code string) tea.Cmd {
	return func() tea.Msg { return setLayoutMsg{err: input.SetLayout(code)} }
}

func fetchSensitivityCmd() tea.Cmd {
	return func() tea.Msg {
		value, err := input.CurrentSensitivity()
		return sensitivityMsg{value: value, err: err}
	}
}

func setSensitivityCmd(v float64) tea.Cmd {
	return func() tea.Msg { return setSensitivityMsg{err: input.SetSensitivity(v)} }
}

func resetCmd() tea.Cmd {
	return func() tea.Msg { return resetMsg{err: input.Reset()} }
}
