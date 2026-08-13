package input

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	inputpkg "github.com/anotherhadi/settuings/internal/input"
	"github.com/anotherhadi/settuings/internal/keys"
)

const sensitivityStep = 0.05

func (m *Model) handleSensitivity(msg sensitivityMsg) {
	m.sensitivityErr = msg.err
	if msg.err != nil {
		return
	}
	m.sensitivity = msg.value
	m.sensitivityKnown = true
}

func (m *Model) handleSetSensitivity(msg setSensitivityMsg) tea.Cmd {
	m.mousePending = false
	if msg.err != nil {
		m.mouseFeedback = msg.err.Error()
		return nil
	}
	m.mouseFeedback = ""
	return fetchSensitivityCmd()
}

func (m *Model) updateMouse(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	g := keys.Keys.Global

	switch {
	case key.Matches(keyMsg, g.Left):
		return m.adjustSensitivity(-sensitivityStep)
	case key.Matches(keyMsg, g.Right):
		return m.adjustSensitivity(sensitivityStep)
	}
	return nil
}

// adjustSensitivity nudges the mouse sensitivity by delta, updating it
// locally right away so the bar responds instantly instead of waiting on
// the round trip to hyprctl and back, mirroring the audio page's volume
// adjustment.
func (m *Model) adjustSensitivity(delta float64) tea.Cmd {
	if !m.sensitivityKnown {
		return nil
	}
	v := m.sensitivity + delta
	switch {
	case v < inputpkg.MinSensitivity:
		v = inputpkg.MinSensitivity
	case v > inputpkg.MaxSensitivity:
		v = inputpkg.MaxSensitivity
	}
	m.sensitivity = v
	m.mouseFeedback = ""
	m.mousePending = true
	return setSensitivityCmd(v)
}

func (m Model) renderMouse(width int) string {
	title := sectionTitleStyle(m.focus == focusMouse).Render("Mouse Speed")

	switch {
	case m.sensitivityErr != nil:
		return lipgloss.JoinVertical(lipgloss.Left, title, "", errorStyle().Render(m.sensitivityErr.Error()))
	case !m.sensitivityKnown:
		return lipgloss.JoinVertical(lipgloss.Left, title, "", ilovetui.S.Faint.Render("Loading mouse sensitivity…"))
	}

	lines := []string{title, "", renderSensitivityBar(m.sensitivity, width)}

	switch {
	case m.mousePending:
		lines = append(lines, "", mutedStyle().Render("applying…"))
	case m.mouseFeedback != "":
		lines = append(lines, "", errorStyle().Render(m.mouseFeedback))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderSensitivityBar draws a static block-bar gauge for a sensitivity in
// [MinSensitivity, MaxSensitivity], in the same spirit as the audio page's
// volume bar and the power page's battery gauge.
func renderSensitivityBar(value float64, width int) string {
	barWidth := width / 2
	switch {
	case barWidth < 10:
		barWidth = 10
	case barWidth > 30:
		barWidth = 30
	}

	span := inputpkg.MaxSensitivity - inputpkg.MinSensitivity
	filled := int((value-inputpkg.MinSensitivity)/span*float64(barWidth) + 0.5)
	switch {
	case filled > barWidth:
		filled = barWidth
	case filled < 0:
		filled = 0
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	gauge := lipgloss.NewStyle().Foreground(ilovetui.S.Primary).Render("[" + bar + "]")
	return fmt.Sprintf("%s %+.2f", gauge, value)
}
