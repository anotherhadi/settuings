package pages

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/anotherhadi/foam/commands"
	"github.com/anotherhadi/foam/draw"
	"github.com/anotherhadi/foam/geom"
	"github.com/anotherhadi/foam/list"
	"github.com/anotherhadi/foam/theme"
)

type Item struct {
	Key    string
	Header bool
	Icon   string
	Label  string
	Note   string
	Dim    bool
	Data   any
}

type Frame struct {
	Title       string
	Right       string
	DetailTitle string
	Detail      []string
	Empty       string
}

type List struct {
	raw     []Item
	compact bool
	items   []Item
	l       list.Model
	rows    geom.Rect
	detail  bool
}

func (p *List) seek(i, dir int) int {
	for ; i >= 0 && i < len(p.items); i += dir {
		if !p.items[i].Header {
			return i
		}
	}
	return -1
}

func (p *List) Set(items []Item) {
	p.raw = items
	p.layout()
}

func (p *List) layout() {
	items := p.raw
	key := ""
	if it, ok := p.Selected(); ok {
		key = it.Key
	}
	prev := p.l.Selected()
	var spaced []Item
	for _, it := range items {
		if it.Header && !it.Dim && !p.compact {
			if len(spaced) > 0 {
				spaced = append(spaced, Item{Header: true})
			}
			spaced = append(spaced, it, Item{Header: true})
			continue
		}
		spaced = append(spaced, it)
	}
	items = spaced
	p.items = items
	p.l.SetLen(len(items))
	for i, it := range items {
		if !it.Header && key != "" && it.Key == key {
			p.l.Select(i)
			return
		}
	}
	i := min(prev, len(items)-1)
	if j := p.seek(i, 1); j >= 0 {
		p.l.Select(j)
	} else if j := p.seek(i, -1); j >= 0 {
		p.l.Select(j)
	}
}

func (p *List) Selected() (Item, bool) {
	i := p.l.Selected()
	if i < 0 || i >= len(p.items) || p.items[i].Header {
		return Item{}, false
	}
	return p.items[i], true
}

func (p *List) Len() int { return len(p.items) }

func (p *List) move(target, dir int) {
	target = min(max(target, 0), len(p.items)-1)
	if target < 0 || target >= len(p.items) {
		return
	}
	if p.items[target].Header {
		j := p.seek(target, dir)
		if j < 0 {
			j = p.seek(target, -dir)
		}
		if j < 0 {
			return
		}
		target = j
	}
	p.l.Select(target)
}

func (p *List) Move(d int) {
	dir := 1
	if d < 0 {
		dir = -1
	}
	p.move(p.l.Selected()+d, dir)
}

func (p *List) DetailHidden() bool { return !p.detail }

func (p *List) Commands() []commands.Command {
	page := func() int { return max(p.rows.H-1, 1) }
	run := func(f func()) func() tea.Cmd { return func() tea.Cmd { f(); return nil } }
	return []commands.Command{
		{ID: "list.down", Help: "move", Run: run(func() { p.Move(1) })},
		{ID: "list.up", Run: run(func() { p.Move(-1) })},
		{ID: "list.page-down", Run: run(func() { p.Move(page()) })},
		{ID: "list.page-up", Run: run(func() { p.Move(-page()) })},
		{ID: "list.top", Run: run(func() { p.move(0, 1) })},
		{ID: "list.bottom", Run: run(func() { p.move(len(p.items)-1, -1) })},
	}
}

func (p *List) Mouse(msg tea.MouseMsg) (activate bool) {
	switch msg := msg.(type) {
	case tea.MouseWheelMsg:
		switch msg.Button {
		case tea.MouseWheelUp:
			p.Move(-3)
		case tea.MouseWheelDown:
			p.Move(3)
		}
	case tea.MouseClickMsg:
		mm := msg.Mouse()
		if mm.Button != tea.MouseLeft || !p.rows.Contains(mm.X, mm.Y) {
			return false
		}
		start, end := p.l.Visible()
		i := start + mm.Y - p.rows.Y
		if i < start || i >= end || p.items[i].Header {
			return false
		}
		before := p.l.Selected()
		p.l.Select(i)
		return i == before
	}
	return false
}

func (p *List) hidden(start, end int) string {
	count := func(from, to int) (n int) {
		for _, it := range p.items[from:to] {
			if !it.Header {
				n++
			}
		}
		return n
	}
	var parts []string
	if n := count(0, start); n > 0 {
		parts = append(parts, "↑ "+itoa(n))
	}
	if n := count(end, len(p.items)); n > 0 {
		parts = append(parts, "↓ "+itoa(n))
	}
	return strings.Join(parts, "  ")
}

func (p *List) row(it Item, w int) string {
	if it.Header {
		st := theme.Accent().Style().Bold(true)
		if it.Dim {
			st = theme.Subtle().Style()
		}
		return " " + st.Render(ansi.Truncate(it.Label, max(w-2, 0), "…"))
	}
	lead := " "
	if it.Icon != "" {
		lead += it.Icon + " "
	}
	note := it.Note
	avail := w - ansi.StringWidth(lead) - 1
	if nw := ansi.StringWidth(note); nw > 0 {
		if avail-nw-2 >= 8 {
			avail -= nw + 2
		} else {
			note = ""
		}
	}
	label := ansi.Truncate(it.Label, max(avail, 0), "…")
	st := theme.Fg().Style()
	if it.Dim {
		st = theme.Subtle().Style()
	}
	gap := max(w-ansi.StringWidth(lead)-ansi.StringWidth(label)-ansi.StringWidth(note)-1, 0)
	return lead + st.Render(label) + strings.Repeat(" ", gap) + note + " "
}

func (p *List) Render(w, h int, f Frame) string {
	body := geom.Rect{W: w, H: h}
	listR, detR := body, geom.Rect{}
	switch {
	case len(f.Detail) == 0:
	case w >= 90:
		c := body.Columns(geom.Ratio(2), geom.Ratio(3))
		listR, detR = c[0], c[1]
	case h >= 22:
		r := body.Rows(geom.Ratio(3), geom.Ratio(2))
		listR, detR = r[0], r[1]
	}
	p.detail = !detR.Empty()

	iw, ih := max(listR.W-2, 0), max(listR.H-2, 0)
	p.rows = geom.Rect{X: listR.X + 1, Y: listR.Y + 1, W: iw, H: ih}
	if compact := ih < 16; compact != p.compact {
		p.compact = compact
		p.layout()
	}
	p.l.SetHeight(ih)
	footer := ""
	var content string
	if len(p.items) == 0 {
		content = draw.Hatch(iw, ih, f.Empty)
	} else {
		start, end := p.l.Visible()
		footer = p.hidden(start, end)
		lines := make([]string, 0, ih)
		for i := start; i < end; i++ {
			line := p.row(p.items[i], iw)
			if i == p.l.Selected() {
				line = draw.Selected(line, iw, true)
			}
			lines = append(lines, draw.Pad(line, iw))
		}
		content = draw.Fit(strings.Join(lines, "\n"), iw, ih)
	}
	out := draw.Frame(content, listR.W, listR.H, draw.FrameOpts{Title: f.Title, Right: f.Right, Footer: footer, Active: true, Scroll: p.l.Scroll()})
	if !p.detail {
		return out
	}
	dw, dh := max(detR.W-2, 0), max(detR.H-2, 0)
	var lines []string
	for _, l := range f.Detail {
		for _, w := range strings.Split(ansi.Wrap(l, max(dw-2, 1), ""), "\n") {
			lines = append(lines, " "+w)
		}
	}
	d := draw.Frame(draw.Fit(strings.Join(lines, "\n"), dw, dh), detR.W, detR.H, draw.FrameOpts{Title: f.DetailTitle})
	if detR.X > 0 {
		return lipgloss.JoinHorizontal(lipgloss.Top, out, d)
	}
	return lipgloss.JoinVertical(lipgloss.Left, out, d)
}

func kv(k, v string) string {
	return theme.Subtle().Render(draw.Pad(k, 14)) + theme.Fg().Render(v)
}

func blank() string { return "" }

func level(pct, w int) string {
	pct = min(max(pct, 0), 100)
	return draw.ProgressBar(pct, 100, w) + theme.Subtle().Render(" "+itoa(pct)+"%")
}
