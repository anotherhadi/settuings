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

type inputState struct {
	loading  bool
	err      error
	devices  []audio.Device
	cursor   int
	pending  bool
	feedback string

	testing bool
	monitor *audio.LevelMonitor
	level   float64 // last peak reading in [0, 1], only meaningful while testing
}

func newInputState() inputState {
	return inputState{loading: true}
}

// stopLevelMonitor tears down a running live input-level test, if any. Safe
// to call unconditionally.
func (s *inputState) stopLevelMonitor() {
	if s.monitor != nil {
		s.monitor.Stop()
		s.monitor = nil
	}
	s.testing = false
	s.level = 0
}

func (m *Model) handleInputList(msg inputListMsg) {
	m.input.loading = false
	m.input.err = msg.err
	if msg.err != nil {
		return
	}
	m.input.devices = msg.devices
	if m.input.cursor >= len(m.input.devices) {
		m.input.cursor = max(0, len(m.input.devices)-1)
	}
}

func (m *Model) handleInputAction(msg inputActionMsg) tea.Cmd {
	m.input.pending = false
	if msg.err != nil {
		m.input.feedback = msg.err.Error()
		return nil
	}
	return fetchInputCmd()
}

// handleLevelStarted reacts to a StartLevelMonitor result. If the user
// toggled testing off again before this arrived, the freshly-started monitor
// is stopped immediately instead of being adopted.
func (m *Model) handleLevelStarted(msg levelStartedMsg) tea.Cmd {
	if msg.err != nil {
		m.input.testing = false
		m.input.feedback = msg.err.Error()
		return nil
	}
	if !m.input.testing {
		msg.monitor.Stop()
		return nil
	}
	m.input.monitor = msg.monitor
	return waitLevelCmd(msg.monitor)
}

// handleLevel reacts to one peak reading, re-arming the wait for the next
// one as long as the test is still active.
func (m *Model) handleLevel(msg levelMsg) tea.Cmd {
	if !m.input.testing || m.input.monitor == nil {
		return nil
	}
	if msg.err != nil {
		m.input.feedback = "Input test stopped: " + msg.err.Error()
		m.input.stopLevelMonitor()
		return nil
	}
	m.input.level = msg.peak
	return waitLevelCmd(m.input.monitor)
}

func (m Model) selectedInput() *audio.Device {
	if m.input.cursor < 0 || m.input.cursor >= len(m.input.devices) {
		return nil
	}
	return &m.input.devices[m.input.cursor]
}

func (m *Model) updateInput(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	g := keys.Keys.Global
	a := keys.Keys.Audio

	switch {
	case key.Matches(keyMsg, g.Up):
		if m.input.cursor > 0 {
			m.input.stopLevelMonitor()
			m.input.cursor--
		}
	case key.Matches(keyMsg, g.Down):
		if m.input.cursor < len(m.input.devices)-1 {
			m.input.stopLevelMonitor()
			m.input.cursor++
		}
	case key.Matches(keyMsg, g.Left):
		return m.adjustInputVolume(-volumeStep)
	case key.Matches(keyMsg, g.Right):
		return m.adjustInputVolume(volumeStep)
	case key.Matches(keyMsg, a.Refresh):
		m.input.pending = true
		return fetchInputCmd()
	case key.Matches(keyMsg, a.Mute):
		dev := m.selectedInput()
		if dev == nil {
			return nil
		}
		m.input.feedback = ""
		m.input.pending = true
		return inputToggleMuteCmd(dev.ID)
	case key.Matches(keyMsg, a.Select):
		dev := m.selectedInput()
		if dev == nil {
			return nil
		}
		if dev.Default {
			m.input.feedback = "Already the default input."
			return nil
		}
		m.input.feedback = ""
		m.input.pending = true
		return inputSetDefaultCmd(dev.ID)
	case key.Matches(keyMsg, a.Test):
		dev := m.selectedInput()
		if dev == nil {
			return nil
		}
		if m.input.testing {
			m.input.stopLevelMonitor()
			return nil
		}
		m.input.testing = true
		m.input.feedback = ""
		return startLevelMonitorCmd(dev.ID)
	}
	return nil
}

func (m *Model) adjustInputVolume(delta int) tea.Cmd {
	dev := m.selectedInput()
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
	m.input.feedback = ""
	return inputSetVolumeCmd(dev.ID, pct)
}

func (m Model) renderInput(width, height int) string {
	if m.input.loading {
		return ilovetui.S.Faint.Render("Loading input devices…")
	}

	header := lipgloss.NewStyle().Bold(true).Foreground(ilovetui.S.Subtle).Render("INPUT DEVICES")
	if m.input.pending {
		header += mutedStyle().Render("  refreshing…")
	}

	if m.input.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", errorStyle().Render(m.input.err.Error()))
	}
	if len(m.input.devices) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", ilovetui.S.Faint.Render("No input devices found."))
	}

	available := (height - 3) / deviceRowHeight
	if available < 1 {
		available = 1
	}
	p, start, end := paginate(m.input.cursor, len(m.input.devices), available)

	var rows []string
	for i := start; i < end; i++ {
		selected := i == m.input.cursor
		row := renderDeviceRow(m.input.devices[i], selected, width)
		if selected && m.input.testing {
			row += "\n    " + renderLevelMeter(m.input.level, barWidth(width))
		}
		rows = append(rows, row)
	}
	list := strings.Join(rows, "\n\n")

	footer := renderPaginatorDots(p)
	if m.input.feedback != "" {
		footer = subtleStyle().Render(m.input.feedback)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, "", list, footer)
}
