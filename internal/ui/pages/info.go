package pages

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/anotherhadi/foam/commands"
	"github.com/anotherhadi/foam/draw"
	"github.com/anotherhadi/foam/overlay"
)

type infoBox struct {
	title string
	lines []string
	off   int
	rows  int
}

func openInfo(title string, lines []string) tea.Cmd {
	return overlay.Open(&infoBox{title: title, lines: lines})
}

func (m *infoBox) scroll(d int) tea.Cmd {
	m.off = max(m.off+d, 0)
	return nil
}

func (m *infoBox) Scope() commands.Scope {
	return commands.Scope{Name: m.title, Commands: []commands.Command{
		{ID: "help.close", Help: "close", Run: func() tea.Cmd { return overlay.Close }},
		{ID: "help.down", Help: "scroll", Run: func() tea.Cmd { return m.scroll(1) }},
		{ID: "help.up", Run: func() tea.Cmd { return m.scroll(-1) }},
		{ID: "help.page-down", Run: func() tea.Cmd { return m.scroll(max(m.rows-1, 1)) }},
		{ID: "help.page-up", Run: func() tea.Cmd { return m.scroll(-max(m.rows-1, 1)) }},
		{ID: "help.top", Run: func() tea.Cmd { m.off = 0; return nil }},
		{ID: "help.bottom", Run: func() tea.Cmd { m.off = 1 << 30; return nil }},
	}}
}

func (m *infoBox) Update(msg tea.Msg) tea.Cmd {
	if w, ok := msg.(tea.MouseWheelMsg); ok {
		switch w.Button {
		case tea.MouseWheelUp:
			return m.scroll(-3)
		case tea.MouseWheelDown:
			return m.scroll(3)
		}
	}
	return nil
}

func (m *infoBox) View(maxW, maxH int) string {
	w := min(64, maxW)
	inner := max(w-4, 0)
	var lines []string
	for _, l := range m.lines {
		for _, r := range strings.Split(ansi.Wrap(l, max(inner, 1), ""), "\n") {
			lines = append(lines, " "+r)
		}
	}
	m.rows = max(min(len(lines), maxH-2), 0)
	m.off = min(m.off, max(len(lines)-m.rows, 0))
	return draw.Frame(strings.Join(lines[m.off:m.off+m.rows], "\n"), w, m.rows+2, draw.FrameOpts{
		Title: m.title, Active: true,
		Scroll: draw.Scroll{Total: len(lines), Visible: m.rows, Offset: m.off},
	})
}
