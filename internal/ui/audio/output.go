package audio

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/audio"
	"github.com/anotherhadi/settuings/internal/keys"
)

const volumeStep = 5

type outputState struct {
	loading  bool
	err      error
	devices  []audio.Device
	cursor   int
	pending  bool
	feedback string
}

func newOutputState() outputState {
	return outputState{loading: true}
}

func (m *Model) handleOutputList(msg outputListMsg) {
	m.output.loading = false
	m.output.err = msg.err
	if msg.err != nil {
		return
	}
	m.output.devices = msg.devices
	if m.output.cursor >= len(m.output.devices) {
		m.output.cursor = max(0, len(m.output.devices)-1)
	}
}

func (m *Model) handleOutputAction(msg outputActionMsg) tea.Cmd {
	m.output.pending = false
	if msg.err != nil {
		m.output.feedback = msg.err.Error()
		return nil
	}
	return fetchOutputCmd()
}

func (m Model) selectedOutput() *audio.Device {
	if m.output.cursor < 0 || m.output.cursor >= len(m.output.devices) {
		return nil
	}
	return &m.output.devices[m.output.cursor]
}

func (m *Model) updateOutput(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	g := keys.Keys.Global
	a := keys.Keys.Audio

	switch {
	case key.Matches(keyMsg, g.Up):
		if m.output.cursor > 0 {
			m.output.cursor--
		}
	case key.Matches(keyMsg, g.Down):
		if m.output.cursor < len(m.output.devices)-1 {
			m.output.cursor++
		}
	case key.Matches(keyMsg, g.Left):
		return m.adjustOutputVolume(-volumeStep)
	case key.Matches(keyMsg, g.Right):
		return m.adjustOutputVolume(volumeStep)
	case key.Matches(keyMsg, a.Refresh):
		m.output.pending = true
		return fetchOutputCmd()
	case key.Matches(keyMsg, a.Mute):
		dev := m.selectedOutput()
		if dev == nil {
			return nil
		}
		m.output.feedback = ""
		m.output.pending = true
		return outputToggleMuteCmd(dev.ID)
	case key.Matches(keyMsg, a.Select):
		dev := m.selectedOutput()
		if dev == nil {
			return nil
		}
		if dev.Default {
			m.output.feedback = "Already the default output."
			return nil
		}
		m.output.feedback = ""
		m.output.pending = true
		return outputSetDefaultCmd(dev.ID)
	case key.Matches(keyMsg, a.Test):
		dev := m.selectedOutput()
		if dev == nil {
			return nil
		}
		return tea.Batch(outputTestToneCmd(dev.ID), notifyCmd("Audio", "Playing test sound on "+dev.Name+"…"))
	}
	return nil
}

// adjustOutputVolume nudges the selected device's volume by delta, updating
// it locally right away so the bar responds instantly instead of waiting on
// the round trip to wpctl and back.
func (m *Model) adjustOutputVolume(delta int) tea.Cmd {
	dev := m.selectedOutput()
	if dev == nil {
		return nil
	}
	pct := dev.Volume + delta
	if pct < 0 {
		pct = 0
	} else if pct > 100 {
		pct = 100
	}
	dev.Volume = pct
	m.output.feedback = ""
	return outputSetVolumeCmd(dev.ID, pct)
}

func (m Model) renderOutput(width, height int) string {
	if m.output.loading {
		return ilovetui.S.Faint.Render("Loading output devices…")
	}

	header := lipgloss.NewStyle().Bold(true).Foreground(ilovetui.S.Subtle).Render("OUTPUT DEVICES")
	if m.output.pending {
		header += mutedStyle().Render("  refreshing…")
	}

	if m.output.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", errorStyle().Render(m.output.err.Error()))
	}
	if len(m.output.devices) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", ilovetui.S.Faint.Render("No output devices found."))
	}

	available := (height - 3) / deviceRowHeight
	if available < 1 {
		available = 1
	}
	p, start, end := paginate(m.output.cursor, len(m.output.devices), available)

	var rows []string
	for i := start; i < end; i++ {
		rows = append(rows, renderDeviceRow(m.output.devices[i], i == m.output.cursor, width))
	}
	list := strings.Join(rows, "\n\n")

	footer := renderPaginatorDots(p)
	if m.output.feedback != "" {
		footer = subtleStyle().Render(m.output.feedback)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, "", list, footer)
}
