package about

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/config"
	"github.com/anotherhadi/settuings/internal/util"
	"github.com/charmbracelet/x/ansi"
)

func windowStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ilovetui.S.Subtle).
		Padding(0, 0)
}

func (m Model) View() tea.View {
	statusBar := m.renderStatusBar()
	if len(m.matches) > 0 {
		var countText string
		if m.searching {
			countText = fmt.Sprintf("%d matches", len(m.matches))
		} else {
			countText = fmt.Sprintf("%d/%d", m.matchIndex+1, len(m.matches))
		}
		count := lipgloss.NewStyle().Padding(0, 1).
			Foreground(ilovetui.S.Muted).
			Render(countText)
		statusBar = lipgloss.JoinHorizontal(lipgloss.Top, statusBar, count)
	}
	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left,
		windowStyle().Render(m.viewport.View()),
		statusBar,
	))
}

func (m *Model) renderMarkdown() {
	cfg := config.Global
	data := struct {
		Cfg      *config.Config
		OS       string
		Kernel   string
		Hostname string
		Arch     string
	}{
		Cfg:      cfg,
		OS:       util.OSName(),
		Kernel:   util.KernelVersion(),
		Hostname: util.Hostname(),
		Arch:     util.Arch(),
	}

	tmpl, err := template.New("about").Parse(contentMarkdown)
	if err != nil {
		return
	}

	var processed bytes.Buffer
	if err := tmpl.Execute(&processed, data); err != nil {
		return
	}

	width := m.viewport.Width() - util.ScrollbarGutterWidth
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStyles(ilovetui.GlamourStyleConfig()),
		glamour.WithWordWrap(width),
	)

	str, _ := renderer.Render(processed.String())
	m.renderedLines = strings.Split(str, "\n")
	m.strippedLines = make([]string, len(m.renderedLines))
	for i, l := range m.renderedLines {
		m.strippedLines[i] = ansi.Strip(l)
	}
	m.applySearch()
	util.RefreshScrollbar(&m.viewport)
}
