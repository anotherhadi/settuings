package input

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/icons"
	inputpkg "github.com/anotherhadi/settuings/internal/input"
	"github.com/anotherhadi/settuings/internal/keys"
)

func (m *Model) handleLayout(msg layoutMsg) {
	m.layoutErr = msg.err
	if msg.err != nil {
		return
	}
	m.currentLayout = msg.layout
	if !m.cursorInit {
		m.cursorInit = true
		for i, l := range inputpkg.Layouts {
			if l.Code == msg.layout {
				m.cursor = i
				break
			}
		}
	}
}

func (m *Model) handleSetLayout(msg setLayoutMsg) tea.Cmd {
	m.layoutPending = false
	if msg.err != nil {
		m.layoutFeedback = msg.err.Error()
		return nil
	}
	m.layoutFeedback = ""
	return fetchLayoutCmd()
}

func (m Model) selectedLayout() *inputpkg.Layout {
	if m.cursor < 0 || m.cursor >= len(inputpkg.Layouts) {
		return nil
	}
	return &inputpkg.Layouts[m.cursor]
}

func (m *Model) updateLayout(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	g := keys.Keys.Global
	in := keys.Keys.Input

	switch {
	case key.Matches(keyMsg, g.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(keyMsg, g.Down):
		if m.cursor < len(inputpkg.Layouts)-1 {
			m.cursor++
		}
	case key.Matches(keyMsg, in.Select):
		return m.applySelectedLayout()
	}
	return nil
}

func (m *Model) applySelectedLayout() tea.Cmd {
	sel := m.selectedLayout()
	if sel == nil || sel.Code == m.currentLayout {
		return nil
	}
	m.layoutPending = true
	m.layoutFeedback = ""
	return setLayoutCmd(sel.Code)
}

func (m Model) renderLayout(width, height int) string {
	title := sectionTitleStyle(m.focus == focusLayout).Render("Keyboard Layout")

	switch {
	case m.layoutErr != nil:
		return lipgloss.JoinVertical(lipgloss.Left, title, "", errorStyle().Render(m.layoutErr.Error()))
	case m.currentLayout == "" && !m.cursorInit:
		return lipgloss.JoinVertical(lipgloss.Left, title, "", ilovetui.S.Faint.Render("Loading current layout…"))
	}

	current := subtleStyle().Render("Current: ") + inputpkg.LayoutName(m.currentLayout) + subtleStyle().Render(" ("+m.currentLayout+")")
	lines := []string{title, "", current, ""}

	available := height - len(lines) - 2
	if available < 1 {
		available = 1
	}
	p, start, end := paginate(m.cursor, len(inputpkg.Layouts), available)

	for i := start; i < end; i++ {
		lines = append(lines, m.renderLayoutRow(inputpkg.Layouts[i], i == m.cursor))
	}

	footer := renderPaginatorDots(p)
	switch {
	case m.layoutPending:
		footer = mutedStyle().Render("applying…")
	case m.layoutFeedback != "":
		footer = errorStyle().Render(m.layoutFeedback)
	}
	lines = append(lines, "", footer)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderLayoutRow(l inputpkg.Layout, selected bool) string {
	textStyle := lipgloss.NewStyle().Foreground(ilovetui.S.Text)
	if selected {
		textStyle = textStyle.Bold(true)
		if m.focus == focusLayout {
			textStyle = textStyle.Foreground(ilovetui.S.Primary)
		}
	}
	cursor := "  "
	if selected {
		cursor = "> "
	}
	check := " "
	if l.Code == m.currentLayout {
		check = successStyle().Render(icons.I.Check)
	}
	return fmt.Sprintf("%s%s%s", cursor, check, textStyle.Render(l.Name+" ("+l.Code+")"))
}
