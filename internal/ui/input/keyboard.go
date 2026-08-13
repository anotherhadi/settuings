package input

import (
	"fmt"
	"strings"

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

// visibleLayouts returns the layouts matching the current filter query (by
// code or name, case-insensitive), or every layout when there's no query.
func (m Model) visibleLayouts() []inputpkg.Layout {
	q := strings.ToLower(strings.TrimSpace(m.filterInput.Value()))
	if q == "" {
		return inputpkg.Layouts
	}
	matches := make([]inputpkg.Layout, 0, len(inputpkg.Layouts))
	for _, l := range inputpkg.Layouts {
		if strings.Contains(strings.ToLower(l.Name), q) || strings.Contains(strings.ToLower(l.Code), q) {
			matches = append(matches, l)
		}
	}
	return matches
}

// clampCursor keeps the cursor within the currently visible (filtered)
// list, called whenever the filter query — and so the list length —
// changes.
func (m *Model) clampCursor() {
	n := len(m.visibleLayouts())
	switch {
	case m.cursor >= n:
		m.cursor = n - 1
	case m.cursor < 0:
		m.cursor = 0
	}
}

func (m Model) selectedLayout() *inputpkg.Layout {
	list := m.visibleLayouts()
	if m.cursor < 0 || m.cursor >= len(list) {
		return nil
	}
	return &list[m.cursor]
}

func (m *Model) updateLayout(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	g := keys.Keys.Global
	in := keys.Keys.Input

	if m.filtering {
		switch {
		case key.Matches(keyMsg, g.Escape):
			m.filtering = false
			m.filterInput.Blur()
			m.filterInput.SetValue("")
			m.cursor = 0
		case key.Matches(keyMsg, in.Select):
			m.filtering = false
			m.filterInput.Blur()
			m.clampCursor()
		default:
			var cmd tea.Cmd
			m.filterInput, cmd = m.filterInput.Update(keyMsg)
			m.clampCursor()
			return cmd
		}
		return nil
	}

	switch {
	case key.Matches(keyMsg, in.Filter):
		m.filtering = true
		m.filterInput.Focus()
	case key.Matches(keyMsg, g.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(keyMsg, g.Down):
		if m.cursor < len(m.visibleLayouts())-1 {
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
	lines := []string{title, "", current}

	switch {
	case m.filtering:
		lines = append(lines, m.filterInput.View())
	case m.filterInput.Value() != "":
		lines = append(lines, subtleStyle().Render(fmt.Sprintf("Filter: %q — press / to change", m.filterInput.Value())))
	}
	lines = append(lines, "")

	list := m.visibleLayouts()

	available := height - len(lines) - 2
	if available < 1 {
		available = 1
	}

	if len(list) == 0 {
		lines = append(lines, ilovetui.S.Faint.Render("No layout matches the filter."))
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	p, start, end := paginate(m.cursor, len(list), available)

	for i := start; i < end; i++ {
		lines = append(lines, m.renderLayoutRow(list[i], i == m.cursor))
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
