package input

import tea "charm.land/bubbletea/v2"

// applyReset kicks off a `hyprctl reload`, which restores the keyboard
// layout and mouse sensitivity to their hyprland.conf values (along with
// anything else set there); both sections are marked pending since a
// reload touches them together.
func (m *Model) applyReset() tea.Cmd {
	m.layoutPending = true
	m.mousePending = true
	m.layoutFeedback = ""
	m.mouseFeedback = ""
	return resetCmd()
}

func (m *Model) handleReset(msg resetMsg) tea.Cmd {
	m.layoutPending = false
	m.mousePending = false
	if msg.err != nil {
		m.layoutFeedback = msg.err.Error()
		m.mouseFeedback = msg.err.Error()
		return nil
	}
	return tea.Batch(fetchLayoutCmd(), fetchSensitivityCmd())
}
