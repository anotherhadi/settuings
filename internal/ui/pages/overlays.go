package pages

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/anotherhadi/foam/commands"
	"github.com/anotherhadi/foam/draw"
	"github.com/anotherhadi/foam/list"
	"github.com/anotherhadi/foam/overlay"
	"github.com/anotherhadi/foam/theme"

	"github.com/anotherhadi/settuings/internal/monitor"
)

type CloseIfTop struct{ O overlay.Overlay }

func OverlayScopes() []commands.Scope {
	return []commands.Scope{
		newChoice("", nil, 0, nil).Scope(),
		(&configBox{}).Scope(),
	}
}

type choiceBox struct {
	title  string
	opts   []string
	l      list.Model
	onPick func(int) tea.Cmd
}

func newChoice(title string, opts []string, sel int, onPick func(int) tea.Cmd) *choiceBox {
	m := &choiceBox{title: title, opts: opts, onPick: onPick}
	m.l.SetLen(len(opts))
	m.l.Select(sel)
	return m
}

func (m *choiceBox) pick() tea.Cmd {
	if len(m.opts) == 0 {
		return nil
	}
	return tea.Sequence(overlay.Close, m.onPick(m.l.Selected()))
}

func (m *choiceBox) Scope() commands.Scope {
	return commands.Scope{Name: m.title, Commands: []commands.Command{
		{ID: "choice.close", Help: "close", Run: func() tea.Cmd { return overlay.Close }},
		{ID: "choice.pick", Help: "pick", Run: m.pick},
		{ID: "choice.up", Run: func() tea.Cmd { m.l.Move(-1); return nil }},
		{ID: "choice.down", Help: "move", Run: func() tea.Cmd { m.l.Move(1); return nil }},
	}}
}

func (m *choiceBox) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		mm := msg.Mouse()
		mm.Y--
		if m.l.Mouse(tea.MouseClickMsg(mm)) >= 0 {
			return m.pick()
		}
	case tea.MouseWheelMsg:
		m.l.Mouse(msg)
	}
	return nil
}

func (m *choiceBox) View(maxW, maxH int) string {
	w := 0
	for _, o := range m.opts {
		w = max(w, ansi.StringWidth(o))
	}
	w = min(max(w+4, 24), min(60, maxW))
	inner := max(w-2, 0)
	m.l.SetHeight(max(min(len(m.opts), maxH-2), 1))
	start, end := m.l.Visible()
	var lines []string
	for i := start; i < end; i++ {
		row := " " + ansi.Truncate(m.opts[i], max(inner-2, 0), "…")
		if i == m.l.Selected() {
			row = draw.Selected(row, inner, true)
		}
		lines = append(lines, row)
	}
	return draw.Frame(strings.Join(lines, "\n"), w, len(lines)+2, draw.FrameOpts{Title: m.title, Active: true, Scroll: m.l.Scroll()})
}

type configBox struct {
	outs   []monitor.Output
	kanshi bool
}

func (m *configBox) text() string {
	if m.kanshi {
		return monitor.Kanshi(m.outs)
	}
	return monitor.SwayConfig(m.outs)
}

func (m *configBox) name() string {
	if m.kanshi {
		return "kanshi"
	}
	return "sway"
}

func (m *configBox) Scope() commands.Scope {
	return commands.Scope{Name: "Config", Commands: []commands.Command{
		{ID: "config.close", Help: "close", Run: func() tea.Cmd { return overlay.Close }},
		{ID: "config.copy", Help: "copy", Run: func() tea.Cmd {
			return copyText(m.text(), "the "+m.name()+" config")
		}},
		{ID: "config.format", Help: "sway / kanshi", Run: func() tea.Cmd { m.kanshi = !m.kanshi; return nil }},
	}}
}

func (m *configBox) Update(tea.Msg) tea.Cmd { return nil }

func (m *configBox) View(maxW, maxH int) string {
	w := min(110, maxW)
	inner := max(w-4, 1)
	tab := func(name string, on bool) string {
		if on {
			return theme.Primary().Style().Bold(true).Render("[" + name + "]")
		}
		return theme.Subtle().Render(" " + name + " ")
	}
	lines := []string{" " + tab("sway", !m.kanshi) + " " + tab("kanshi", m.kanshi), ""}
	for _, l := range strings.Split(ansi.Wrap(m.text(), inner, ""), "\n") {
		lines = append(lines, " "+theme.Fg().Render(l))
	}
	dest := "~/.config/sway/config"
	if m.kanshi {
		dest = "~/.config/kanshi/config"
	}
	keys := theme.Primary().Style()
	hint := []string{
		keys.Render(keyOf("config.copy", "y")) + theme.Subtle().Render(" copy"),
		keys.Render(keyOf("config.format", "tab")) + theme.Subtle().Render(" sway / kanshi"),
		keys.Render(keyOf("config.close", "esc")) + theme.Subtle().Render(" close"),
	}
	lines = append(lines, "", " "+theme.Subtle().Render("paste into "+dest), " "+strings.Join(hint, theme.Subtle().Render("  ·  ")))
	if len(lines) > max(maxH-2, 1) {
		lines = lines[:max(maxH-2, 1)]
	}
	return draw.Frame(strings.Join(lines, "\n"), w, len(lines)+2, draw.FrameOpts{Title: "Config", Active: true})
}
