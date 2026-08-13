package input

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/anotherhadi/settuings/internal/keys"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case checkMsg:
		wasUnavailable := m.checkErr != nil
		m.checked = true
		m.checkErr = msg.err
		if msg.err == nil && wasUnavailable {
			// recovered: resume fetching
			return m, tea.Batch(fetchLayoutCmd(), fetchSensitivityCmd())
		}
		return m, nil

	case tickMsg:
		if m.checkErr != nil {
			// Not manually retriable anymore (no refresh key): keep
			// re-checking on the tick so the page self-heals once
			// hyprctl/Hyprland becomes available again.
			return m, tea.Batch(tickCmd(), checkCmd())
		}
		return m, tea.Batch(tickCmd(), fetchLayoutCmd(), fetchSensitivityCmd())

	case layoutMsg:
		m.handleLayout(msg)
		return m, nil
	case setLayoutMsg:
		return m, m.handleSetLayout(msg)

	case sensitivityMsg:
		m.handleSensitivity(msg)
		return m, nil
	case setSensitivityMsg:
		return m, m.handleSetSensitivity(msg)

	case resetMsg:
		return m, m.handleReset(msg)

	case tea.KeyPressMsg:
		if m.filtering {
			// Every key goes straight to the filter box while typing (a
			// query like "tab" or "r" must not also trigger CycleFocus or
			// Reset), same precedence the About page's search box uses.
			return m, m.updateLayout(msg)
		}

		if key.Matches(msg, keys.Keys.Global.Help) {
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}

		if m.checkErr != nil {
			return m, nil
		}

		if key.Matches(msg, keys.Keys.Global.CycleFocus) {
			m.cycleFocus()
			return m, nil
		}

		if key.Matches(msg, keys.Keys.Input.Reset) {
			return m, m.applyReset()
		}

		switch m.focus {
		case focusLayout:
			return m, m.updateLayout(msg)
		case focusMouse:
			return m, m.updateMouse(msg)
		}
	}
	return m, nil
}

func (m *Model) cycleFocus() {
	if m.focus == focusLayout {
		m.focus = focusMouse
	} else {
		m.focus = focusLayout
	}
}
