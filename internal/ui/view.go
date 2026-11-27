package ui

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/anotherhadi/foam/commands"
	"github.com/anotherhadi/foam/draw"
	"github.com/anotherhadi/foam/geom"
	"github.com/anotherhadi/foam/minsize"
	"github.com/anotherhadi/foam/theme"
)

var minSize = minsize.Size{W: 32, H: 10}

func (m *Model) body() geom.Rect {
	top, bottom := 1, 0
	if theme.HelpBar() {
		bottom = 1
	}
	if theme.Spacious() {
		top++
		if bottom > 0 {
			bottom++
		}
	}
	return geom.Rect{X: 0, Y: top, W: m.w, H: max(m.h-top-bottom, 0)}
}

func (m *Model) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true
	v.WindowTitle = "settuings"
	if theme.Mouse() {
		v.MouseMode = tea.MouseModeCellMotion
	}
	if m.w == 0 || m.h == 0 {
		return v
	}
	if !minSize.Fits(m.w, m.h) {
		v.SetContent(minSize.View(m.w, m.h))
		return v
	}
	c := lipgloss.NewCanvas(m.w, m.h)
	b := m.body()
	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(m.topBar()),
		lipgloss.NewLayer(m.page().View(b.W, b.H)).X(b.X).Y(b.Y),
	}
	if theme.HelpBar() {
		layers = append(layers, lipgloss.NewLayer(m.helpBar()).Y(m.h-1))
	}
	c.Compose(lipgloss.NewCompositor(layers...))
	draw.FillRow(c, 0, theme.Surface())
	if theme.HelpBar() {
		draw.FillRow(c, m.h-1, theme.Surface())
	}
	m.overlays.Draw(c)
	if t := m.toasts(); len(t) > 0 {
		c.Compose(lipgloss.NewCompositor(t...))
	}
	v.SetContent(c.Render())
	return v
}

func (m *Model) brand() string {
	return theme.Accent().Style().Bold(true).Render(" "+theme.Icon("\U000f0493 ", "")+"settuings ") + " "
}

func (m *Model) tabs(brandW int) string {
	render := func(mode int) (string, []tabRect) {
		var b strings.Builder
		var rects []tabRect
		x := brandW
		for i, p := range m.pages {
			label := p.Label()
			switch {
			case mode == 1 && i != m.cur:
				label = itoa(i + 1)
			case mode == 2 && i != m.cur:
				continue
			}
			st := theme.Subtle().Style()
			if i == m.cur {
				st = lipgloss.NewStyle().Foreground(theme.Bg().Value()).Background(theme.Primary().Value()).Bold(true)
			}
			s := st.Render(" " + label + " ")
			rects = append(rects, tabRect{x, x + ansi.StringWidth(s), i})
			b.WriteString(s + " ")
			x += ansi.StringWidth(s) + 1
		}
		return b.String(), rects
	}
	var out string
	for mode := 0; mode < 3; mode++ {
		var rects []tabRect
		out, rects = render(mode)
		m.tabRects = rects
		if brandW+ansi.StringWidth(out) <= m.w {
			break
		}
	}
	return out
}

func (m *Model) topBar() string {
	brand := m.brand()
	bw := ansi.StringWidth(brand)
	if bw+ansi.StringWidth(m.pageLabelOnly()) > m.w {
		brand, bw = "", 0
	}
	line := brand + m.tabs(bw)
	return ansi.Truncate(line, m.w, "")
}

func (m *Model) pageLabelOnly() string { return " " + m.page().Label() + " " }

func (m *Model) helpBar() string {
	ks, ls := theme.Primary().Style(), theme.Subtle().Style()
	var parts []string
	scopes := m.path()
	if o := m.overlays.Top(); o != nil {
		scopes = []commands.Scope{o.Scope()}
	}
	for _, h := range m.reg.Hints(scopes) {
		first, _, _ := strings.Cut(h.Keys, " / ")
		parts = append(parts, ks.Render(first)+" "+ls.Render(h.Label))
	}
	sep := ls.Render("  ·  ")
	line := func() string { return " " + strings.Join(parts, sep) }
	for len(parts) > 1 && ansi.StringWidth(line()) > m.w {
		parts = append(parts[:len(parts)-2], parts[len(parts)-1])
	}
	return draw.Pad(ansi.Truncate(line(), m.w, "…"), m.w)
}

func (m *Model) toasts() []*lipgloss.Layer {
	var out []*lipgloss.Layer
	w := min(48, max(m.w/3, 24), m.w)
	y := 0
	m.toastRects = m.toastRects[:0]
	live := m.notes.Live()
	for i := len(live) - 1; i >= 0 && i >= len(live)-5; i-- {
		s := live[i].Toast(w)
		r := geom.Rect{X: max(m.w-w, 0), Y: y, W: w, H: lipgloss.Height(s)}
		if len(out) > 0 && y+r.H > m.h {
			break
		}
		m.toastRects = append(m.toastRects, toastRect{r, live[i].ID})
		out = append(out, lipgloss.NewLayer(s).X(r.X).Y(r.Y).Z(20))
		y += r.H
	}
	return out
}

func itoa(n int) string { return strconv.Itoa(n) }
