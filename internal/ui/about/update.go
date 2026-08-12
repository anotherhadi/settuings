package about

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/anotherhadi/settuings/internal/keys"
	"github.com/anotherhadi/settuings/internal/util"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	g := keys.Keys.Global
	d := keys.Keys.Docs

	switch msg := msg.(type) {
	case tea.MouseWheelMsg:
		util.HandleMouseWheel(msg, &m.viewport)

	case tea.KeyPressMsg:
		if m.searching {
			switch {
			case key.Matches(msg, g.Escape):
				m.searching = false
				m.searchInput.Blur()
				m.SetSize(m.width, m.height)
			case msg.String() == "enter":
				m.searching = false
				m.searchInput.Blur()
				m.SetSize(m.width, m.height)
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.applySearch()
				return m, cmd
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, d.Search):
			m.searching = true
			m.searchInput.SetValue("")
			m.searchInput.Focus()
			m.SetSize(m.width, m.height)
		case key.Matches(msg, d.SearchReset):
			m.matches = nil
			m.matchIndex = 0
			m.rebuildViewportContent()
		case key.Matches(msg, d.SearchNext):
			m.searchNext()
		case key.Matches(msg, d.SearchPrev):
			m.searchPrev()
		case key.Matches(msg, g.Up):
			util.ScrollLines(&m.viewport, -1)
		case key.Matches(msg, g.Down):
			util.ScrollLines(&m.viewport, 1)
		case key.Matches(msg, g.ScrollUp):
			util.ScrollViewport(&m.viewport, -1)
		case key.Matches(msg, g.ScrollDown):
			util.ScrollViewport(&m.viewport, 1)
		case key.Matches(msg, g.Help):
			m.help.ShowAll = !m.help.ShowAll
			m.SetSize(m.width, m.height)
		}
	}
	return m, nil
}
